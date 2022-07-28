package server

import (
	"io"

	v31 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	s.logger.Debug("new stream")
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		resp := &pb.ProcessingResponse{}
		switch v := req.Request.(type) {
		case *pb.ProcessingRequest_RequestHeaders:
			s.logger.Debug("pb.ProcessingRequest_RequestHeaders")
			r := req.Request
			h := r.(*pb.ProcessingRequest_RequestHeaders)
			resp = s.handleRequestHeaders(h)
			break
		default:
			s.logger.Error("unknown request type", zap.Any("req", v))
		}
		if err := srv.Send(resp); err != nil {
			s.logger.Error("unable to send response", zap.Error(err))
		}
	}
}

func (s *Server) handleRequestHeaders(h *pb.ProcessingRequest_RequestHeaders) *pb.ProcessingResponse {
	// extract ip from headers
	for _, v := range h.RequestHeaders.GetHeaders().GetHeaders() {
		if v.GetKey() == "x-forwarded-for" {
			s.logger.Debug("XFF header found", zap.String("xff", v.GetValue()))
		}
	}
	// TODO: maxmind magic
	countryCode := "US"
	// add x-country-code header to resp
	if countryCode != "" {
		return countryCodeResp(countryCode)
	}

	return &pb.ProcessingResponse{}
}

func countryCodeResp(countryCode string) *pb.ProcessingResponse {
	opt := &v31.HeaderValueOption{
		Header: &v31.HeaderValue{
			Key:   "x-country-code",
			Value: countryCode,
		},
	}

	return &pb.ProcessingResponse{
		Response: &pb.ProcessingResponse_RequestHeaders{
			RequestHeaders: &pb.HeadersResponse{
				Response: &pb.CommonResponse{
					HeaderMutation: &pb.HeaderMutation{
						SetHeaders: []*v31.HeaderValueOption{opt},
					},
				},
			},
		},
	}
}
