package main

import (
	"context"
	pb "github.com/argatu/todo-grpc/proto/todo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var fakeDb *FakeDb = NewFakeDb()

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	var testServer *server = &server{
		d: fakeDb,
	}

	pb.RegisterTodoServiceServer(s, testServer)
	go func() {
		if err := s.Serve(lis); err != nil && err.Error() != "closed" {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
