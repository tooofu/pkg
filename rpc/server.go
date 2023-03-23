package rpc

import (
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
)

//  1. grpcServerKeepAliveTime < grpcClientKeepAliveTime
//     rpcSvr 主动发起 ping
//  2. grpcMaxConnectionIdle > grpcServerKeepAliveTime * 3
//     TODO idle > ping 间隔 * 3, ping 一次超时不应该直接 goaway
//  3. grpcMaxConnectionAge
//     NOTE 一个 conn 使用最长时间, emmmm. 这个意义?
const (
	grpcMaxConnectionIdle      = time.Second * 600
	grpcMaxConnectionAge       = time.Second * 3600
	grpcMaxConnectionAgeGrace  = time.Second * 5
	grpcServerKeepAliveTime    = time.Second * 5
	grpcServerKeepAliveTimeout = time.Second * 1
)

var (
	keepEnforce = grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		// 客户端 ping 的间隔应该不小于这个时长, 默认是5分钟
		MinTime: 10 * time.Second,

		// 服务端是否允许在没有RPC调用时发送PING, 默认不允许
		// 在不允许的情况下, 客户端发送了PING, 服务端将发送 GOAWAY 帧, 关闭连接
		PermitWithoutStream: true,
	})

	// grpcClientKeepAliveTime > grpcServerKeepAliveTime + grpcServerKeepAliveTimeout
	keepParams = grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     grpcMaxConnectionIdle,
		MaxConnectionAge:      grpcMaxConnectionAge,
		MaxConnectionAgeGrace: grpcMaxConnectionAgeGrace,
		Time:                  grpcServerKeepAliveTime,
		Timeout:               grpcServerKeepAliveTimeout,
	})
)

type svrOptions struct {
	log *logrus.Logger
	tr  opentracing.Tracer
}

type ServerOptions func(*svrOptions)

// func ServerLogger(log *logrus.Logger) ServerOptions {
//     return func(options *svrOptions) {
//         options.log = log
//     }
// }

func ServerTracer(tr opentracing.Tracer) ServerOptions {
	return func(options *svrOptions) {
		options.tr = tr
	}
}

func recoveryInterceptor() grpc_recovery.Option {
	return grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		return grpc.Errorf(codes.Unknown, "panic triggered=%v", p)
	})
}

func logLevel(code codes.Code) logrus.Level {
	if code == codes.OK {
		return logrus.InfoLevel
	}
	return logrus.ErrorLevel
}

// InitServer init server
func InitServer(log *logrus.Logger, options ...ServerOptions) *grpc.Server {
	opts := svrOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// log
	logrusEntry := logrus.NewEntry(log)
	logOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(logLevel),
	}
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)

	// ctx
	// ctxOpts := []grpc_ctxtags.Option{
	//     grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.TagBasedRequestFieldExtractor("")),
	// }

	interceptors := grpc_middleware.ChainUnaryServer(
		// ctxtag
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		// grpc_ctxtags.UnaryServerInterceptor(ctxOpts...),
		// log
		grpc_logrus.UnaryServerInterceptor(logrusEntry, logOpts...),
		// // recovery
		// grpc_recovery.UnaryServerInterceptor(recoveryInterceptor()),
	)

	if opts.tr != nil {
		interceptors = grpc_middleware.ChainUnaryServer(
			interceptors,
			otgrpc.OpenTracingServerInterceptor(opts.tr),
		)
	}

	interceptors = grpc_middleware.ChainUnaryServer(
		interceptors,
		// recovery
		grpc_recovery.UnaryServerInterceptor(recoveryInterceptor()),
	)

	s := grpc.NewServer(keepEnforce, keepParams, grpc.UnaryInterceptor(interceptors))
	return s
}
