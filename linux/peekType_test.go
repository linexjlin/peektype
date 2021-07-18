package peektype

import (
	"net"
	"testing"

	"github.com/linexjlin/simple-log"
)

//refer https://golang.hotexamples.com/zh/examples/syscall/-/Recvmsg/golang-recvmsg-function-examples.html
/*
func handleConnection(c net.Conn) {
	var err error
	buf := make([]byte, 4096, 4096)
	//oobbuf := make([]byte, 512, 512)

	defer c.Close()

	conn := c.(*net.TCPConn)
	f, _ := conn.File()
	n, _, _, from, err := syscall.Recvmsg(int(f.Fd()), buf, nil, syscall.MSG_PEEK)
	if err != nil {
		fmt.Println(from, err)
		return

	}
	f.Close()

	log.Debug(n)
	log.Debugf("%x", buf)
}

func connCheck(conn net.Conn) error {
	var sysErr error = nil
	rc, err := conn.(syscall.Conn).SyscallConn()
	if err != nil {
		return err
	}
	err = rc.Read(func(fd uintptr) bool {
		//	var buf []byte = []byte{128}
		buf := make([]byte, 4096, 4096)
		//	n, _, err := syscall.Recvfrom(int(fd), buf, syscall.MSG_PEEK|syscall.MSG_DONTWAIT)
		n, _, _, _, err := syscall.Recvmsg(int(fd), buf, nil, 0)
		switch {
		case n == 0 && err == nil:
			sysErr = io.EOF
		case err == syscall.EAGAIN || err == syscall.EWOULDBLOCK:
			sysErr = nil
		default:
			sysErr = err

		}
		log.Debugf("%s\n", string(buf))
		return true

	})
	if err != nil {
		return err
	}

	return sysErr

} */

func TestListenServe(t *testing.T) {
	listenAddr := "0.0.0.0:58080"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	} else {
		log.Println("Listen on", listenAddr)
	}

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Println(err)
		} else {
			log.Println("accept connectino from:", conn.RemoteAddr())
			log.Debug(Peek(conn))
			//connCheck(conn)
			/*
				go func() {
					var buf = make([]byte, 512)
					if n, e := conn.Read(buf); err != nil {
						log.Println(e)
					} else {
						log.Println("read:", n)
						peek := NewPeek()
						buf = buf[:n]
						peek.SetBuf(buf)
						log.Println(string(buf))
						log.Println(buf)
						t := peek.Parse()
						log.Println(t, peek.Hostname)
						//log.Println(string(buf))
					}
				}() */

		}
	}
}
