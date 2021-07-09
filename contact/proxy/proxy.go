package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"strconv"
)

const (
	FnNoTransactionCall = "NoTransactionCall"
)

var logger = shim.NewLogger("proxy")

type Proxy struct{}

func (s *Proxy) Init(stub shim.ChaincodeStubInterface) pb.Response {
	err := stub.PutState(TransactionListLenKey, []byte("0"))
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to init xa task list len key, err:%s", err.Error()))
	}
	err = stub.PutState(TransactionHeadKey, []byte("0"))
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to init xa task head key, err:%s", err.Error()))
	}
	return shim.Success(nil)
}

func (s *Proxy) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case FnNoTransactionCall:
		return noTransactionCall(stub, args)
	default:
		return shim.Error(fmt.Sprintf("Incorrect function: %s", fn))
	}
}

func noTransactionCall(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error(fmt.Sprintf("Failed to put state, incorrect number of arguments: %d", len(args)))
	}
	var lockedContract LockedContract
	isLocked, err := getLockedContract(stub, args[0], &lockedContract)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get locked contract, err:%s", err.Error()))
	}

	if isLocked {
		return shim.Error("resource is locked by unfinished xa transaction: " + lockedContract.TransactionID)
	}
	return callContract(stub, args[0], args[1], args[2])
}

func callContract(stub shim.ChaincodeStubInterface, chainCodeName, fncName, realArgs string) pb.Response {
	var args []string
	err := json.Unmarshal([]byte(realArgs), &args)
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to unmarshal json args, args:%s, err:%s", realArgs, err.Error()))
	}

	var trans [][]byte
	trans = append(trans, []byte(fncName))
	for _, param := range args {
		trans = append(trans, []byte(param))
	}

	return stub.InvokeChaincode(chainCodeName, trans, "")
}

func getLockedContract(stub shim.ChaincodeStubInterface, contract string, lc *LockedContract) (bool, error) {
	state, err := stub.GetState(getLockContractKey(contract))
	if err != nil {
		return false, errors.Wrapf(err, "failed to get state by %s", getLockContractKey(contract))
	}

	if state == nil {
		return false, nil
	} else {
		err = json.Unmarshal(state, lc)
		if err != nil {
			return false, errors.Wrap(err, "failed to unmarshal LockedContract")
		}
		return true, nil
	}
}

func getLockContractKey(contract string) string {
	return fmt.Sprintf(LockContractKey, contract)
}

func getTransactionKey(transactionID string) string {
	return fmt.Sprintf(TransactionKey, transactionID)
}

func getTransactionTaskKey(index uint64) string {
	return fmt.Sprintf(TransactionTaskKey, index)
}

func stringToUint64(str string) uint64 {
	i, e := strconv.Atoi(str)
	if e != nil {
		return 0
	}
	return uint64(i)
}

func stringToInt(str string) int {
	i, e := strconv.Atoi(str)
	if e != nil {
		return 0
	}
	return i
}

func bytesToUint64(bts []byte) uint64 {
	u, err := strconv.ParseUint(string(bts), 10, 64)
	checkError(err)

	return u
}

func uint64ToString(u uint64) string {
	return strconv.FormatUint(u, 10)
}

func uint64ToBytes(u uint64) []byte {
	return []byte(uint64ToString(u))
}
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
