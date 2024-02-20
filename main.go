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
	From Client
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
			From: Client{conn},
			Type: Connected,
			Text: "",
		}

		go client(conn, messages)
	}

}

func client(conn net.Conn, messages chan Message) {

	client := Client{conn}

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {

			client.conn.Close()

			log.Printf("Error reading from connection: %v", err)

			msg := Message{
				From: client,
				Type: Disconnected,
				Text: "",
			}

			messages <- msg

			return
		}

		if string(buf[0:n]) == "exit" {

			client.conn.Close()

			messages <- Message{
				From: client,
				Type: Disconnected,
				Text: "",
			}

			return

		}

		msg := Message{
			Text: string(buf[0:n]),
			From: client,
			Type: NewMessage,
		}

		messages <- msg
	}

}

func server(messages chan Message) {

	clients := make(map[string]*Client)

	for {
		msg := <-messages

		switch msg.Type {

		case Connected:

			clients[msg.From.conn.RemoteAddr().String()] = &msg.From

			log.Printf("Client connected: %s", msg.From.conn.RemoteAddr())

		case Disconnected:

			log.Printf("Client disconnected: %s", msg.From.conn.RemoteAddr())

			delete(clients, msg.From.conn.RemoteAddr().String())

		case NewMessage:

			log.Printf("Message from %s: %s", msg.From.conn.RemoteAddr(), msg.Text)

			for _, client := range clients {

				if client.conn.RemoteAddr().String() != msg.From.conn.RemoteAddr().String() {
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
