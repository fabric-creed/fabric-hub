package fabric

type ChainCodeInvokeRequest struct {
	// 来源链ID
	From string `json:"from"`
	// 目标链ID
	To string `json:"to"`
	// 交易ID
	TransactionID string `json:"transactionID"`
	// 跨链内容
	Payload []byte `json:"payload"`
	// 签名
	Signer string `json:"signer"`
	// 时间戳
	Timestamp int64 `json:"timestamp"`
}

type Payload struct {
	// 目标链通道名称
	ChannelName string `json:"channelName"`
	// 目标链方法名称
	FuncName string `json:"funcName"`
	// 目标链方法所需参数
	Args string `json:"args"`
	// 源链回调的通道名称
	CallbackChannelName string `json:"callbackChannelName"`
	// 源链回调方法名
	CallbackFuncName string `json:"callbackFuncName"`
	// 源链回调方法所需参数
	CallbackArgs string `json:"callbackArgs"`
}
