package adapter

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ontio/ontology/common"
	"github.com/zhiqiangxu/bihs"
)

type Block types.Block

func (b *Block) Default() bihs.Block {
	return &Block{}
}

func (b *Block) Height() uint64 {
	return (*types.Block)(b).Header().Number.Uint64()
}

func (b *Block) Hash() bihs.Hash {
	hash := (*types.Block)(b).Header().Hash()
	return hash[:]
}

func (b *Block) Empty() bool {
	return len((*types.Block)(b).Transactions()) == 0
}

func (b *Block) Serialize(sink *common.ZeroCopySink) {
	// (*types.Block)(b).EncodeRLP()
	bytes, err := rlp.EncodeToBytes(b)
	if err != nil {
		panic(fmt.Sprintf("rlp.EncodeToBytes failed:%v", err))
	}
	sink.WriteBytes(bytes)
}

func (b *Block) Deserialize(source *common.ZeroCopySource) error {
	bytes, err := source.ReadVarBytes()
	if err != nil {
		return fmt.Errorf("source.ReadVarBytes failed:%v", err)
	}
	return rlp.DecodeBytes(bytes, b)
}
