package main

import (
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"

	pb "github.com/argatu/todo-grpc/proto/todo/v1"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatalln("usage: server [IP_ADDR]")
	}

	addr := args[0]
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	defer func(lis net.Listener) {
		if err := lis.Close(); err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
	}(lis)

	creds, err := credentials.NewServerTLSFromFile("./certs/server_cert.pem", "./certs/server_key.pem")
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryAuthInterceptor),
		grpc.StreamInterceptor(streamAuthInterceptor),
	}
	s := grpc.NewServer(opts...)

	pb.RegisterTodoServiceServer(s, &server{
		d: New(),
	})

	log.Printf("listening at %s\n", addr)

	defer s.Stop()
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
