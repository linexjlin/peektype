package peektype

import (
	"net"
	"strings"

	"errors"
	"fmt"
	"github.com/linexjlin/simple-log"
	"sync"
	"syscall"
)

const (
	HTTP = iota
	HTTPS
	SSH
	UNKNOWN
)

const (
	HTTPPATTAN = "GET HEAD POST POST PUT DELE TRAC OPTI CONN PATC"
	SSHPATTAN  = "SSH-"
)

var pool *sync.Pool

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 4096, 4096)
			return buf
		},
	}
}

func Peek(c net.Conn) (int, string, error) {
	var err error
	buf := pool.Get().([]byte)
	defer pool.Put(buf)
	//oobbuf := make([]byte, 512, 512)

	//	defer c.Close()

	conn := c.(*net.TCPConn)
	f, _ := conn.File()
	n, _, _, from, err := syscall.Recvmsg(int(f.Fd()), buf, nil, syscall.MSG_PEEK)
	if err != nil {
		fmt.Println(from, err)
		return UNKNOWN, "", errors.New("Unknown type")
	}
	buf = buf[:n]
	f.Close()

	hostname := ""
	switch {
	case isSSH(buf):
		log.Println("ssh")
		return SSH, "", nil
	case isHTTP(buf):
		log.Println("http")
		hostname = parseHTTPHostname(buf)
		if hostname == "" {
			return UNKNOWN, "", errors.New("Unknown type")
		}
		return HTTP, hostname, nil
	case isHTTPS(buf):
		log.Println("https")
		hostname = parseSNIHostname(buf)
		if hostname == "" {
			return UNKNOWN, "", errors.New("Unknown type")
		}
		return HTTPS, hostname, nil
	default:
		log.Println("UNKNOWN")
		return UNKNOWN, "", errors.New("Unknown type")
	}
	return UNKNOWN, "", errors.New("Unknown type")
}

func isSSH(data []byte) bool {
	return strings.Contains(SSHPATTAN, string(data[:4]))
}

func isHTTP(data []byte) bool {
	if strings.Contains(HTTPPATTAN, strings.TrimSpace(string(data[:4]))) {
		return true
	} else {
		return false
	}
}

func isHTTPS(data []byte) bool {
	if data[0] == 0x16 {
		return true
	} else {
		return false
	}
}

func parseHTTPHostname(data []byte) string {
	s := 0
	var line string
	for i, b := range data {
		if b == byte('\n') {
			line = strings.TrimSpace(string(data[s:i]))
			if strings.HasPrefix(line, "Host:") {
				hostname := strings.TrimSpace(strings.Split(line, ":")[1])
				return hostname
			}
			s = i
		}
	}
	return ""
}

func parseSNIHostname(data []byte) string {
	dataLen := len(data)
	if dataLen < 128 {
		return ""
	}
	// Simple SNI Protocol : SNI Handling Code from https://github.com/gpjt/stupid-proxy/
	//firstbyte
	current := 0
	if data[0] != 0x16 {
		log.Printf("Not TLS :-(")
		return ""
	}

	current++
	//version bytes
	if data[current] < 3 || (data[current] == 3 && data[current+1] < 1) {
		log.Printf("SSL < 3.1 so it's still not TLS v%d.%d", data[current], data[current+1])
		return ""
	}
	current += 2

	//resetLength
	restLength := (int(data[current]) << 8) + int(data[current+1])
	current += 2

	if current > dataLen {
		return ""
	}

	handshakeType := data[current]
	current += 1
	if handshakeType != 0x1 {
		log.Printf("Not a ClientHello")
		return ""
	}

	// Skip over another length
	current += 3
	// Skip over protocolversion
	current += 2
	// Skip over random number
	current += 4 + 28
	// Skip over session ID
	sessionIDLength := int(data[current])
	current += 1
	current += sessionIDLength

	if current > dataLen {
		return ""
	}
	cipherSuiteLength := (int(data[current]) << 8) + int(data[current+1])
	current += 2
	current += cipherSuiteLength

	if current > dataLen {
		return ""
	}
	compressionMethodLength := int(data[current])
	current += 1
	current += compressionMethodLength

	if current > dataLen {
		return ""
	}
	if current > restLength {
		log.Println("no extensions")
		return ""
	}

	// Skip over extensionsLength
	// extensionsLength := (int(rest[current]) << 8) + int(rest[current + 1])
	current += 2
	var hostname string
	for current+3 < restLength {
		if current+9 > len(data) {
			log.Printf("No hostname")
			return ""
		}
		extensionType := (int(data[current]) << 8) + int(data[current+1])
		current += 2

		extensionDataLength := (int(data[current]) << 8) + int(data[current+1])
		current += 2

		if extensionType == 0 {
			// Skip over number of names as we're assuming there's just one
			current += 2

			nameType := data[current]
			current += 1
			if nameType != 0 {
				log.Printf("Not a hostname")
				return ""
			}
			nameLen := (int(data[current]) << 8) + int(data[current+1])
			current += 2
			hostname = string(data[current : current+nameLen])

			return hostname
		}

		current += extensionDataLength
	}
	return ""

}
