package gov

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/zhiqiangxu/bihs"
)

type Governance struct {
	chain *core.BlockChain
}

func New(chain *core.BlockChain) *Governance {
	return &Governance{chain: chain}
}

func (g *Governance) ValidatorP2PAddrs(height uint64) []common.Address {
	return nil
}

func (g *Governance) ValidatorP2PAddr(account common.Address) common.Address {
	return account
}

func (g *Governance) ValidatorIndex(height uint64, peer bihs.ID) int {
	return 0
}

func (g *Governance) SelectLeader(height, view uint64) bihs.ID {
	return nil
}

func (g *Governance) ValidatorCount(height uint64) int32 {
	return 0
}
func (g *Governance) ValidatorIDs(height uint64) []bihs.ID {
	return nil
}
