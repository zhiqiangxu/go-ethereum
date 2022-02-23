package chamber

import (
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// A FullBlock contains the last commit of previous block.
// The last commit field is generally used for proposing a new block.
// However, once the block is signed/sealed by validators,
// verifying a block does not need LastCommit anymore.
// That means a non-validator node does not need LastCommit to sync and fully verify the chain.
type FullBlock struct {
	types.Block
	LastCommit *Commit
}

func (b *FullBlock) HashTo(hash common.Hash) bool {
	if b == nil {
		return false
	}
	return b.Hash() == hash
}

func (b *FullBlock) EncodeRLP(w io.Writer) error {
	err := b.Block.EncodeRLP(w)
	if err != nil {
		return err
	}
	return rlp.Encode(w, b.LastCommit)
}

func (b *FullBlock) DecodeRLP(s *rlp.Stream) error {
	err := b.Block.DecodeRLP(s)
	if err != nil {
		return err
	}
	b.LastCommit = &Commit{}
	return s.Decode(&b.LastCommit)
}
