package adapter

import (
	"sync"

	"github.com/ethereum/go-ethereum/consensus/bihs/gov"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ontio/ontology/common"
	"github.com/zhiqiangxu/bihs"
)

type StateDB struct {
	sync.RWMutex
	chain                  *core.BlockChain
	gov                    *gov.Governance
	prepareEmptyHeaderFunc func() *types.Header
	saveBlockFunc          func(block *types.Block)
	heightSubs             []bihs.HeightChangeSub
}

func NewStateDB(chain *core.BlockChain, gov *gov.Governance, prepareEmptyHeaderFunc func() *types.Header, saveBlockFunc func(block *types.Block)) *StateDB {
	db := &StateDB{
		chain:                  chain,
		gov:                    gov,
		prepareEmptyHeaderFunc: prepareEmptyHeaderFunc,
		saveBlockFunc:          saveBlockFunc,
	}

	return db
}

func (db *StateDB) StoreBlock(blk bihs.Block, commitQC *bihs.QC) error {

	sink := common.NewZeroCopySink(nil)
	commitQC.Serialize(sink)

	block := blk.(*Block)
	header := block.header()
	header.Extra = sink.Bytes()

	db.saveBlockFunc(block.withSeal(header))
	return nil
}

func (db *StateDB) Validate(blk bihs.Block) error {
	return nil
}

func (db *StateDB) EmptyBlock(height uint64) bihs.Block {
	emptyHeader := db.prepareEmptyHeaderFunc()
	return (*Block)(types.NewBlock(emptyHeader, nil, nil, nil, trie.NewStackTrie(nil)))
}

func (db *StateDB) Height() uint64 {
	return db.chain.CurrentHeader().Number.Uint64()
}

func (db *StateDB) SubscribeHeightChange(sub bihs.HeightChangeSub) {
	db.Lock()
	defer db.Unlock()

	db.heightSubs = append(db.heightSubs, sub)
}

func (db *StateDB) HeightChanged() {
	db.RLock()
	heightSubs := db.heightSubs
	db.RUnlock()

	for _, sub := range heightSubs {
		sub.HeightChanged()
	}
}

func (db *StateDB) UnSubscribeHeightChange(sub bihs.HeightChangeSub) {
	db.Lock()
	defer db.Unlock()

	count := len(db.heightSubs)
	for i, subed := range db.heightSubs {
		if subed == sub {
			db.heightSubs[count-1], db.heightSubs[i] = db.heightSubs[i], db.heightSubs[count-1]
			db.heightSubs = db.heightSubs[0 : count-1]
			return
		}
	}
}

func (db *StateDB) ValidatorIndex(height uint64, peer bihs.ID) int {
	return db.gov.ValidatorIndex(height, peer)
}

func (db *StateDB) SelectLeader(height, view uint64) bihs.ID {
	return db.gov.SelectLeader(height, view)
}

func (db *StateDB) ValidatorCount(height uint64) int32 {
	return db.gov.ValidatorCount(height)
}

func (db *StateDB) ValidatorIDs(height uint64) []bihs.ID {
	return db.gov.ValidatorIDs(height)
}

func (db *StateDB) PKs(height uint64, bitmap []byte) interface{} {
	panic("not used")
}
