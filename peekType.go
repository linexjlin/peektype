package peektype

import (
	"strings"

	"github.com/linexjlin/simple-log"
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

type Peek struct {
	data     []byte
	Hostname string
	Type     int
}

func NewPeek() *Peek {
	var p Peek
	return &p
}

func (p *Peek) SetBuf(data []byte) {
	p.data = data
}

func (p *Peek) Parse() int {
	switch {
	case p.isHTTPS():
		log.Println("https")
		return HTTPS
	case p.isHTTP():
		log.Println("http")
		return HTTP
	case p.isSSH():
		log.Println("ssh")
		return SSH
	default:
		log.Println("UNKNOWN")
		log.Debug(string(p.data))
		return UNKNOWN
	}
}
func (p *Peek) isSSH() bool {
	return strings.Contains(SSHPATTAN, string(p.data[:4]))
}

func (p *Peek) isHTTP() bool {
	if strings.Contains(HTTPPATTAN, strings.TrimSpace(string(p.data[:4]))) {
		p.parseHTTPHostname()
		return true
	} else {
		return false
	}
}

func (p *Peek) isHTTPS() bool {
	if p.data[0] == 0x16 {
		p.parseHTTPSHostname()
		return true
	} else {
		return false
	}
}

func (p *Peek) parseHTTPHostname() {
	s := 0
	var line string
	for i, b := range p.data {
		if b == byte('\n') {
			line = strings.TrimSpace(string(p.data[s:i]))
			if strings.HasPrefix(line, "Host:") {
				hostname := strings.TrimSpace(strings.Split(line, ":")[1])
				p.Hostname = hostname
				break
			}
			s = i
		}
	}
	return
}

func (p *Peek) parseHTTPSHostname() {
	p.parseSNIHostname()
}

func (p *Peek) parseSNIHostname() {
	dataLen := len(p.data)
	if dataLen < 128 {
		return
	}
	// Simple SNI Protocol : SNI Handling Code from https://github.com/gpjt/stupid-proxy/
	//firstbyte
	current := 0
	if p.data[0] != 0x16 {
		log.Printf("Not TLS :-(")
		return
	}

	current++
	//version bytes
	if p.data[current] < 3 || (p.data[current] == 3 && p.data[current+1] < 1) {
		log.Printf("SSL < 3.1 so it's still not TLS v%d.%d", p.data[current], p.data[current+1])
		return
	}
	current += 2

	//resetLength
	restLength := (int(p.data[current]) << 8) + int(p.data[current+1])
	current += 2

	if current > dataLen {
		return
	}

	handshakeType := p.data[current]
	current += 1
	if handshakeType != 0x1 {
		log.Printf("Not a ClientHello")
		return
	}

	// Skip over another length
	current += 3
	// Skip over protocolversion
	current += 2
	// Skip over random number
	current += 4 + 28
	// Skip over session ID
	sessionIDLength := int(p.data[current])
	current += 1
	current += sessionIDLength

	if current > dataLen {
		return
	}
	cipherSuiteLength := (int(p.data[current]) << 8) + int(p.data[current+1])
	current += 2
	current += cipherSuiteLength

	if current > dataLen {
		return
	}
	compressionMethodLength := int(p.data[current])
	current += 1
	current += compressionMethodLength

	if current > dataLen {
		return
	}
	if current > restLength {
		log.Println("no extensions")
		return
	}

	// Skip over extensionsLength
	// extensionsLength := (int(rest[current]) << 8) + int(rest[current + 1])
	current += 2
	var hostname string
	for current+3 < restLength && hostname == "" {
		extensionType := (int(p.data[current]) << 8) + int(p.data[current+1])
		current += 2

		extensionDataLength := (int(p.data[current]) << 8) + int(p.data[current+1])
		current += 2

		if extensionType == 0 {
			// Skip over number of names as we're assuming there's just one
			current += 2

			nameType := p.data[current]
			current += 1
			if nameType != 0 {
				log.Printf("Not a hostname")
				return
			}
			nameLen := (int(p.data[current]) << 8) + int(p.data[current+1])
			current += 2
			hostname = string(p.data[current : current+nameLen])
		}

		current += extensionDataLength
	}

	if hostname == "" {
		log.Printf("No hostname")
		return
	} else {
		p.Hostname = hostname
	}
}
