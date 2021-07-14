package adopter

import (
	"github.com/sirupsen/logrus"
	"time"
)

type CrossChain interface {
	// 解析最新区块信息
	FetchNextBlock() (*BlockInfo, error)
	// 处理跨链请求
	HandleCrossChainRequest(request interface{}) (*CrossChainResponse, error)
	// 处理跨链回调及请求记录
	HandleCrossChainCallbackRequest(payload CrossChainResponse) error
	// 保存最新区块信息
	SaveLatestBlock(blockData []byte) error
}

type BlockInfo struct {
	BlockData          []byte
	CrossChainRequests []interface{}
}

type CrossChainResponse struct {
	Response              interface{}
	ErrorMessage          string
	CallbackChannelName   string
	CallbackChainCodeName string
	CallbackFncName       string
	CallbackArgs          []byte
}

type FabricCrossChainRequest struct {
	TxHash      string
	BlockNumber uint64
	BlockHash   string
	OriginInfo  []byte
	Request     interface{}
}

type CrossChainTask struct {
	cc CrossChain
}

func NewCrossChainTask(cc CrossChain) *CrossChainTask {
	return &CrossChainTask{cc: cc}
}

func (t *CrossChainTask) Run() error {
	for {
		block, err := t.cc.FetchNextBlock()
		if err != nil {
			logrus.Errorf("failed to parse cross chain request, err:%s", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		for _, request := range block.CrossChainRequests {
		retry:
			response, err := t.cc.HandleCrossChainRequest(request)
			if err != nil {
				logrus.Errorf("failed to handle cross chain request, err:%s", err.Error())
				time.Sleep(2 * time.Second)
				goto retry
			}
			err = t.cc.HandleCrossChainCallbackRequest(*response)
			if err != nil {
				logrus.Errorf("failed to handle cross chain callback request, err:%s", err.Error())
				time.Sleep(2 * time.Second)
				goto retry
			}
		}

		err = t.cc.SaveLatestBlock(block.BlockData)
		if err != nil {
			logrus.Errorf("failed to save latest block, err:%s", err.Error())
			time.Sleep(2 * time.Second)
			continue
		}
	}
}
