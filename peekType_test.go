package peektype

import (
	"log"
	"net"
	"testing"
)

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
			go func() {
				var buf = make([]byte, 512)
				if _, e := conn.Read(buf); err != nil {
					log.Println(e)
				} else {
					peek := NewPeek()
					peek.SetBuf(&buf)
					t := peek.Parse()
					log.Println(t, peek.Hostname)
					//log.Println(string(buf))
				}
			}()
		}
	}
}
