package config

import (
	"time"
)

// HTTPServer options
type HTTPServer struct {
	Addr    string
	Timeout time.Duration
}

type HTTPClient struct {
	MaxIdleConn int
	IdleTimeout time.Duration
	Compress    bool
	Verify      bool
}

type EndpointUri struct {
	Addr string
}

// Memcache options
type Memcache struct {
	Addrs  []string
	Expire int32
}

// // Redis options
// type Redis struct {
//     redis.Options
// }

// Worker options
type Worker struct {
	Num  int
	Size int
}

// JaegerServer options
type JaegerServer struct {
	Addr string
}

// RPCServer options
type RPCServer struct {
	Addr    string
	Timeout time.Duration
}

// RPCClient options
type RPCClient struct {
	Addr    string
	Timeout time.Duration
}

// KafkaServer options
type KafkaServer struct {
	Addrs    []KafkaAddr
	MaxRetry int
}

// KafkaAddr
type KafkaAddr struct {
	Host string
	Port int
}
