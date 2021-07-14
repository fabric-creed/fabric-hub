package transaction

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/gob"
	"github.com/asdine/storm/v3/q"
	"github.com/fabric-creed/fabric-hub/pkg/database"
	"path/filepath"
	"sync"
)

const DBName = "transaction.db"

var (
	MT        = database.Transaction{}
	instantDB *storm.DB
	once      sync.Once
)

type Controller struct {
	db *storm.DB
}

func NewController(dbPath string) *Controller {
	once.Do(func() {
		db, err := storm.Open(filepath.Join(dbPath, DBName), storm.Codec(gob.Codec))
		if err != nil {
			panic(err)
		}
		db.Init(new(database.Transaction))
		instantDB = db
	})

	return &Controller{db: instantDB}
}

func (c *Controller) FetchTransactionByTransactionHash(transactionHash string) (*database.Transaction, error) {
	var transaction []database.Transaction
	err := c.db.Select(q.Eq("Transaction", transactionHash)).Limit(1).Find(&transaction)
	if err != nil {
		return nil, err
	}

	return &transaction[0], nil
}

func (c *Controller) Create(blockNumber uint64, blockHash string, txHash string, originInfo []byte) error {
	transaction := &database.Transaction{
		BlockNumber:     blockNumber,
		BlockHash:       blockHash,
		TransactionHash: txHash,
		OriginInfo:      originInfo,
	}
	return c.db.Save(transaction)
}
