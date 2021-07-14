package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
	"strconv"
)

type RouterContract struct {
	contractapi.Contract
}

// 初始化
func (r *RouterContract) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// 跨链调用请求
func (r *RouterContract) ChainCodeInvoke(ctx contractapi.TransactionContextInterface, from, to, transactionID, stepID, payload, signer string) (string, error) {
	stub := ctx.GetStub()
	data, err := stub.GetState(fmt.Sprintf(RequestKey, transactionID, stepID))
	if err != nil {
		return "", errors.Wrapf(err, "failed to get state by %s", fmt.Sprintf(RequestKey, transactionID, stepID))
	}
	if data != nil {
		return "", errors.New("transactionID is already existed")
	}

	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return "", errors.Wrap(err, "failed to get tx timestamp")
	}

	requestData, err := json.Marshal(ChainCodeInvokeRequest{
		From:          from,
		To:            to,
		TransactionID: transactionID,
		StepID:        stepID,
		Payload:       payload,
		Signer:        signer,
		Timestamp:     timestamp.Seconds,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal ChainCodeInvokeRequest")
	}
	err = stub.PutState(fmt.Sprintf(RequestKey, transactionID, stepID), requestData)
	if err != nil {
		return "", errors.Wrapf(err, "failed to put state by %s, data:%s",
			fmt.Sprintf(RequestKey, transactionID, stepID), string(requestData))
	}

	return transactionID, nil
}

// 跨链回调接口
func (r *RouterContract) ChainCodeInvokeResult(ctx contractapi.TransactionContextInterface, from, to, transactionID, stepID, payload, signer, message string) error {
	stub := ctx.GetStub()
	data, err := stub.GetState(fmt.Sprintf(RequestKey, transactionID, stepID))
	if err != nil {
		return errors.Wrapf(err, "failed to get state by %s", fmt.Sprintf(RequestKey, transactionID, stepID))
	}
	if data == nil {
		return errors.New("transactionID is not found")
	}

	resultData, err := json.Marshal(ChainCodeInvokeResult{
		From:          from,
		To:            to,
		TransactionID: transactionID,
		StepID:        stepID,
		Payload:       payload,
		Signer:        signer,
		Message:       message,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal ChainCodeInvokeResult")
	}
	err = ctx.GetStub().PutState(fmt.Sprintf(CallbackResultKey, transactionID, stepID), resultData)
	if err != nil {
		return errors.Wrapf(err, "failed to put state by %s, data:%s", fmt.Sprintf(CallbackResultKey, transactionID, stepID), string(resultData))
	}
	return nil
}

// 获取跨链调用结果
func (r *RouterContract) QueryInvokeResult(ctx contractapi.TransactionContextInterface, transactionID, stepID string) (ChainCodeInvokeResult, error) {
	result := ChainCodeInvokeResult{}
	data, err := ctx.GetStub().GetState(fmt.Sprintf(CallbackResultKey, transactionID, stepID))
	if err != nil {
		return result, errors.Wrapf(err, "failed to get state by %s", fmt.Sprintf(CallbackResultKey, transactionID, stepID))
	}

	if data == nil {
		return result, nil
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, errors.Wrap(err, "failed to unmarshal ChainCodeInvokeResult")
	}

	return result, nil
}

func bytesToUint64(bts []byte) (uint64, error) {
	return strconv.ParseUint(string(bts), 10, 64)
}

func stringToUint64(uid string) (uint64, error) {
	return strconv.ParseUint(uid, 10, 64)
}

func uint64ToBytes(u uint64) []byte {
	return []byte(strconv.FormatUint(u, 10))
}
