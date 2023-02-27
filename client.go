package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"gitub.com/cevdetakiah/grpc_chat/chatserver"
	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Enter Server IP::Port :::")
	reader := bufio.NewReader(os.Stdin)
	serverID, err := reader.ReadString('\n')

	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}

	serverID = strings.Trim(serverID, "\r\n")

	log.Println("Connecting : " + serverID)

	// connect to grpc server
	conn, err := grpc.Dial(serverID, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server :: %v", err)
	}
	defer conn.Close()

	// call ChatService to create a stream
	client := chatserver.NewServicesClient((conn))

	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}

	// implement communication to gRPC server
	ch := clienthandle{stream: stream}
	ch.clientConfig()
	// two go routines for sending and receiving messages
	go ch.sendMessage()
	go ch.receiveMessage()

	// blocker
	bl := make(chan bool)
	<-bl

}

// client handle
type clienthandle struct {
	stream     chatserver.Services_ChatServiceClient
	clientName string
}

func (ch *clienthandle) clientConfig() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Your Name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
}

// send message
func (ch *clienthandle) sendMessage() {

	// loop scanning for new messages
	for {

		reader := bufio.NewReader(os.Stdin)

		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		clientMessageBox := &chatserver.FromClient{
			Name: ch.clientName,
			Body: clientMessage,
		}

		// read message from console and use Send method of stream to push method
		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}
	}
}

// receive message
func (ch *clienthandle) receiveMessage() {

	// loop scanning for new messages
	for {
		msg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message")
		}

		// print message to console
		fmt.Printf("%s : %s \n", msg.Name, msg.Body)
	}
}
