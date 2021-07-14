package main

const (
	RollbackFlag     = "Rollback"
	SuccessFlag      = "success"
	StatusProcessing = "processing"
	StatusCommitted  = "committed"
	StatusRollback   = "rollback"

	TransactionListLenKey = "TransactionLen"
	TransactionHeadKey    = "TransactionTaskHead"
	LockContractKey       = "Contract-%s"         // %s: chain code name
	TransactionKey        = "Transaction-%s-info" // %s: xa transaction id
	TransactionTaskKey    = "Transaction-%d-task" // %d: index
)

type TransactionStep struct {
	Seq       uint64 `json:"seq"`
	Identity  string `json:"identity"`
	Timestamp uint64 `json:"timestamp"`
	FuncName  string `json:"funcName"`
	Args      string `json:"args"`
}

type Transaction struct {
	TransactionID     string            `json:"transactionID"`
	Identity          string            `json:"identity"`
	Contracts         []string          `json:"contracts"`
	Status            string            `json:"status"`
	StartTimestamp    uint64            `json:"startTimestamp"`
	CommitTimestamp   uint64            `json:"commitTimestamp"`
	RollbackTimestamp uint64            `json:"rollbackTimestamp"`
	Seqs              []uint64          `json:"seqs"`
	TransactionSteps  []TransactionStep `json:"transactionSteps"`
}

type LockedContract struct {
	TransactionID string `json:"transactionID"`
}

type GetTransactionStateResponse struct {
	TransactionID string `json:"transactionID"`
	Seq           uint64 `json:"seq"`
}
