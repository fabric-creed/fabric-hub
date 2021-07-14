package main

const (
	RequestKey        = "request-%s-%s"
	CallbackResultKey = "callbackResult-%s-%s"
)

type ChainCodeInvokeRequest struct {
	// 来源链ID
	From string `json:"from"`
	// 目标链ID
	To string `json:"to"`
	// 交易ID
	TransactionID string `json:"transactionId"`
	// 步骤ID
	StepID string `json:"transactionId"`
	// 跨链内容
	Payload string `json:"payload"`
	// 签名
	Signer string `json:"signer"`
	// 时间戳
	Timestamp int64 `json:"timestamp"`
}

type ChainCodeInvokeResult struct {
	// 来源链ID
	From string `json:"from"`
	// 目标链ID
	To string `json:"to"`
	// 交易ID
	TransactionID string `json:"transactionId"`
	// 步骤ID
	StepID string `json:"stepID"`
	// 跨链内容
	Payload string `json:"payload"`
	// 签名
	Signer string `json:"signer"`
	// 错误信息
	Message string `json:"message"`
}
