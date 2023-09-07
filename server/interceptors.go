package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

const (
	authTokenKey   string = "auth_token"
	authTokenValue string = "authd"
)

func validateAuthToken(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	if t, ok := md[authTokenKey]; ok {
		switch {
		case len(t) != 1:
			return status.Error(codes.InvalidArgument, "auth_token should contain only one value")
		case t[0] != authTokenValue:
			return status.Error(codes.Unauthenticated, "incorrect auth_token")
		}
	} else {
		return status.Error(codes.Unauthenticated, "auth_token is missing")
	}
	return nil
}

func unaryAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := validateAuthToken(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func streamAuthInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := validateAuthToken(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

func unaryLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println(info.FullMethod, "called")
	return handler(ctx, req)
}

func streamLogInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println(info.FullMethod, "called")
	return handler(srv, ss)
}
