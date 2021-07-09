package fabric

import (
	"encoding/hex"
	"fmt"
	"github.com/fabric-creed/fabric-sdk-go/pkg/client/ledger"
	"github.com/fabric-creed/fabric-sdk-go/pkg/common/providers/fab"
)

type Ledger struct {
	client *ledger.Client
	isGM   bool
}

func (c *Ledger) QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*Block, error) {
	respFrom, err := c.client.QueryBlock(blockNumber, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to call ledger QueryBlock: %v", err)
	}
	respTo, err := DecodeBlock(respFrom, c.isGM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %v", err)
	}
	return respTo, nil
}

func (c *Ledger) QueryBlockByHash(blockHash string, options ...ledger.RequestOption) (*Block, error) {
	blockHashBytes, err := hex.DecodeString(blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode blockHash(%s): %v", blockHash, err)
	}
	respFrom, err := c.client.QueryBlockByHash(blockHashBytes, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to call ledger QueryBlockByHash: %v", err)
	}
	respTo, err := DecodeBlock(respFrom, c.isGM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %v", err)
	}
	return respTo, nil
}

func (c *Ledger) QueryBlockByTxID(txID string, options ...ledger.RequestOption) (*Block, error) {
	respFrom, err := c.client.QueryBlockByTxID(fab.TransactionID(txID), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to call ledger QueryBlockByTxID: %v", err)
	}
	respTo, err := DecodeBlock(respFrom, c.isGM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %v", err)
	}
	return respTo, nil
}

func (c *Ledger) QueryTransaction(txID string, options ...ledger.RequestOption) (*ProcessedTransaction, error) {
	respFrom, err := c.client.QueryTransaction(fab.TransactionID(txID), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to call ledger QueryTransaction: %v", err)
	}
	respTo, err := DecodeProcessedTransaction(respFrom)
	if err != nil {
		return nil, fmt.Errorf("failed to decode processedTransaction: %v", err)
	}
	return respTo, nil
}
