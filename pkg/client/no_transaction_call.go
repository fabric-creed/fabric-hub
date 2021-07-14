package client

import (
	"context"
	"fmt"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"github.com/fabric-creed/grpc/codes"
	"github.com/fabric-creed/grpc/status"
	"github.com/pkg/errors"
	"time"
)

func (c *HubClient) NoTransactionCall(request *pb.NoTransactionCallRequest) (*pb.CommonResponseMessage, error) {
	if request.Signer == nil {
		fromCSP, ok := c.csp[request.From]
		if !ok {
			return nil, errors.New("the from id is invalid")
		}
		sign, err := fromCSP.Sign(request.Payload)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to sign payload by %s", request.From)
		}
		request.Signer = sign
	}
	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	retryTime := 5
retry:
	conn, err := c.client.NewConnection(fmt.Sprintf("%s:%d", c.address, c.port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp, err := pb.NewHubClient(conn).NoTransactionCall(context.Background(), request)
	if err != nil {
		statu, ok := status.FromError(err)
		if ok {
			// 判断是否为调用超时
			if statu.Code() == codes.DeadlineExceeded {
				if retryTime > 0 {
					retryTime--
					conn.Close()
					goto retry
				}
			}
		}
		return nil, err
	}

	toCSP, ok := c.csp[resp.To]
	if !ok {
		return nil, errors.New("the to channel id is invalid")
	}
	// 使用目的链的公钥核实消息签名
	valid, err := toCSP.Verify(resp.Signer, resp.Payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to verify signer")
	}
	if !valid {
		return nil, errors.New("the message from sever is in invalid")
	}

	return resp, nil
}
