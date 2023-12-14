package server

import (
	"context"
	"crypto/sha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func LogInterceptor(s *Server) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		s.logger.Debug(info.FullMethod)
		return handler(ctx, req)
	}
}

func RequirePassInterceptor(s *Server) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// get metadata
		md, b := metadata.FromIncomingContext(ctx)
		if !b {
			return nil, status.Error(codes.Unauthenticated, "authenticate failed")
		}

		pass := md.Get("river.requirepass-bin")
		if len(pass) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authenticate failed")
		}

		sum := sha1.Sum([]byte(s.opt.Password))
		if string(sum[:]) != pass[0] {
			return nil, status.Error(codes.Unauthenticated, "authenticate failed")
		}
		return handler(ctx, req)
	}
}
