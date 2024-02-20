package main

import (
	"log"
	"net"
)

type MessageType int

const (
	Connected MessageType = iota + 1
	Disconnected
	NewMessage
)

type Client struct {
	conn net.Conn
}

type Message struct {
	Text string
	From net.Conn
	Type MessageType
}

func main() {

	listener, err := net.Listen("tcp", ":6161")

	if err != nil {
		log.Fatalf("Error tcp connection: %v", err)
	}

	log.Printf("Server is listening on %s", listener.Addr())

	messages := make(chan Message)

	go server(messages)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %s, %v", conn.RemoteAddr(), err)
			continue
		}

		messages <- Message{
			From: conn,
			Type: Connected,
			Text: "",
		}

		go client(conn, messages)
	}

}

func client(conn net.Conn, messages chan Message) {

	buf := make([]byte, 1024)

	for {

		n, err := conn.Read(buf)

		if err != nil {

			conn.Close()

			log.Printf("Error reading from connection: %v", err)

			msg := Message{
				From: conn,
				Type: Disconnected,
				Text: "",
			}

			messages <- msg

			return
		}

		if string(buf[0:n]) == "exit" {

			conn.Close()

			messages <- Message{
				From: conn,
				Type: Disconnected,
				Text: "",
			}

			return

		}

		msg := Message{
			Text: string(buf[0:n]),
			From: conn,
			Type: NewMessage,
		}

		messages <- msg
	}

}

func server(messages chan Message) {

	clients := map[string]*Client{}

	for {
		msg := <-messages

		switch msg.Type {

		case Connected:

			clients[msg.From.RemoteAddr().String()] = &Client{
				conn: msg.From,
			}

			log.Printf("Client connected: %s", msg.From.RemoteAddr())

		case Disconnected:

			log.Printf("Client disconnected: %s", msg.From.RemoteAddr())

			delete(clients, msg.From.RemoteAddr().String())

		case NewMessage:

			log.Printf("Message from %s: %s", msg.From.RemoteAddr(), msg.Text)

			for _, client := range clients {

				if client.conn.RemoteAddr().String() != msg.From.RemoteAddr().String() {
					_, err := client.conn.Write([]byte(msg.Text))
					if err != nil {
						log.Printf("Error writing to connection: %v", err)

						client.conn.Close()

					}

				}

			}

		}

	}
}
