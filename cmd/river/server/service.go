package server

import (
	"context"
	riverdb "github.com/246859/river"
	"github.com/246859/river/cmd/river/riverpb"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

func (s *Server) Get(ctx context.Context, data *riverpb.RawData) (*riverpb.DataResult, error) {
	value, err := s.db.Get(data.GetData())
	if errors.Is(err, riverdb.ErrKeyNotFound) {
		return &riverpb.DataResult{
			Ok: false,
		}, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return &riverpb.DataResult{
			Ok: false,
		}, err
	}

	return &riverpb.DataResult{
		Ok:   true,
		Data: value,
	}, nil
}

func (s *Server) TTL(ctx context.Context, data *riverpb.RawData) (*riverpb.TTLResult, error) {
	ttl, err := s.db.TTL(data.GetData())
	if errors.Is(err, riverdb.ErrKeyNotFound) {
		return &riverpb.TTLResult{
			Ok: false,
		}, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return &riverpb.TTLResult{
			Ok: false,
		}, err
	}

	return &riverpb.TTLResult{
		Ok:  true,
		Ttl: ttl.Milliseconds(),
	}, nil
}

func (s *Server) Put(ctx context.Context, record *riverpb.Record) (*riverpb.InfoResult, error) {
	err := s.db.Put(record.Key, record.Value, time.Duration(record.Ttl))
	if err != nil {
		return &riverpb.InfoResult{
			Ok: false,
		}, err
	}

	return &riverpb.InfoResult{Ok: true}, nil
}

func (s *Server) Exp(ctx context.Context, record *riverpb.ExpRecord) (*riverpb.InfoResult, error) {
	err := s.db.Expire(record.Key, time.Duration(record.Ttl))
	if err != nil {
		return &riverpb.InfoResult{
			Ok: false,
		}, err
	}
	return &riverpb.InfoResult{Ok: true}, nil
}

func (s *Server) Del(ctx context.Context, data *riverpb.RawData) (*riverpb.InfoResult, error) {
	err := s.db.Del(data.GetData())
	if err != nil {
		return &riverpb.InfoResult{
			Ok: false,
		}, err
	}
	return &riverpb.InfoResult{Ok: true}, nil
}

func (s *Server) Stat(context.Context, *emptypb.Empty) (*riverpb.Status, error) {
	stats := s.db.Stats()
	return &riverpb.Status{
		Keys:     stats.KeyNums,
		Records:  stats.RecordNums,
		Datasize: stats.DataSize,
		Hintsize: stats.HintSize,
	}, nil
}
