package service

import (
	"context"
	"strings"

	"github.com/sbramin/grpc-demo/cmd/third-party-service/pkg/pb/tps"
	"github.com/utilitywarehouse/grpc-demo/pkg/pb/example"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -package mocks -destination=$GOPATH/src/github.com/sbramin/grpc-demo/internal/mocks/tps.go github.com/sbramin/grpc-demo/cmd/third-party-service/pkg/pb/tps ThirdPartyServiceClient
// SVC implements the example interface and allows embedding additional code
type SVC struct {
	db     string // fake db
	tpsCli tps.ThirdPartyServiceClient
}

// New returns a service with embedded etcd client plus
func New(db string, tpsCli tps.ThirdPartyServiceClient) *SVC {
	return &SVC{
		db,
		tpsCli,
	}
}

// GetExample does nothing
func (s *SVC) GetExample(ctx context.Context, req *example.Request) (*example.Response, error) {
	return &example.Response{Resp: strings.Join([]string{s.db, req.GetReq(), "!"}, "")}, nil
}

func (s *SVC) Echo(ctx context.Context, req *example.Request) (*example.Response, error) {
	resp, err := s.tpsCli.Echo(ctx, &tps.Input{Id: req.GetNo(), SuperMessage: req.GetReq()})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			req.No++
			resp, err = s.tpsCli.Echo(ctx, &tps.Input{Id: req.GetNo(), SuperMessage: req.GetReq()})
			if err != nil {
				return nil, status.Errorf(codes.Internal, "could not echo input(%s), err: %s", req.GetReq(), err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "could not echo input(%s), err: %s", req.GetReq(), err)
		}
	}
	return &example.Response{Resp: resp.GetResp()}, nil
}
