package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
)

type IncomingMessage struct {
	From    string
	Message string
}

type OutgoingMessage struct {
	To      string
	Message string
}

type Server struct {
	clients  map[string]*CustomClient
	listener net.Listener
}

func (server Server) PrintClients() {
	fmt.Println("Clients: ", len(server.clients))
}

type CustomClient struct {
	client net.Conn
	server *Server
}

func (client CustomClient) RemoteAddress() string {
	return client.client.RemoteAddr().String()
}

const (
	CONN_ADDR      = "localhost"
	CONN_PORT      = "9090"
	CONN_INTERFACE = CONN_ADDR + ":" + CONN_PORT
	CONN_TYPE      = "tcp"
)

func (server Server) propagate(sender CustomClient, msg string) {
	enc_msg := base64.RawStdEncoding.EncodeToString([]byte(msg))
	sender_info := sender.RemoteAddress()

	message := STX + PROTOCOL_VERSION + MSG_SEPARATOR + sender_info + MSG_SEPARATOR + enc_msg + ETX

	for i := range server.clients {
		client := server.clients[i]
		if *client == sender {
			continue
		} else {
			client.client.Write([]byte(message))
		}
	}
}

func handle_client(client CustomClient) {
	buff, err := bufio.NewReader(client.client).ReadBytes('\n')

	if err != nil {
		fmt.Println("Client " + client.RemoteAddress() + " has left.")
		delete(client.server.clients, client.RemoteAddress())
		client.server.PrintClients()
		return
	}

	message := string(buff[:len(buff)-1])

	log.Println("[Client "+client.RemoteAddress()+"]: ", message)

	client.server.propagate(client, message)

	handle_client(client)
}

func loop(server Server) {
	for {
		client, err := server.listener.Accept()
		if err != nil {
			fmt.Println("Client error: " + err.Error())
			continue
		}

		fmt.Println("Client at " + client.RemoteAddr().String() + " connected.")
		custom_client := CustomClient{client: client, server: &server}
		// server.clients = append(server.clients, &custom_client)
		server.clients[custom_client.RemoteAddress()] = &custom_client

		server.PrintClients()

		go handle_client(custom_client)
	}
}

func main() {
	fmt.Println("Starting server on interface " + CONN_INTERFACE)
	l, err := net.Listen(CONN_TYPE, CONN_INTERFACE)
	if err != nil {
		fmt.Println("Error listening on interface " + CONN_INTERFACE + "\nError: " + err.Error())
		os.Exit(1)
	}

	server := Server{listener: l, clients: make(map[string]*CustomClient)}

	defer server.listener.Close()

	loop(server)
}
