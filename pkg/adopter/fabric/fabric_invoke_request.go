package fabric

import (
	"encoding/json"
	"fmt"
	"github.com/fabric-creed/fabric-hub/global"
	cc "github.com/fabric-creed/fabric-hub/pkg/adopter"
	"github.com/fabric-creed/fabric-hub/pkg/client"
	"github.com/fabric-creed/fabric-hub/pkg/database"
	"github.com/fabric-creed/fabric-hub/pkg/fabric"
	"github.com/fabric-creed/fabric-hub/pkg/modules/block"
	"github.com/fabric-creed/fabric-hub/pkg/modules/transaction"
	"github.com/fabric-creed/fabric-hub/pkg/protos/pb"
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	FuncChainCodeInvoke       = "ChainCodeInvoke"
	FuncChainCodeInvokeResult = "ChainCodeInvokeResult"
)

type Fabric struct {
	dbPath              string
	fab                 *fabric.Client
	hubClientMap        map[string]*client.HubClient
	channelID           string
	isGM                bool
	blockNum            uint64
	routerChainCodeName string
}

func NewFabric(dbPath string, channelID, routerChainCodeName string, isGM bool, options ...Option) *Fabric {
	fabric := &Fabric{
		dbPath:              dbPath,
		channelID:           channelID,
		isGM:                isGM,
		routerChainCodeName: routerChainCodeName,
	}
	for _, f := range options {
		f(fabric)
	}
	return fabric
}

type Option func(f *Fabric)

func WithFabricClient(fab *fabric.Client) Option {
	return func(f *Fabric) {
		f.fab = fab
	}
}

func WithHubClientMap(hubClientMap map[string]*client.HubClient) Option {
	return func(f *Fabric) {
		f.hubClientMap = hubClientMap
	}
}

func (f *Fabric) FetchNextBlock() (*cc.BlockInfo, error) {
	if f.blockNum == 0 {
		blockNum, err := block.NewController(f.dbPath).FetchLatestBlockNum()
		if err != nil {
			logrus.Errorf("failed to fetch latest block num, err:%s", err.Error())
			return nil, err
		}
		f.blockNum = blockNum
	}
	for {
		data, err := f.queryBlock(f.blockNum + 1)
		if err != nil {
			if !strings.Contains(err.Error(), "Entry not found in index") {
				logrus.Errorf("failed to handle(%v): %v", f.blockNum+1, err)
			}
			time.Sleep(2 * time.Second)
		} else {
			logrus.Infof("succeeded to handle(%v)", f.blockNum+1)
			f.blockNum++
			return data, nil
		}
	}
}

func (f *Fabric) HandleCrossChainRequest(request interface{}) (*cc.CrossChainResponse, error) {
	response := &cc.CrossChainResponse{}
	tctl := transaction.NewController(f.dbPath)
	if fccr, ok := request.(cc.FabricCrossChainRequest); ok {
		// 首先判断txHash是否已经存在了,存在则跳过
		_, err := tctl.FetchTransactionByTransactionHash(fccr.TxHash)
		if err == nil {
			return nil, nil
		}
		
		switch fccr.Request.(type) {
		case *pb.NoTransactionCallRequest:
			req := fccr.Request.(*pb.NoTransactionCallRequest)
			if _, ok := f.hubClientMap[req.To]; ok {
				resp, err := f.hubClientMap[req.To].NoTransactionCall(req)
				if err != nil {
					logrus.Errorf("failed to call no transaction, err:%s", err.Error())
					response.ErrorMessage = err.Error()
					return response, nil
				}
				response.Response = resp
			} else {
				response.ErrorMessage = fmt.Sprintf("the %s client is not found", req.To)
			}
		default:
			return nil, errors.New("invalid cross chain request")
		}

		err = tctl.Create(fccr.BlockNumber, fccr.BlockHash, fccr.TxHash, fccr.OriginInfo)
		if err != nil {
			logrus.Errorf("failed to create transaction, err:%s", err.Error())
			return nil, err
		}

		return response, nil
	}

	return nil, errors.New("invalid fabric cross chain request")
}

func (f *Fabric) HandleCrossChainCallbackRequest(response cc.CrossChainResponse) error {
	cl, err := f.fab.Channel(f.channelID)
	if err != nil {
		return err
	}
	switch response.Response.(type) {
	case *pb.CommonResponseMessage:
		msg := response.Response.(*pb.CommonResponseMessage)
		_, err = cl.ChannelExecute(channel.Request{
			ChaincodeID: f.routerChainCodeName,
			Fcn:         FuncChainCodeInvokeResult,
			Args: [][]byte{
				[]byte(msg.From),
				[]byte(msg.To),
				[]byte(msg.TransactionID),
				[]byte(msg.StepID),
				msg.Payload,
				msg.Signer,
				[]byte(response.ErrorMessage),
			},
			IsInit: false,
		})
		if err != nil {
			return errors.Wrap(err, "failed to call router invoke result")
		}

		if msg.Callback != nil {
			var callback pb.FabricCallback
			err = json.Unmarshal(msg.Callback, &callback)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal fabric callback")
			}
			if callback.CallbackChainCodeName != "" {
				var args [][]byte
				for i := range callback.CallbackArgs {
					args = append(args, []byte(callback.CallbackArgs[i]))
				}
				_, err = cl.ChannelExecute(channel.Request{
					ChaincodeID: callback.CallbackChainCodeName,
					Fcn:         callback.CallbackFncName,
					Args:        args,
					IsInit:      false,
				})
				if err != nil {
					return errors.Wrapf(err, "failed to execute call back chain code:%s", callback.CallbackChainCodeName)
				}
			}

		}
	}

	return nil
}

func (f *Fabric) SaveLatestBlock(blockData []byte) error {
	var pbBlock fabric.Block
	err := json.Unmarshal(blockData, &pbBlock)
	if err != nil {
		return err
	}
	dbBlock := &database.Block{
		BlockNumber:  pbBlock.Header.Number,
		PreviousHash: pbBlock.Header.PreviousHash,
		DataHash:     pbBlock.Header.DataHash,
		BlockHash:    pbBlock.BlockHash,
	}
	blockTime := time.Unix(pbBlock.BlockTime, 0)
	dbBlock.BlockTime = blockTime
	data, err := json.Marshal(pbBlock.Data)
	if err != nil {
		return err
	}
	dbBlock.OriginInfo = data
	dbBlock.TxNum = len(pbBlock.Data.Data)

	err = block.NewController(global.Config.DBPath).CreateBlock(dbBlock)
	if err != nil {
		return err
	}

	return nil
}

func (f *Fabric) queryBlock(blockNumber uint64) (*cc.BlockInfo, error) {
	ledger, err := f.fab.Ledger(f.channelID, f.isGM)
	if err != nil {
		return nil, err
	}
	pbBlock, err := ledger.QueryBlock(blockNumber)
	if err != nil {
		return nil, err
	}
	if pbBlock.Header == nil {
		return nil, fmt.Errorf("block header should not be nil")
	}

	var requests []interface{}
	if pbBlock.Data != nil {
		for i := range pbBlock.Data.Data {
			request, txHash, err := parseEnvelope(pbBlock.Data.Data[i])
			if err != nil {
				return nil, err
			}
			if request != nil {
				requests = append(requests, cc.FabricCrossChainRequest{
					TxHash:      txHash,
					BlockNumber: pbBlock.Header.Number,
					BlockHash:   pbBlock.BlockHash,
					OriginInfo:  pbBlock.OriginData,
					Request:     request,
				})
			}
		}
	}
	blockData, err := json.Marshal(*pbBlock)
	if err != nil {
		return nil, err
	}

	return &cc.BlockInfo{
		BlockData:          blockData,
		CrossChainRequests: requests,
	}, nil
}

func parseEnvelope(envelope *fabric.Envelope) (interface{}, string, error) {
	if envelope.Payload == nil {
		return nil, "", nil
	}
	payload := envelope.Payload
	if payload.Header == nil {
		return nil, "", nil
	}
	header := payload.Header
	if header.ChannelHeader == nil {
		return nil, "", nil
	}
	txHash := header.ChannelHeader.TxId
	if payload.Transaction == nil {
		return nil, txHash, nil
	}
	if len(payload.Transaction.Actions) == 0 {
		return nil, txHash, nil
	}
	action := payload.Transaction.Actions[0]
	if action.Payload == nil {
		return nil, txHash, nil
	}
	if action.Payload.ChaincodeProposalPayload == nil {
		return nil, txHash, nil
	}
	chainCodeProposalPayload := action.Payload.ChaincodeProposalPayload
	if chainCodeProposalPayload.Input == nil {
		return nil, txHash, nil
	}
	input := chainCodeProposalPayload.Input
	if input.ChaincodeSpec == nil {
		return nil, txHash, nil
	}

	if input.ChaincodeSpec.Input == nil {
		return nil, txHash, nil
	}
	inputArgs := input.ChaincodeSpec.Input.Args
	logrus.Infof("========input args:%v", inputArgs)
	if len(inputArgs) == 0 {
		return nil, txHash, nil
	}
	switch inputArgs[0] {
	case FuncChainCodeInvoke:
		if len(inputArgs) != 7 {
			return nil, txHash, errors.New("input args is not equal 7")
		}
		req := &pb.NoTransactionCallRequest{
			From:          inputArgs[1],
			To:            inputArgs[2],
			TransactionID: inputArgs[3],
			StepID:        inputArgs[4],
			Payload:       []byte(inputArgs[5]),
			Signer:        []byte(inputArgs[6]),
			Timestamp:     time.Now().Unix(),
		}
		return req, txHash, nil
	default:
		return nil, txHash, nil
	}

	return nil, txHash, nil
}
