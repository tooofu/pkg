package rpc

import (
	"context"
	"time"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	// "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

const (
	grpcInitWinSize            = 1 << 24
	grpcInitConnWinSize        = 1 << 24
	grpcMaxRecvMsgSize         = 1 << 24
	grpcMaxSendMsgSize         = 1 << 24
	grpcBackoffMaxDelay        = time.Second * 3
	grpcClientKeepAliveTime    = time.Second * 10
	grpcClientKeepAliveTimeout = time.Second * 1
)

const (
	defaultServiceConfig = `{"LoadBalancingPolicy": "%s"}`
)

type cliOptions struct {
	tr opentracing.Tracer
}

type ClientOptions func(*cliOptions)

func ClientTracer(tr opentracing.Tracer) ClientOptions {
	return func(options *cliOptions) {
		options.tr = tr
	}
}

// InitConn init conn
func InitConn(c context.Context, addr string, options ...ClientOptions) (conn *grpc.ClientConn, err error) {
	opts := cliOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// func NewClientConn(addr string, conf *RPCClient) grpc.ClientConnInterface {
	// c, cancel := context.WithTimeout(context.Background(), time.Duration(conf.Dial))
	// defer cancel()
	// conn, err = grpc.DialContext(c, addr,
	//     []grpc.DialOption{
	//         // grpc.WithBalancerName(roundrobin.Name),
	//         // grpc.WithDefaultServiceConfig(fmt.Sprintf(defaultServiceConfig, roundrobin.Name)),
	//         grpc.WithInsecure(),
	//         grpc.WithInitialWindowSize(grpcInitWinSize),
	//         grpc.WithInitialConnWindowSize(grpcInitConnWinSize),
	//         grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxRecvMsgSize)),
	//         grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
	//         grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
	//         grpc.WithKeepaliveParams(keepalive.ClientParameters{
	//             Time:                grpcClientKeepAliveTime,
	//             Timeout:             grpcClientKeepAliveTimeout,
	//             PermitWithoutStream: true,
	//         }),
	//     }...)

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithInitialWindowSize(grpcInitWinSize),
		grpc.WithInitialConnWindowSize(grpcInitConnWinSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxRecvMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
		grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                grpcClientKeepAliveTime,
			Timeout:             grpcClientKeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}

	if opts.tr != nil {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opts.tr)))
	}
	conn, err = grpc.DialContext(c, addr, dialOpts...) // grpc.WithBalancerName(roundrobin.Name),
	// grpc.WithDefaultServiceConfig(fmt.Sprintf(defaultServiceConfig, roundrobin.Name)),

	return
}
