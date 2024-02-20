package main

import (
	"log"
	"net"
	"os"
)

func main() {

	listener, err := net.Listen("tcp", ":6161")

	if err != nil {
		log.Fatalf("Error tcp connection: %v", err)

		os.Exit(1)
	}

	log.Printf("Server is listening on %s", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s, %v", conn.RemoteAddr(), err)
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	message := []byte("Hello from server\n")

	n, err := conn.Write(message)

	if err != nil {
		log.Fatalf("Error from server: %v", err)

		return

	}

	if n != len(message) {
		log.Fatalf("Error for sending all message: %v", err)

		return
	}

}
