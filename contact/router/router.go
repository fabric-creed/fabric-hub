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
func (r *RouterContract) ChainCodeInvoke(ctx contractapi.TransactionContext, from, to, transactionID, payload string) (string, error) {
	stub := ctx.GetStub()
	data, err := stub.GetState(fmt.Sprintf(RequestKey, transactionID))
	if err != nil {
		return "", errors.Wrapf(err, "failed to get state by %s", fmt.Sprintf(RequestKey, transactionID))
	}
	if data != nil {
		return "", errors.New("transactionID is already existed")
	}

	var payloadData Payload
	err = json.Unmarshal([]byte(payload), &payloadData)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal payload")
	}

	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return "", errors.Wrap(err, "failed to get tx timestamp")
	}

	requestData, err := json.Marshal(ChainCodeInvokeRequest{
		From:          from,
		To:            to,
		TransactionID: transactionID,
		Payload:       []byte(payload),
		Signer:        "",
		Timestamp:     timestamp.Seconds,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal ChainCodeInvokeRequest")
	}
	err = stub.PutState(fmt.Sprintf(RequestKey, transactionID), requestData)
	if err != nil {
		return "", errors.Wrapf(err, "failed to put state by %s, data:%s",
			fmt.Sprintf(RequestKey, transactionID), string(requestData))
	}

	return transactionID, nil
}

// 跨链回调接口
func (r *RouterContract) ChainCodeInvokeResult(ctx contractapi.TransactionContext, from, to, payload, transactionID, signer, code, message string) error {
	stub := ctx.GetStub()
	data, err := stub.GetState(fmt.Sprintf(RequestKey, transactionID))
	if err != nil {
		return errors.Wrapf(err, "failed to get state by %s", fmt.Sprintf(RequestKey, transactionID))
	}
	if data == nil {
		return errors.New("transactionID is not found")
	}

	resultData, err := json.Marshal(ChainCodeInvokeResult{
		From:          from,
		To:            to,
		TransactionID: transactionID,
		Payload:       []byte(payload),
		Signer:        signer,
		Code:          code,
		Message:       message,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal ChainCodeInvokeResult")
	}
	err = ctx.GetStub().PutState(fmt.Sprintf(CallbackResultKey, transactionID), resultData)
	if err != nil {
		return errors.Wrapf(err, "failed to put state by %s, data:%s", fmt.Sprintf(CallbackResultKey, transactionID), string(resultData))
	}
	return nil
}

// 获取跨链调用结果
func (r *RouterContract) QueryInvokeResult(ctx contractapi.TransactionContext, uid string) (ChainCodeInvokeResult, error) {
	result := ChainCodeInvokeResult{}
	id, err := stringToUint64(uid)
	if err != nil {
		return result, errors.Wrap(err, "failed to transfer string to uint64")
	}

	data, err := ctx.GetStub().GetState(fmt.Sprintf(CallbackResultKey, id))
	if err != nil {
		return result, errors.Wrapf(err, "failed to get state by %s", fmt.Sprintf(CallbackResultKey, id))
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
