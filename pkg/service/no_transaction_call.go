package service

import (
	"context"
	"encoding/json"
	"github.com/fabric-creed/fabric-hub/global"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

const (
	FncNoTransactionCall = "NoTransactionCall"
)

func (s *HubService) NoTransactionCall(ctx context.Context, req *pb.NoTransactionCallRequest) (*pb.CommonResponseMessage, error) {
	// 首先判断目的链是否本地
	if _, ok := global.Config.LocalChannelManager[req.To]; !ok {
		return nil, errors.New("the to channel id:[" + req.To + "] is invalid")
	}

	fromCSP, ok := global.Config.CSPManager[req.From]
	if !ok {
		return nil, errors.New("the from channel id is invalid")
	}
	toCSP, ok := global.Config.CSPManager[req.To]
	if !ok {
		return nil, errors.New("the to channel id is invalid")
	}

	// 使用来源链的公钥核实消息签名
	valid, err := fromCSP.Verify(req.Signer, req.Payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to verify signer")
	}
	if !valid {
		return nil, errors.New("signer is invalid")
	}
	var fabricPayload = &pb.FabricPayload{}
	err = json.Unmarshal(req.Payload, fabricPayload)
	if err != nil {
		errors.Wrapf(err, "payload is invalid")
	}
	channelClient, err := global.Config.FabricClientManager[req.To].Channel(fabricPayload.GetChannelID())
	if err != nil || channelClient == nil {
		return nil, errors.Wrapf(err, "failed to get channel by %s", fabricPayload.GetChannelID())
	}

	var args [][]byte
	args = append(args, []byte(fabricPayload.ChainCodeName))
	args = append(args, []byte(fabricPayload.FncName))
	if fabricPayload.Args == nil {
		fabricPayload.Args = []string{}
	}
	ccArgs, err := json.Marshal(fabricPayload.Args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal args:[%v]", fabricPayload.Args)
	}
	args = append(args, ccArgs)

	resp, err := channelClient.ChannelExecute(channel.Request{
		ChaincodeID: global.Config.LocalChannelManager[req.To].ProxyChainCodeName,
		Fcn:         FncNoTransactionCall,
		Args:        args,
		IsInit:      false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call channel execute")
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal resp:[%v]", resp)
	}

	// 用该链的私钥对payload进行签名
	sig, err := toCSP.Sign(payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sign payload by %s ", req.To)
	}
	return &pb.CommonResponseMessage{
		From:          req.From,
		To:            req.To,
		TransactionID: req.TransactionID,
		Payload:       payload,
		Signer:        sig,
	}, nil
}
