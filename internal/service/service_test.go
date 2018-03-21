package service_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbramin/grpc-demo/internal/mocks"
	"github.com/sbramin/grpc-demo/internal/service"
)

var ts testSuite

type testSuite struct {
	ctx  context.Context
	svc  *service.SVC
	mock *mockz
}

type mockz struct {
	tps *mocks.MockThirdPartyServiceClient
	db  string
}

// TestSetup handles the test suite setup
func TestSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTPS := mocks.NewMockThirdPartyServiceClient(ctrl)
	db := "joe"
	svc := service.New(db, mockTPS)

	ts = testSuite{
		ctx,
		svc,
		&mockz{
			tps: mockTPS,
		},
	}
}
