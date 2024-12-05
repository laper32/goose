package rpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/laper32/goose/logging"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcGatewayTag = opentracing.Tag{Key: string(ext.Component), Value: "grpc-gateway"}

type regHandler func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

type Gateway struct {
	server     *http.Server
	mux        *runtime.ServeMux
	opts       []grpc.DialOption
	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *Server
}

func (s *Server) InitGateway() (gateway *Gateway) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Gateway{
		ctx:        ctx,
		cancel:     cancel,
		mux:        runtime.NewServeMux(),
		grpcServer: s,
		opts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

func (g *Gateway) AddHandler(handler regHandler) {
	err := handler(g.ctx, g.mux, fmt.Sprintf(":%v", g.grpcServer.config.GrpcPort), g.opts)
	if err != nil {
		panic(err)
	}
}

func tracingWrapper(h http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		parentSpanContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		if err == nil || err == opentracing.ErrSpanContextNotFound {
			serverSpan := opentracing.GlobalTracer().StartSpan(
				"ServeHTTP",
				// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
				ext.RPCServerOption(parentSpanContext),
				grpcGatewayTag,
			)
			r = r.WithContext(opentracing.ContextWithSpan(r.Context(), serverSpan))
			defer serverSpan.Finish()
		}
		h.ServeHTTP(w, r)
	}
}

func (g *Gateway) Activate() {
	if g.grpcServer.healthFunc == nil {
		logging.Error("Health check function is not set")
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", HttpHealthHandler(g.grpcServer.healthFunc))
	mux.HandleFunc("/", tracingWrapper(g.mux))
	g.server = &http.Server{
		Addr:    fmt.Sprintf(":%v", g.grpcServer.config.GatewayPort),
		Handler: mux,
	}
	logging.Info(fmt.Sprintf("Serving gRPC-Gateway on 0.0.0.0:%v", g.grpcServer.config.GatewayPort))
	go func() {
		err := g.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (g *Gateway) Close() {
	if g.server != nil {
		if err := g.server.Shutdown(context.Background()); err != nil {
			logging.Error(err)
		}
	}
	g.cancel()
}
