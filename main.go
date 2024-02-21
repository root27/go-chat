package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"unicode/utf8"
)

type MessageType int

const (
	Connected MessageType = iota + 1
	Disconnected
	NewMessage
)

type Client struct {
	conn         net.Conn
	last_message time.Time
	strokes      int
}

type Message struct {
	Text string
	From net.Conn
	Type MessageType
}

const (
	timeLimit = 10 * 60.0

	strokeLimit = 10

	Rate = 1.0
)

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

		text := string(buf[:n])

		if string(buf[:n]) == "exit\r\n" {

			conn.Close()

			msg := Message{
				From: conn,
				Type: Disconnected,
				Text: "",
			}

			messages <- msg

			return
		}

		msg := Message{
			Text: text,
			From: conn,
			Type: NewMessage,
		}

		messages <- msg

	}

}

func server(messages chan Message) {

	clients := map[string]*Client{}

	bans_list := map[string]time.Time{}

	for {
		msg := <-messages

		switch msg.Type {

		case Connected:

			addr := msg.From.RemoteAddr().(*net.TCPAddr)

			bannedAt, banned := bans_list[addr.IP.String()]

			now := time.Now()

			if banned {

				if now.Sub(bannedAt).Seconds() >= timeLimit {

					delete(bans_list, addr.IP.String())

					banned = false

				}

			}

			if !banned {

				log.Printf("New client connected: %s", msg.From.RemoteAddr())

				clients[msg.From.RemoteAddr().String()] = &Client{

					conn: msg.From,

					last_message: time.Now(),
				}
			} else {

				msg.From.Write([]byte(fmt.Sprintf("You are banned: %f seconds left\n", timeLimit-now.Sub(bannedAt).Seconds())))

				msg.From.Close()
			}

		case Disconnected:

			log.Printf("Client disconnected: %s", msg.From.RemoteAddr())

			delete(clients, msg.From.RemoteAddr().String())

		case NewMessage:

			author_addr := msg.From.RemoteAddr().(*net.TCPAddr)

			author := clients[author_addr.String()]

			now := time.Now()

			if now.Sub(author.last_message).Seconds() >= Rate {

				if utf8.ValidString(msg.Text) {

					author.last_message = now

					author.strokes = 0

					log.Printf("Client %s sent a message: %s", msg.From.RemoteAddr(), msg.Text)

					for _, client := range clients {

						if client.conn.RemoteAddr().String() != msg.From.RemoteAddr().String() {

							client.conn.Write([]byte(msg.Text))

						}

					}

				} else {

					author.strokes += 1

					if author.strokes > strokeLimit {

						bans_list[author_addr.IP.String()] = now

						author.conn.Write([]byte("You are banned !!!"))

						author.conn.Close()

					}

				}

			} else {

				author.strokes += 1

				if author.strokes > strokeLimit {

					bans_list[author_addr.IP.String()] = now

					author.conn.Write([]byte("You are banned !!!"))

					author.conn.Close()

				}

			}

		}
	}

}
