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
	if err := req.Validate(); err != nil {
		return nil, err
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
			case context.DeadlineExceeded:
				log.Printf("deadline exceeded: %s", ctx.Err())
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
			return stream.SendAndClose(&pb.UpdateTasksResponse{})
		}

		if err != nil {
			return err
		}

		if err := s.d.updateTask(
			req.Id,
			req.Description,
			req.DueDate.AsTime(),
			req.Done,
		); err != nil {
			return status.Errorf(
				codes.Internal,
				"error while updating task: %s",
				err.Error(),
			)
		}
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

		if err := s.d.deleteTask(req.Id); err != nil {
			return status.Errorf(
				codes.Internal,
				"error while deleting task: %s",
				err.Error(),
			)
		}

		stream.Send(&pb.DeleteTasksResponse{})
	}
}
