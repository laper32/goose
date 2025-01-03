package rpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/laper32/goose/ecode"
	"github.com/laper32/goose/logging"

	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func (s *Server) Check(ctx context.Context, in *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	status := health.HealthCheckResponse_UNKNOWN
	if s.healthFunc != nil {
		if s.healthFunc() {
			status = health.HealthCheckResponse_SERVING
		} else {
			status = health.HealthCheckResponse_NOT_SERVING
		}
	}
	return &health.HealthCheckResponse{Status: status}, nil
}

func (s *Server) Watch(in *health.HealthCheckRequest, _ health.Health_WatchServer) error {
	// Example of how to register both methods but only implement the Check method.
	return ecode.Errorf(codes.Unimplemented, "unimplemented")
}

// HttpHealthHandler http health check handler
func HttpHealthHandler(health func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if health() {
			if _, err := fmt.Fprintln(w, "Healthy"); err != nil {
				logging.Error(err)
			}
		} else {
			http.Error(w, "Unhealthy", http.StatusServiceUnavailable)
		}
	}
}
