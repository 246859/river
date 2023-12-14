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
	if data == nil {
		return &riverpb.DataResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

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

	s.logger.Debug("Get", "ksz", len(data.Data), "vsz", len(value))

	return &riverpb.DataResult{
		Ok:   true,
		Data: value,
	}, nil
}

func (s *Server) TTL(ctx context.Context, data *riverpb.RawData) (*riverpb.TTLResult, error) {
	if data == nil {
		return &riverpb.TTLResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

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

	s.logger.Debug("TTL", "ksz", len(data.Data), "ttl", ttl)

	return &riverpb.TTLResult{
		Ok:  true,
		Ttl: ttl.Milliseconds(),
	}, nil
}

func (s *Server) Put(ctx context.Context, record *riverpb.Record) (*riverpb.InfoResult, error) {
	if record == nil {
		return &riverpb.InfoResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

	err := s.db.Put(record.Key, record.Value, time.Duration(record.Ttl))
	if err != nil {
		return &riverpb.InfoResult{
			Ok: false,
		}, err
	}

	s.logger.Debug("Put", "ksz", len(record.Key), "vsz", len(record.Value), "ttl", time.Duration(record.Ttl))

	return &riverpb.InfoResult{Ok: true}, nil
}

func (s *Server) Exp(ctx context.Context, record *riverpb.ExpRecord) (*riverpb.InfoResult, error) {
	if record == nil {
		return &riverpb.InfoResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

	err := s.db.Expire(record.Key, time.Duration(record.Ttl))
	if err != nil {
		return &riverpb.InfoResult{
			Ok: false,
		}, err
	}

	s.logger.Debug("Exp", "ksz", len(record.Key), "ttl", time.Duration(record.Ttl))

	return &riverpb.InfoResult{Ok: true}, nil
}

func (s *Server) Del(ctx context.Context, data *riverpb.RawData) (*riverpb.InfoResult, error) {
	if data == nil {
		return &riverpb.InfoResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

	err := s.db.Del(data.GetData())
	if err != nil {
		return &riverpb.InfoResult{
			Ok: false,
		}, err
	}
	s.logger.Debug("Del", "ksz", len(data.Data))
	return &riverpb.InfoResult{Ok: true}, nil
}

func (s *Server) PutInBatch(ctx context.Context, opt *riverpb.BatchPutOption) (*riverpb.BatchResult, error) {
	if opt == nil {
		return &riverpb.BatchResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

	batch, err := s.db.Batch(riverdb.BatchOption{
		Size:        opt.BatchSize,
		SyncOnFlush: true,
	})

	if err != nil {
		return &riverpb.BatchResult{Ok: false}, status.Errorf(codes.InvalidArgument, "batch options invalid")
	}

	var records []riverdb.Record
	for _, record := range opt.Records {
		records = append(records, riverdb.Record{
			K:   record.Key,
			V:   record.Value,
			TTL: time.Duration(record.Ttl),
		})
	}

	if err := batch.WriteAll(records); err != nil {
		return &riverpb.BatchResult{Ok: false}, err
	}

	if err := batch.Flush(); err != nil {
		return &riverpb.BatchResult{Ok: false}, err
	}

	s.logger.Debug("PutInBatch", "records", len(records), "batch", opt.BatchSize, "effected", batch.Effected())

	return &riverpb.BatchResult{
		Ok:       true,
		Effected: batch.Effected(),
	}, nil
}

func (s *Server) DelInBatch(ctx context.Context, opt *riverpb.BatchDelOption) (*riverpb.BatchResult, error) {
	if opt == nil {
		return &riverpb.BatchResult{Ok: false}, status.Error(codes.InvalidArgument, "args is nil")
	}

	batch, err := s.db.Batch(riverdb.BatchOption{
		Size:        opt.BatchSize,
		SyncOnFlush: true,
	})

	if err != nil {
		return &riverpb.BatchResult{Ok: false}, status.Errorf(codes.InvalidArgument, "batch options invalid")
	}

	if err := batch.DeleteAll(opt.Keys); err != nil {
		return &riverpb.BatchResult{Ok: false}, err
	}

	if err := batch.Flush(); err != nil {
		return &riverpb.BatchResult{Ok: false}, err
	}

	s.logger.Debug("DelInBatch", "records", len(opt.Keys), "batch", opt.BatchSize, "effected", batch.Effected())

	return &riverpb.BatchResult{
		Ok:       true,
		Effected: batch.Effected(),
	}, nil
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

func (s *Server) Range(ctx context.Context, opt *riverpb.RangeOption) (*riverpb.RangeResult, error) {
	if opt == nil {
		return &riverpb.RangeResult{}, status.Error(codes.InvalidArgument, "args is nil")
	}

	ropt := riverdb.RangeOptions{
		Min:     opt.MinKey,
		Max:     opt.MaxKey,
		Pattern: opt.Pattern,
		Descend: opt.Descend,
	}

	var ks []riverdb.Key

	err := s.db.Range(ropt, func(key riverdb.Key) bool {
		ks = append(ks, key)
		return true
	})

	if err != nil {
		return &riverpb.RangeResult{}, status.Error(codes.InvalidArgument, "invalid range options")
	}

	s.logger.Debug("Range", "range_size", len(ks))

	return &riverpb.RangeResult{
		Keys:  ks,
		Count: int64(len(ks)),
	}, nil
}
