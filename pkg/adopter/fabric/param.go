package fabric

type FabricCrossChainRequest struct {
	TxHash      string
	BlockNumber uint64
	BlockHash   string
	OriginInfo  []byte
	Request     interface{}
}
