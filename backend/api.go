package main

import (
	"net"
)

const ()

func StartAPI() (err error) {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			l.Err("Error getting connection:", err)
			continue
		}
		go Negotiate(conn)
	}
}

// Negotiate performs all negotation steps with the incoming
// connection and ensures that it is closed.
func Negotiate(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("hello iam nodeatlas\n"))
}
