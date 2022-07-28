package server

import (
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	logger *zap.Logger
}

func NewServer(l *zap.Logger) *Server {
	return &Server{logger: l}
}

// RegisterServer registers server as an ExternalProcessorServer on provided GRPC server
func (s *Server) RegisterServer(srv *grpc.Server) {
	pb.RegisterExternalProcessorServer(srv, s)
}

func (s *Server) Process(srv pb.ExternalProcessor_ProcessServer) error {
	return nil
}
