package fabric

import (
	"encoding/json"
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/fabric-creed/fabric-hub/global"
	"github.com/fabric-creed/fabric-hub/pkg/client"
	"github.com/fabric-creed/fabric-hub/pkg/database"
	"github.com/fabric-creed/fabric-hub/pkg/fabric"
	"github.com/fabric-creed/fabric-hub/pkg/modules/block"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const FuncChainCodeInvoke = "ChainCodeInvoke"

type Task struct {
	db        *storm.DB
	fab       *fabric.Client
	channelID string
	isGM      bool
	hubClient *client.HubClient
}

func (t *Task) Run() {
	nextBlockNumber, err := t.fetchNextBlockNumber()
	if err != nil {
		logrus.Errorf("failed to fetch nextBlockNumber: %v", err)
		return
	}
	logrus.Debugf("nextBlockNumber %d", nextBlockNumber)

	for {
		err := t.handle(nextBlockNumber)
		if err != nil {
			if !strings.Contains(err.Error(), "Entry not found in index") {
				logrus.Errorf("failed to handle(%v): %v", nextBlockNumber, err)
			}
			time.Sleep(2 * time.Second)
		} else {
			logrus.Infof("succeeded to handle(%v)", nextBlockNumber)
			nextBlockNumber++
		}
	}
}

func (t *Task) fetchNextBlockNumber() (uint64, error) {
	blockNum, err := block.NewController(global.Config.DBPath).FetchLatestBlockNum()
	if err != nil {
		return 0, err
	}

	return blockNum + 1, nil
}

func (t *Task) handle(blockNumber uint64) error {
	ledger, err := t.fab.Ledger(t.channelID, t.isGM)
	if err != nil {
		return err
	}
	pbBlock, err := ledger.QueryBlock(blockNumber)
	if err != nil {
		return err
	}
	if pbBlock.Header == nil {
		return fmt.Errorf("block header should not be nil")
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

	if pbBlock.Data != nil {
		for i := range pbBlock.Data.Data {
			if req := t.parseEnvelope(pbBlock.Data.Data[i]); req != nil {
				resp,err:=t.hubClient.NoTransactionCall(req.From, req.To, req.TransactionID, nil, req.Payload, req.Timestamp)
				if err!=nil {
					
				}
			}
		}
	}

	return nil
}

func (t *Task) parseEnvelope(envelope *fabric.Envelope) *ChainCodeInvokeRequest {
	if envelope.Payload == nil {
		return nil
	}
	payload := envelope.Payload
	if payload.Header == nil {
		return nil
	}
	header := payload.Header
	if header.ChannelHeader == nil {
		return nil
	}
	channelHeader := header.ChannelHeader
	timestamp := channelHeader.Timestamp
	if payload.Transaction == nil {
		return nil
	}
	if len(payload.Transaction.Actions) == 0 {
		return nil
	}
	action := payload.Transaction.Actions[0]
	if action.Payload == nil {
		return nil
	}
	if action.Payload.ChaincodeProposalPayload == nil {
		return nil
	}
	chaincodeProposalPayload := action.Payload.ChaincodeProposalPayload
	if chaincodeProposalPayload.Input == nil {
		return nil
	}
	input := chaincodeProposalPayload.Input
	if input.ChaincodeSpec == nil {
		return nil
	}

	if input.ChaincodeSpec.Input == nil {
		return nil
	}
	inputArgs := input.ChaincodeSpec.Input.Args
	logrus.Infof("========input args:%v", inputArgs)
	if len(inputArgs) == 0 {
		return nil
	}
	switch inputArgs[0] {
	case FuncChainCodeInvoke:
		if len(inputArgs) != 5 {
			return nil
		}
		return &ChainCodeInvokeRequest{
			From:          inputArgs[1],
			To:            inputArgs[2],
			TransactionID: inputArgs[3],
			Payload:       []byte(inputArgs[4]),
			Timestamp:     timestamp,
		}
	default:
		return nil
	}

	return nil
}
