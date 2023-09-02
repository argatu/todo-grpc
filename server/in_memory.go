package main

import (
	"fmt"
	pb "github.com/argatu/todo-grpc/proto/todo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type inMemoryDb struct {
	tasks []*pb.Task
}

func New() db {
	return &inMemoryDb{}
}

func (d *inMemoryDb) addTask(description string, dueDate time.Time) (uint64, error) {
	id := uint64(len(d.tasks) + 1)
	d.tasks = append(d.tasks, &pb.Task{
		Id:          id,
		Description: description,
		DueDate:     timestamppb.New(dueDate),
	})
	return id, nil
}

func (d *inMemoryDb) getTasks(f func(any) error) error {
	for _, t := range d.tasks {
		if err := f(t); err != nil {
			return err
		}
	}
	return nil
}

func (d *inMemoryDb) updateTask(id uint64, description string, dueDate time.Time, done bool) error {
	for _, t := range d.tasks {
		if t.Id == id {
			t.Description = description
			t.DueDate = timestamppb.New(dueDate)
			t.Done = done
			return nil
		}
	}
	return fmt.Errorf("task with id %d not found", id)
}

func (d *inMemoryDb) deleteTask(id uint64) error {
	for i, t := range d.tasks {
		if t.Id == id {
			d.tasks = append(d.tasks[:i], d.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task with id %d not found", id)
}
