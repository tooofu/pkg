package config

import (
	"time"
)

// HTTPServer options
type HTTPServer struct {
	Addr    string
	Timeout time.Duration
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
