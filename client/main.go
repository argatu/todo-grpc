package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/argatu/todo-grpc/proto/todo/v1"
)

func addTask(c pb.TodoServiceClient, description string, dueDate time.Time) uint64 {
	req := &pb.AddTaskRequest{
		Description: description,
		DueDate:     timestamppb.New(dueDate),
	}

	res, err := c.AddTask(context.Background(), req)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.InvalidArgument, codes.Internal:
				log.Fatalf("%s: %s", s.Code().String(), s.Message())
			default:
				log.Fatalf("unexpected error: %v", err)
			}
		} else {
			panic(err)
		}
	}

	fmt.Printf("added task: %d\n", res.Id)
	return res.Id
}

func printTasks(c pb.TodoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &pb.ListTasksRequest{}
	stream, err := c.ListTasks(ctx, req)
	if err != nil {
		log.Fatalf("error while calling ListTasks RPC: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}

		fmt.Println(res.Task.String(), "overdue: ", res.Overdue)
	}
}

func updateTask(c pb.TodoServiceClient, reqs ...*pb.UpdateTaskRequest) {
	stream, err := c.UpdateTasks(context.Background())
	if err != nil {
		log.Fatalf("error while calling UpdateTask RPC: %v", err)
	}

	for _, r := range reqs {
		if err := stream.Send(r); err != nil {
			log.Fatalf("error while sending request to server: %v", err)
		}

		if r.Task != nil {
			fmt.Println("updated task with id: ", r.Task.Id)
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		log.Fatalf("error while receiving response from server: %v", err)
	}
}

func deleteTasks(c pb.TodoServiceClient, reqs ...*pb.DeleteTaskRequest) {
	stream, err := c.DeleteTasks(context.Background())
	if err != nil {
		log.Fatalf("error while calling DeleteTasks RPC: %v", err)
	}

	waitc := make(chan struct{})
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}

			if err != nil {
				log.Fatalf("error while reading stream: %v", err)
			}

			log.Println("deleted task")
		}
	}()

	for _, req := range reqs {
		if err := stream.Send(req); err != nil {
			return
		}
		if err := stream.CloseSend(); err != nil {
			return
		}
		<-waitc
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatalln("usage: client [IP_ADDR]")
	}

	addr := args[0]

	creds, err := credentials.NewClientTLSFromFile("./certs/server_cert.pem", "x.test.example.com")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(unaryAuthInterceptor),
		grpc.WithStreamInterceptor(streamAuthInterceptor),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{""round_robin:{}}]}`),
	}
	conn, err := grpc.Dial(addr, opts...)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
	}(conn)

	c := pb.NewTodoServiceClient(conn)

	fmt.Println("--------ADD--------")
	dueDate := time.Now().Add(5 * time.Second)
	id1 := addTask(c, "This is a task", dueDate)
	id2 := addTask(c, "This is another task", dueDate)
	fmt.Println("--------------------")

	fmt.Println("--------LIST--------")
	printTasks(c)
	fmt.Println("--------------------")

	upReq := []*pb.UpdateTaskRequest{
		{
			Task: &pb.Task{
				Id:          id1,
				Description: "This is an updated task",
				Done:        true,
			},
		},
		{
			Task: &pb.Task{
				Id:          id2,
				Description: "This is another updated task",
				DueDate:     timestamppb.New(dueDate.Add(5 * time.Hour)),
			},
		},
	}

	fmt.Println("--------UPDATE--------")
	updateTask(c, upReq...)
	printTasks(c)
	fmt.Println("--------------------")

	delReq := []*pb.DeleteTaskRequest{
		{Id: id2},
	}
	fmt.Println("--------DELETE--------")
	deleteTasks(c, delReq...)
	printTasks(c)
	fmt.Println("--------------------")

	fmt.Println("--------ERROR--------")
	addTask(c, "not empty", time.Now().Add(-5*time.Second))
	fmt.Println("--------------------")
}
