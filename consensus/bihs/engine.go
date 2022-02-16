package bihs

import (
	"context"
	"fmt"
	"math/big"
	"path"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/bihs/adapter"
	"github.com/ethereum/go-ethereum/consensus/bihs/gov"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core"
	ethcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/zhiqiangxu/bihs"
	"github.com/zhiqiangxu/util"
)

type BiHS struct {
	sync.Once
	wg         sync.WaitGroup
	nodeConfig *node.Config
	bihsConfig *params.BiHSConfig
	signer     *adapter.Signer
	db         ethdb.Database
	core       *bihs.HotStuff
	p2p        *adapter.P2P

	chainHeadCh  chan core.ChainHeadEvent
	chainHeadSub event.Subscription
}

func New(bihsConfig *params.BiHSConfig, nodeConfig *node.Config, db ethdb.Database) *BiHS {
	privateKey := nodeConfig.NodeKey()
	signer := adapter.NewSigner(privateKey)

	return &BiHS{
		nodeConfig:  nodeConfig,
		bihsConfig:  bihsConfig,
		signer:      signer,
		db:          db,
		chainHeadCh: make(chan core.ChainHeadEvent, 1),
	}
}

func (bh *BiHS) Init(chain *ethcore.BlockChain, bc adapter.Broadcaster, consensusMsgCode int) {

	adapter.ConsensusMsgCode = uint64(consensusMsgCode)

	dir := path.Join(bh.nodeConfig.DataDir, "bihs")
	proposer := bh.signer.Address()
	bihsConfig := bh.bihsConfig

	conf := bihs.Config{
		BlockInterval: time.Duration(bihsConfig.Period) * time.Second,
		DataDir:       dir,
		ProposerID:    proposer[:],
		EcSigner:      bh.signer,
	}

	governance := gov.New(chain)

	store := adapter.NewStateDB(chain, governance)
	p2p := adapter.NewP2P(bc, chain, governance)

	core := bihs.New(store, p2p, conf)
	bh.core = core
	bh.p2p = p2p

	bh.chainHeadSub = chain.SubscribeChainHeadEvent(bh.chainHeadCh)
	util.GoFunc(&bh.wg, func() {
		select {
		case <-bh.chainHeadCh:
			store.HeightChanged()
		case <-bh.chainHeadSub.Err():
		}
	})
}

func (bh *BiHS) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

func (bh *BiHS) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return bh.verifyHeader(chain, header, nil)
}

func (bh *BiHS) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := bh.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

func (bh *BiHS) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errUnclesNotAllowed
	}
	return nil
}

func (bh *BiHS) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}

	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}

	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if header.Time <= parent.Time {
		return errInvalidTime
	}

	if header.Time > uint64(time.Now().Unix()+deltaSeconds) {
		return consensus.ErrFutureBlock
	}

	if header.Coinbase == (common.Address{}) {
		return errInvalidCoinbase
	}

	if header.Nonce.Uint64() != 0 {
		return errInvalidNonce
	}

	if header.UncleHash != defaultUncleHash {
		return errInvalidUncleHash
	}

	if header.MixDigest != types.BiHSDigest {
		return errInvalidDigest
	}

	if header.Difficulty.Cmp(defaultDifficulty) != 0 {
		return errInvalidDifficulty
	}

	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}

	if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
		return err
	}

	// All basic checks passed, verify signatures fields
	return nil
}

func (bh *BiHS) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	header.Coinbase = bh.signer.Address()
	header.Nonce = defaultNonce
	header.MixDigest = types.BiHSDigest

	parent, err := bh.getParentHeader(chain, header)
	if err != nil {
		return err
	}

	header.Difficulty = defaultDifficulty
	header.Time = parent.Time + bh.bihsConfig.Period
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}

	return nil
}

func (bh *BiHS) getParentHeader(chain consensus.ChainHeaderReader, header *types.Header) (*types.Header, error) {
	number := header.Number.Uint64()
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return nil, consensus.ErrUnknownAncestor
	}
	return parent, nil
}

func (bh *BiHS) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header) {
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = defaultUncleHash
}

func (bh *BiHS) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = defaultUncleHash

	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

func (bh *BiHS) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) (err error) {
	bh.Do(func() {
		bh.core.Start()
	})

	bh.core.Propose(context.Background(), (*adapter.Block)(block))

	return
}

func (bh *BiHS) SealHash(header *types.Header) common.Hash {
	return header.Hash()
}

func (bh *BiHS) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return defaultDifficulty
}

func (bh *BiHS) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "bihs",
		Version:   "1.0",
		Service:   &API{chain: chain, bihs: bh},
		Public:    true,
	}}
}

func (bh *BiHS) Close() error {
	if bh.chainHeadSub != nil {
		bh.chainHeadSub.Unsubscribe()
	}

	return bh.core.Stop()
}

func (bh *BiHS) HandleP2pMsg(msg p2p.Msg) error {
	return bh.p2p.HandleP2pMsg(msg)
}
