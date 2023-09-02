package main

import (
	"context"
	"io"
	"time"

	pb "github.com/argatu/todo-grpc/proto/todo/v1"
)

func (s *server) AddTask(_ context.Context, req *pb.AddTaskRequest) (*pb.AddTaskResponse, error) {
	id, err := s.d.addTask(req.Description, req.DueDate.AsTime())
	if err != nil {
		return nil, err
	}

	return &pb.AddTaskResponse{
		Id: id,
	}, nil
}

func (s *server) ListTasks(_ *pb.ListTasksRequest, stream pb.TodoService_ListTasksServer) error {
	return s.d.getTasks(func(t any) error {
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