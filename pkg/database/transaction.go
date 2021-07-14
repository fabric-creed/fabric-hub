package database

type Transaction struct {
	PrimaryID int64 `storm:"id,increment" json:"primaryID"`
	// 区块编号
	BlockNumber uint64 `json:"blockNumber"`
	// 区块hash
	BlockHash string `json:"blockHash"`
	// 交易hash
	TransactionHash string `storm:"unique" json:"transactionHash"`
	// 源信息
	OriginInfo []byte `json:"originInfo"`
	//
}
