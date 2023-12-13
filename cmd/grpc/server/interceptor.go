package server

import (
	"context"
	"crypto/sha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
		requirepass := ctx.Value("river.requirepass")
		sum := sha1.Sum([]byte(s.opt.Password))
		if sum != requirepass {
			return nil, status.Error(codes.Unauthenticated, "authenticate failed")
		}
		return handler(ctx, req)
	}
}
