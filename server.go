package main

import (
	"log"
	"net"
	"os"

	"gitub.com/cevdetakiah/grpc_chat/chatserver"
	"google.golang.org/grpc"
)

func main() {

	// assign port
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "5000" // default port to 5000 if PORT is not set in env
	}

	// init listener
	listen, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", Port, err)
	}

	log.Println("Listening @ : ", Port)

	// gRPC server instance
	grpcserver := grpc.NewServer()

	// register ChatService
	cs := chatserver.ChatServer{}
	chatserver.RegisterServicesServer(grpcserver, &cs)

	// grpc listen and serve
	err = grpcserver.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v\n", err)
	}
}
