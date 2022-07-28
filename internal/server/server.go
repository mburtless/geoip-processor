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

var (
	defaultIPReqHeader  = "x-forwarded-for"
	defaultCCRespHeader = "x-country-code"
)

type Server struct {
	logger       *zap.Logger
	ipReqHeader  string
	ccRespHeader string
}

func NewServer(l *zap.Logger, opts ...func(s *Server)) *Server {
	svr := &Server{
		logger:       l,
		ipReqHeader:  defaultIPReqHeader,
		ccRespHeader: defaultCCRespHeader,
	}
	for _, opt := range opts {
		opt(svr)
	}
	return svr
}

// WithIPReqHeader configures the header that IP of request is extracted from
func WithIPReqHeader(h string) func(s *Server) {
	return func(s *Server) {
		s.ipReqHeader = h
	}
}

// WithCCRespHeader configures the header that country code of request is injected in
func WithCCRespHeader(h string) func(s *Server) {
	return func(s *Server) {
		s.ccRespHeader = h
	}
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
			resp = s.handleReqHeaders(h)
			break
		default:
			s.logger.Error("unknown request type", zap.Any("req", v))
		}
		if err := srv.Send(resp); err != nil {
			s.logger.Error("unable to send response", zap.Error(err))
		}
	}
}

func (s *Server) handleReqHeaders(h *pb.ProcessingRequest_RequestHeaders) *pb.ProcessingResponse {
	ip := s.extractIPFromReqHeaders(h.RequestHeaders.GetHeaders().GetHeaders())

	// if ip was extracted add x-country-code header to resp
	if ip != "" {
		// TODO: maxmind magic
		countryCode := "US"
		if countryCode != "" {
			return s.countryCodeResp(countryCode)
		}
	}

	return &pb.ProcessingResponse{}
}

func (s *Server) extractIPFromReqHeaders(h []*v31.HeaderValue) string {
	ip := ""
	for _, v := range h {
		if v.GetKey() == s.ipReqHeader {
			s.logger.Debug("ip header found", zap.String(s.ipReqHeader, v.GetValue()))
			ip = v.GetValue()
		}
	}
	return ip
}

func (s *Server) countryCodeResp(countryCode string) *pb.ProcessingResponse {
	if countryCode == "" {
		return &pb.ProcessingResponse{}
	}

	h := &v31.HeaderValueOption{
		Header: &v31.HeaderValue{
			Key:   s.ccRespHeader,
			Value: countryCode,
		},
	}

	return &pb.ProcessingResponse{
		Response: &pb.ProcessingResponse_RequestHeaders{
			RequestHeaders: &pb.HeadersResponse{
				Response: &pb.CommonResponse{
					HeaderMutation: &pb.HeaderMutation{
						SetHeaders: []*v31.HeaderValueOption{h},
					},
				},
			},
		},
	}
}
