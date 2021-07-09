package client

import (
	"context"
	"github.com/fabric-creed/fabric-hub/global"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"github.com/pkg/errors"
	"time"
)

func (c *HubClient) NoTransactionCall(from, to, transactionID string, signer, payload []byte, timestamp int64) (*pb.CommonResponseMessage, error) {
	if signer == nil {
		fromCSP, ok := global.Config.CSPManager[from]
		if !ok {
			return nil, errors.New("the from id is invalid")
		}
		sign, err := fromCSP.Sign(payload)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to sign payload by %s", from)
		}
		signer = sign
	}
	conn, err := c.client.NewConnection(c.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}

	resp, err := pb.NewHubClient(conn).NoTransactionCall(context.Background(), &pb.NoTransactionCallRequest{
		From:          from,
		To:            to,
		TransactionID: transactionID,
		Payload:       payload,
		Signer:        signer,
		Timestamp:     timestamp,
	})
	if err != nil {
		return nil, err
	}

	toCSP, ok := global.Config.CSPManager[resp.To]
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
