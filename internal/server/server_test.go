package server

import (
	"testing"

	v31 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServer_handleRequestHeaders(t *testing.T) {

	tests := []struct {
		name       string
		reqHeaders *pb.ProcessingRequest_RequestHeaders
		want       *pb.ProcessingResponse
	}{
		{
			name: "XFF present",
			reqHeaders: &pb.ProcessingRequest_RequestHeaders{
				RequestHeaders: &pb.HttpHeaders{
					Headers: &v31.HeaderMap{
						Headers: []*v31.HeaderValue{
							{Key: "x-forwarded-for", Value: "8.8.8.8"},
						},
					},
				},
			},
			want: &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_RequestHeaders{
					RequestHeaders: &pb.HeadersResponse{
						Response: &pb.CommonResponse{
							HeaderMutation: &pb.HeaderMutation{
								SetHeaders: []*v31.HeaderValueOption{
									{Header: &v31.HeaderValue{Key: "x-country-code", Value: "US"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "XFF missing",
			reqHeaders: &pb.ProcessingRequest_RequestHeaders{
				RequestHeaders: &pb.HttpHeaders{
					Headers: &v31.HeaderMap{
						Headers: []*v31.HeaderValue{},
					},
				},
			},
			want: &pb.ProcessingResponse{},
		},
		{
			name: "XFF value missing",
			reqHeaders: &pb.ProcessingRequest_RequestHeaders{
				RequestHeaders: &pb.HttpHeaders{
					Headers: &v31.HeaderMap{
						Headers: []*v31.HeaderValue{
							{Key: "x-forwarded-for", Value: ""},
						},
					},
				},
			},
			want: &pb.ProcessingResponse{},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				logger: zap.NewNop(),
			}
			got := s.handleReqHeaders(tt.reqHeaders)
			require.Equal(t, tt.want, got)
		})
	}
}
