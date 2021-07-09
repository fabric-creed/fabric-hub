package service

import (
	"context"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
)

const (
	FncStartTransaction = "StartTransaction"
)

func (s *HubService) StartTransaction(ctx context.Context, req *pb.StartTransactionRequest) (*pb.CommonResponseMessage, error) {
	return nil, nil
}
