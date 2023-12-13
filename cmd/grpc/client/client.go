package client

import (
	"context"
	"crypto/sha1"
	riverdb "github.com/246859/river"
	"github.com/246859/river/cmd/grpc/riverpb"
	"github.com/246859/river/entry"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

func requirePassInterceptor(pass string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		hash := sha1.New()
		hash.Write([]byte(pass))
		metadata.AppendToOutgoingContext(ctx, "river.requirepass", string(hash.Sum(nil)))
		return nil
	}
}

type Options struct {
	Target string

	Password   string
	ReqTimeout time.Duration

	TlsCert   string
	TlsDomain string
}

func NewClient(ctx context.Context, opt Options) (*Client, error) {
	if len(opt.Target) == 0 {
		return nil, errors.New("target must be specified")
	}

	if opt.ReqTimeout == 0 {
		opt.ReqTimeout = time.Second * 20
	}

	var dialopts []grpc.DialOption

	// tls transport
	if len(opt.TlsCert) > 0 && len(opt.TlsDomain) > 0 {
		creds, err := credentials.NewClientTLSFromFile(opt.TlsCert, opt.TlsDomain)
		if err != nil {
			return nil, err
		}
		dialopts = append(dialopts, grpc.WithTransportCredentials(creds))
	} else {
		dialopts = append(dialopts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// password
	if len(opt.Password) > 0 {
		dialopts = append(dialopts, grpc.WithChainUnaryInterceptor(requirePassInterceptor(opt.Password)))
	}

	dialopts = append(dialopts, grpc.WithBlock())
	conn, err := grpc.DialContext(ctx, opt.Target, dialopts...)
	if err != nil {
		return nil, err
	}
	return &Client{opt: opt, cnn: riverpb.NewRiverClient(conn)}, nil
}

type Client struct {
	opt Options
	cnn riverpb.RiverClient
}

func (c Client) Get(ctx context.Context, key []byte) ([]byte, error) {
	result, err := c.cnn.Get(ctx, &riverpb.RawData{Data: key})
	if err != nil {
		return nil, err
	}

	// if not found
	stat, ok := status.FromError(err)
	if ok && stat.Code() == codes.NotFound {
		return nil, riverdb.ErrKeyNotFound
	}

	return result.Data, nil
}

func (c Client) TTL(ctx context.Context, key []byte) (time.Duration, error) {
	result, err := c.cnn.TTL(ctx, &riverpb.RawData{Data: key})
	if err != nil {
		return 0, err
	}

	// if not found
	stat, ok := status.FromError(err)
	if ok && stat.Code() == codes.NotFound {
		return 0, riverdb.ErrKeyNotFound
	}

	return entry.LeftTTl(result.Ttl), nil
}

func (c Client) Put(ctx context.Context, key []byte, value []byte, ttl time.Duration) (bool, error) {
	result, err := c.cnn.Put(ctx, &riverpb.Record{Key: key, Value: value, Ttl: ttl.Milliseconds()})
	if err != nil {
		return false, err
	}
	return result.Ok, nil
}

func (c Client) Exp(ctx context.Context, key []byte, ttl time.Duration) (bool, error) {
	result, err := c.cnn.Exp(ctx, &riverpb.ExpRecord{Key: key, Ttl: ttl.Milliseconds()})
	if err != nil {
		return false, err
	}
	return result.Ok, nil
}

func (c Client) Del(ctx context.Context, key []byte) (bool, error) {
	result, err := c.cnn.Del(ctx, &riverpb.RawData{Data: key})
	if err != nil {
		return false, err
	}
	return result.Ok, nil
}

func (c Client) Stat(ctx context.Context) (*riverdb.Stats, error) {
	stat, err := c.cnn.Stat(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return &riverdb.Stats{
		KeyNums:    stat.Keys,
		RecordNums: stat.Records,
		DataSize:   stat.Datasize,
		HintSize:   stat.Hintsize,
	}, nil
}
