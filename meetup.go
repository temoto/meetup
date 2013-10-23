// Meetup is a micro proxy program that enables to connect two endpoints via TCP sockets.
// Either of endpoint may listen or connect.
//
// Example usage:
// - you have a faulty PHP application and you want to debug it with xdebug
// - xdebug can connect to your machine:9000, but you are behind NAT
// - so you run `meetup -listen1=:9000 -listen2=:9001` on the application server
// - and another `meetup -connect1=appserver:9001 -connect=localhost:9000` on your machine
// First instance listens two ports and when a connection arrives on both, it creates
// a bidirectional buffered pipe between the two. The other instance connects to
// first meetup on appserver:9001 and also to your local IDE :9000 and likewise
// pipes data in both directions.
//
// Disconnect on one end of the pipe breaks the connection to the other end.
// -connect mode forever attempts to reconnect with 5 second interval between attempts.

package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var timeZero time.Time

func connect(addr string, out chan net.Conn) {
	for {
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			log.Printf("connect: %s %s", addr, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		out <- conn
	}
}

func listen(addr string, out chan net.Conn) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("listen: accept: %s %s", addr, err.Error())
			continue
		}
		out <- conn
	}
}

func copyTimeout(dst, src net.Conn, readTimeout, writeTimeout time.Duration) (written int64, err error) {
	buf := make([]byte, 32<<10)
	for {
		src.SetDeadline(time.Now().Add(readTimeout))
		nr, er := src.Read(buf)
		src.SetDeadline(timeZero)
		if nr > 0 {
			dst.SetDeadline(time.Now().Add(writeTimeout))
			nw, ew := dst.Write(buf[0:nr])
			dst.SetDeadline(timeZero)
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

func main() {
	var (
		addrListen1, addrListen2   string
		addrConnect1, addrConnect2 string
	)
	flag.StringVar(&addrListen1, "listen1", "", "endpoint 1 listen")
	flag.StringVar(&addrConnect1, "connect1", "", "endpoint 1 connect")
	flag.StringVar(&addrListen2, "listen2", "", "endpoint 2 listen")
	flag.StringVar(&addrConnect2, "connect2", "", "endpoint 2 connect")
	flag.Parse()

	conns1 := make(chan net.Conn)
	conns2 := make(chan net.Conn)
	if addrListen1 != "" {
		go listen(addrListen1, conns1)
	} else if addrConnect1 != "" {
		go connect(addrConnect1, conns1)
	} else {
		log.Printf("Either -listen1 or -connect1 must be specified")
		os.Exit(1)
	}
	if addrListen2 != "" {
		go listen(addrListen2, conns2)
	} else if addrConnect2 != "" {
		go connect(addrConnect2, conns2)
	} else {
		log.Printf("Either -listen2 or -connect2 must be specified")
		os.Exit(1)
	}

	readTimeout := 60 * time.Second
	writeTimeout := 60 * time.Second

	for {
		c1 := <-conns1
		c2 := <-conns2
		go copyTimeout(c2, c1, readTimeout, writeTimeout)
		go copyTimeout(c1, c2, readTimeout, writeTimeout)
	}
}
