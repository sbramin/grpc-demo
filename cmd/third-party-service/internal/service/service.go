package service

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sbramin/grpc-demo/cmd/third-party-service/pkg/pb/tps"
)

// SVC implements the example interface and allows embedding additional code
type SVC struct {
}

// New returns a service with embedded etcd client plus
func New() *SVC {
	return &SVC{}
}

// Echo duplicates its input
func (s *SVC) Echo(ctx context.Context, req *tps.Input) (*tps.Output, error) {
	if req.GetId() < 2 {
		return nil, status.Errorf(codes.InvalidArgument, "cant echo %d time", req.GetId())
	}
	var resp []string
	var i int64
	for ; i < req.GetId(); i++ {
		resp = append(resp, req.SuperMessage)
	}
	return &tps.Output{Resp: strings.Join(resp, " ")}, nil
}
