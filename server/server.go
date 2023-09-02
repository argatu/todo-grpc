package main

import (
	pb "github.com/argatu/todo-grpc/proto/todo/v1"
)

type server struct {
	d db
	pb.UnimplementedTodoServiceServer
}
