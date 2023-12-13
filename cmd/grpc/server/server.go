package server

import (
	"context"
	"fmt"
	riverdb "github.com/246859/river"
	"github.com/246859/river/cmd/grpc/riverpb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
)

var (
	ErrClosed = errors.New("server is closed")
)

// NewServer return a new grpc server for river db
func NewServer(ctx context.Context, cfgpath string) (*Server, error) {
	var option Options

	if len(cfgpath) > 0 {
		readOpt, err := readOption(cfgpath)
		if err != nil {
			return nil, err
		}
		option = readOpt
	}

	if len(option.Address) == 0 {
		option.Address = ":6868"
	}

	if option.MaxSize == 0 {
		option.MaxSize = riverdb.DefaultOptions.MaxSize
	}

	if option.TxnLevel == 0 {
		option.TxnLevel = uint8(riverdb.DefaultOptions.Level)
	}

	if len(option.LogFile) == 0 {
		option.LogFile = DefaultLogFile
	}

	if len(option.LogLevel) == 0 {
		option.LogLevel = slog.LevelInfo.String()
	}

	dbopt := riverdb.Options{
		Dir:             DefaultDir,
		MaxSize:         option.MaxSize,
		BlockCache:      option.BlockCache,
		Fsync:           option.Fsync,
		FsyncThreshold:  option.FsyncThreshold,
		MergeCheckpoint: option.Checkpoint,

		WatchSize:   riverdb.DefaultOptions.WatchSize,
		WatchEvents: riverdb.DefaultOptions.WatchEvents,
		Compare:     riverdb.DefaultOptions.Compare,
		Level:       riverdb.TxnLevel(option.TxnLevel),
		ClosedGc:    riverdb.DefaultOptions.ClosedGc,
	}

	return &Server{
		ctx:   ctx,
		opt:   option,
		dbopt: dbopt,
		db:    nil,
	}, nil
}

type Server struct {
	riverpb.UnimplementedRiverServer

	opt   Options
	dbopt riverdb.Options

	ctx context.Context

	db     *riverdb.DB
	server *grpc.Server
	logger *Logger

	once sync.Once

	closed atomic.Bool
}

func (s *Server) init() error {
	var initErr error
	// make sure init once
	s.once.Do(func() {
		// initialize db
		db, err := riverdb.OpenWithCtx(s.ctx, s.dbopt)
		if err != nil {
			initErr = err
			return
		}
		s.db = db

		logger, err := newLogger(s.opt.LogFile, s.opt.LogLevel)
		if err != nil {
			initErr = err
			return
		}
		s.logger = logger

		var grpcopts []grpc.ServerOption

		// interceptor
		grpcopts = append(grpcopts, grpc.ChainUnaryInterceptor(LogInterceptor(s)))
		if len(s.opt.Password) >= 0 {
			grpcopts = append(grpcopts, grpc.ChainUnaryInterceptor(RequirePassInterceptor(s)))
		}

		// tsl support
		if len(s.opt.TlsKey) > 0 && len(s.opt.TlsPem) > 0 {
			creds, err := credentials.NewServerTLSFromFile(s.opt.TlsPem, s.opt.TlsKey)
			if err != nil {
				initErr = err
				return
			}
			grpcopts = append(grpcopts, grpc.Creds(creds))
		}

		// create grpc server
		s.server = grpc.NewServer(grpcopts...)
		// register services
		riverpb.RegisterRiverServer(s.server, s)
	})
	return initErr
}

func (s *Server) Run() error {
	if s.closed.Load() {
		return ErrClosed
	}

	// init server
	if err := s.init(); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", s.opt.Address)
	if err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("river server listening at %s", s.opt.Address))
	return s.server.Serve(listener)
}

func (s *Server) Close() error {
	if s.closed.Load() {
		return ErrClosed
	}
	s.closed.Store(true)
	s.logger.Info("river server closed")

	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.logger != nil {
		_ = s.logger.Close()
	}

	if s.db != nil {
		_ = s.db.Close()
	}

	return nil
}
