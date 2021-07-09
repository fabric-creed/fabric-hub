package database

import "time"

type Block struct {
	PrimaryID int64 `storm:"id,increment" json:"primaryID"`
	// 区块编号
	BlockNumber uint64 `storm:"unique" json:"blockNumber"`
	// 前驱哈希
	PreviousHash string `json:"previousHash"`
	// 后驱哈希
	NextHash string `json:"nextHash"`
	// 数据哈希
	DataHash string `storm:"unique" json:"dataHash"`
	// 区块哈希
	BlockHash string `storm:"unique" json:"blockHash"`
	// 出块时间
	BlockTime time.Time `json:"blockTime"`
	// 区块中的交易量
	TxNum int `json:"txNum"`
	// 源信息
	OriginInfo []byte `json:"originInfo"`
}
