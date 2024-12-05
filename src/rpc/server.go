package rpc

import (
	"fmt"
	"net"
	"time"

	"github.com/laper32/goose/logging"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config     *ServerConfig
	grpcServer *grpc.Server
	healthFunc func() bool
}

type ServerConfig struct {
	Network               string
	GrpcPort              int
	GatewayPort           int
	MaxConnectionIdle     time.Duration
	MaxConnectionAge      time.Duration
	MaxConnectionAgeGrace time.Duration
	KeepAliveInterval     time.Duration
	KeepAliveTimeout      time.Duration
}

var _defaultServerConf = &ServerConfig{
	Network:               "tcp",
	GrpcPort:              2333,
	GatewayPort:           23333,
	MaxConnectionIdle:     60 * time.Second,
	MaxConnectionAge:      2 * time.Hour,
	MaxConnectionAgeGrace: 20 * time.Second,
	KeepAliveInterval:     60 * time.Second,
	KeepAliveTimeout:      20 * time.Second,
}

func NewServer(conf *ServerConfig) *Server {
	if conf == nil {
		conf = _defaultServerConf
	}
	conf.Init()
	return &Server{config: conf}

}

func (conf *ServerConfig) Init() {
	if conf.MaxConnectionIdle <= 0 {
		conf.MaxConnectionIdle = _defaultServerConf.MaxConnectionIdle
	}
	if conf.MaxConnectionAge <= 0 {
		conf.MaxConnectionAge = _defaultServerConf.MaxConnectionAge
	}
	if conf.MaxConnectionAgeGrace <= 0 {
		conf.MaxConnectionAgeGrace = _defaultServerConf.MaxConnectionAgeGrace
	}
	if conf.KeepAliveInterval <= 0 {
		conf.KeepAliveInterval = _defaultServerConf.KeepAliveInterval
	}
	if conf.KeepAliveTimeout <= 0 {
		conf.KeepAliveTimeout = _defaultServerConf.KeepAliveTimeout
	}
	if conf.GrpcPort == 0 {
		conf.GrpcPort = _defaultServerConf.GrpcPort
	}
	if conf.Network == "" {
		conf.Network = _defaultServerConf.Network
	}
	if conf.GatewayPort == 0 {
		conf.GatewayPort = _defaultServerConf.GatewayPort
	}

}

func (s *Server) AddHealthCheck(health func() bool) {
	s.healthFunc = health
}

func (s *Server) Grpc(reg func(s *grpc.Server)) {
	keepParam := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     s.config.MaxConnectionIdle,
		MaxConnectionAge:      s.config.MaxConnectionAge,
		MaxConnectionAgeGrace: s.config.MaxConnectionAgeGrace,
		Time:                  s.config.KeepAliveInterval,
		Timeout:               s.config.KeepAliveTimeout,
	})
	keepPolicy := grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		PermitWithoutStream: true,
	})
	tracer := grpc_opentracing.WithTracer(opentracing.GlobalTracer())
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_opentracing.UnaryServerInterceptor(tracer),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_opentracing.StreamServerInterceptor(tracer),
		)),
		keepParam,
		keepPolicy,
	}

	// Network
	lis, err := net.Listen(s.config.Network, fmt.Sprintf(":%v", s.config.GrpcPort))
	if err != nil {
		panic(err)
	}

	// Initialize
	s.grpcServer = grpc.NewServer(opts...)

	// Register
	if reg != nil {
		reg(s.grpcServer)
	}

	// Health
	grpc_health_v1.RegisterHealthServer(s.grpcServer, s)

	// Register reflection service on gRPC server.
	reflection.Register(s.grpcServer)
	logging.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", s.config.GrpcPort))
	// grpc server
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

}

func (s *Server) Close() {

	s.healthFunc = func() bool { return false }
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}
