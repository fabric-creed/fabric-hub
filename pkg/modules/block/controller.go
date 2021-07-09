package block

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/gob"
	"github.com/asdine/storm/v3/q"
	"github.com/fabric-creed/fabric-hub/pkg/database"
	"path/filepath"
	"sync"
)

const DBName = "block.db"

var (
	MT        = database.Block{}
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
		instantDB = db
	})

	return &Controller{db: instantDB}
}

func (c *Controller) FetchLatestBlockNum() (uint64, error) {
	block := &database.Block{}
	err := c.db.Select().Limit(1).Reverse().Find(block)
	if err != nil {
		return 0, err
	}

	return block.BlockNumber, nil
}

func (c *Controller) FetchBlockByBlockHash(blockHash string) (*database.Block, error) {
	var preBlock = &database.Block{}
	err := c.db.Select(q.Eq("BlockHash", blockHash)).Find(preBlock)
	if err != nil {
		return nil, err
	}

	return preBlock, nil
}

func (c *Controller) CreateBlock(block *database.Block) error {
	var preBlock = &database.Block{}
	err := c.db.Select(q.Eq("BlockHash", block.PreviousHash)).Find(preBlock)
	if err != nil && err != storm.ErrNotFound {
		return err
	}

	tx, err := c.db.Begin(true)
	if err = tx.Save(block); err != nil {
		tx.Rollback()
		return err
	}
	if preBlock != nil {
		preBlock.NextHash = block.BlockHash
		if err = tx.Update(preBlock); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
