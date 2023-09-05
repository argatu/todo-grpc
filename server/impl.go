package main

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"

	pb "github.com/argatu/todo-grpc/proto/todo/v1"
)

func (s *server) AddTask(_ context.Context, req *pb.AddTaskRequest) (*pb.AddTaskResponse, error) {
	if len(req.Description) == 0 {
		return nil, status.Error(codes.InvalidArgument, "description cannot be empty")
	}

	if req.DueDate.AsTime().Before(time.Now().UTC()) {
		return nil, status.Error(codes.InvalidArgument, "due date cannot be in the past")
	}

	id, err := s.d.addTask(req.Description, req.DueDate.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while adding task: %s", err.Error())
	}

	return &pb.AddTaskResponse{
		Id: id,
	}, nil
}

func (s *server) ListTasks(_ *pb.ListTasksRequest, stream pb.TodoService_ListTasksServer) error {
	ctx := stream.Context()

	return s.d.getTasks(func(t any) error {
		select {
		case <-ctx.Done():
			switch ctx.Err() {
			case context.Canceled:
				log.Printf("request cancelled: %s", ctx.Err())
			default:
			}
			return ctx.Err()
		case <-time.After(1 * time.Millisecond):
		}

		task := t.(*pb.Task)
		overdue := task.DueDate != nil && task.Done && task.DueDate.AsTime().Before(time.Now().UTC())

		return stream.Send(&pb.ListTasksResponse{
			Task:    task,
			Overdue: overdue,
		})
	})
}

func (s *server) UpdateTasks(stream pb.TodoService_UpdateTasksServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UpdateTaskResponse{})
		}

		if err != nil {
			return err
		}

		s.d.updateTask(
			req.Task.Id,
			req.Task.Description,
			req.Task.DueDate.AsTime(),
			req.Task.Done,
		)
	}
}

func (s *server) DeleteTasks(stream pb.TodoService_DeleteTasksServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		s.d.deleteTask(req.Id)

		stream.Send(&pb.DeleteTaskResponse{})
	}
}
