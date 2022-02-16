package adapter

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/bihs/gov"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/zhiqiangxu/bihs"
)

type P2P struct {
	bc    Broadcaster
	chain *core.BlockChain
	gov   *gov.Governance
	ch    chan *bihs.Msg
}

type Broadcaster interface {
	Unicast(target common.Address, msgcode uint64, data interface{})
	Multicast(targets []common.Address, msgcode uint64, data interface{})
}

const chanSize = 20

func NewP2P(bc Broadcaster, chain *core.BlockChain, gov *gov.Governance) *P2P {
	return &P2P{bc: bc, chain: chain, gov: gov, ch: make(chan *bihs.Msg, chanSize)}
}

func (p *P2P) Broadcast(msg *bihs.Msg) {
	payload, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Warn("P2P.Broadcast rlp.EncodeToBytes failed", err)
		return
	}
	validators := p.gov.ValidatorP2PAddrs(msg.Height)
	p.bc.Multicast(validators, ConsensusMsgCode, payload)
}

var ConsensusMsgCode uint64

func (p *P2P) Send(id bihs.ID, msg *bihs.Msg) {
	target := p.gov.ValidatorP2PAddr(common.BytesToAddress(id))
	payload, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Warn("P2P.Send rlp.EncodeToBytes failed", err)
		return
	}
	p.bc.Unicast(target, ConsensusMsgCode, payload)
}

func (p *P2P) MsgCh() <-chan *bihs.Msg {
	return p.ch
}

func (p *P2P) HandleP2pMsg(msg p2p.Msg) (err error) {
	var data []byte
	if err = msg.Decode(&data); err != nil {
		return
	}

	var bihsMsg bihs.Msg
	if err = rlp.DecodeBytes(data, &bihsMsg); err != nil {
		return
	}

	select {
	case p.ch <- &bihsMsg:
	default:
	}

	// TODO help propagation?
	return nil
}
