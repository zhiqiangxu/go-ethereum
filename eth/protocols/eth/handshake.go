// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	// handshakeTimeout is the maximum allowed time for the `eth` handshake to
	// complete before dropping the connection.= as malicious.
	handshakeTimeout = 5 * time.Second
)

// Handshake executes the eth protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *Peer) Handshake(network uint64, td *big.Int, head common.Hash, genesis common.Hash, forkID forkid.ID, forkFilter forkid.Filter) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)

	var status StatusPacket // safe to read after two values have been received from errc

	go func() {
		errc <- p2p.Send(p.RW, StatusMsg, &StatusPacket{
			ProtocolVersion: uint32(p.version),
			NetworkID:       network,
			TD:              td,
			Head:            head,
			Genesis:         genesis,
			ForkID:          forkID,
		})
	}()
	go func() {
		errc <- p.readStatus(network, &status, genesis, forkFilter)
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}
	p.td, p.head = status.TD, status.Head

	// TD at mainnet block #7753254 is 76 bits. If it becomes 100 million times
	// larger, it will still fit within 100 bits
	if tdlen := p.td.BitLen(); tdlen > 100 {
		return fmt.Errorf("too large total difficulty: bitlen %d", tdlen)
	}
	return nil
}

func (p *Peer) handshakeOld(network uint64, genesis common.Hash) (ms MinStatus, err error) {
	// statusData is the network packet for the status message.
	type statusData struct {
		ProtocolVersion uint32
		NetworkId       uint64
		TD              *big.Int
		CurrentBlock    common.Hash
		GenesisBlock    common.Hash
	}

	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()

	errc := make(chan error, 1)
	var status statusData
	go func() {
		msg, err := p.RW.ReadMsg()
		if err != nil {
			errc <- err
			return
		}
		if msg.Code != StatusMsg {
			errc <- fmt.Errorf("%w: first msg has code %x (!= %x)", errNoStatusMsg, msg.Code, StatusMsg)
			return
		}

		if msg.Size > maxMessageSize {
			errc <- fmt.Errorf("%w: %v > %v", errMsgTooLarge, msg.Size, maxMessageSize)
			return
		}

		// Decode the handshake and make sure everything matches
		if err := msg.Decode(&status); err != nil {
			errc <- fmt.Errorf("%w: message %v: %v", errDecode, msg, err)
			return
		}
		errc <- nil
	}()

	select {
	case err = <-errc:
		if err != nil {
			return
		}
	case <-timeout.C:
		err = p2p.DiscReadTimeout
		return
	}

	if status.NetworkId != network {
		err = fmt.Errorf("%w: %d (!= %d)", errNetworkIDMismatch, status.NetworkId, network)
		return
	}
	if uint(status.ProtocolVersion) != p.version {
		err = fmt.Errorf("%w: %d (!= %d)", errProtocolVersionMismatch, status.ProtocolVersion, p.version)
		return
	}
	if status.GenesisBlock != genesis {
		err = fmt.Errorf("%w: %x (!= %x)", errGenesisMismatch, status.GenesisBlock, genesis)
		return
	}
	p.td, p.head = status.TD, status.CurrentBlock

	// ensure not chosen
	status.CurrentBlock = genesis
	status.TD = big.NewInt(1)

	go func() {
		errc <- p2p.PSend(p.RW.(p2p.PriorityMsgWriter), StatusMsg, status)
	}()

	select {
	case err = <-errc:
		if err != nil {
			return
		}
	case <-timeout.C:
		err = p2p.DiscReadTimeout
		return
	}

	ms = MinStatus{TD: status.TD, Head: status.CurrentBlock}
	return
}

type MinStatus struct {
	TD   *big.Int
	Head common.Hash
}

func (p *Peer) HandshakeLite(network uint64, genesis common.Hash, upgrade bool) (ms MinStatus, err error) {
	if p.version <= 63 {
		return p.handshakeOld(network, genesis)
	}

	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()

	errc := make(chan error, 1)
	var status StatusPacket
	go func() {
		msg, err := p.RW.ReadMsg()
		if err != nil {
			errc <- err
			return
		}
		if msg.Code != StatusMsg {
			errc <- fmt.Errorf("%w: first msg has code %x (!= %x)", errNoStatusMsg, msg.Code, StatusMsg)
			return
		}

		if msg.Size > maxMessageSize {
			errc <- fmt.Errorf("%w: %v > %v", errMsgTooLarge, msg.Size, maxMessageSize)
			return
		}

		// Decode the handshake and make sure everything matches
		if err := msg.Decode(&status); err != nil {
			errc <- fmt.Errorf("%w: message %v: %v", errDecode, msg, err)
			return
		}
		errc <- nil
	}()

	select {
	case err = <-errc:
		if err != nil {
			return
		}
	case <-timeout.C:
		err = p2p.DiscReadTimeout
		return
	}

	if status.NetworkID != network {
		err = fmt.Errorf("%w: %d (!= %d)", errNetworkIDMismatch, status.NetworkID, network)
		return
	}
	if uint(status.ProtocolVersion) != p.version {
		err = fmt.Errorf("%w: %d (!= %d)", errProtocolVersionMismatch, status.ProtocolVersion, p.version)
		return
	}
	if status.Genesis != genesis {
		err = fmt.Errorf("%w: %x (!= %x)", errGenesisMismatch, status.Genesis, genesis)
		return
	}
	p.td, p.head = status.TD, status.Head

	go func() {
		errc <- p2p.PSend(p.RW.(p2p.PriorityMsgWriter), StatusMsg, status)
	}()

	select {
	case err = <-errc:
		if err != nil {
			return
		}
	case <-timeout.C:
		err = p2p.DiscReadTimeout
		return
	}

	if p.version >= ETH67 && upgrade {
		extension := &UpgradeStatusExtension{
			DisablePeerTxBroadcast: false,
		}

		var extensionRaw *rlp.RawValue
		extensionRaw, err = extension.Encode()
		if err != nil {
			return
		}

		go func() {
			errc <- p2p.PSend(p.RW.(p2p.PriorityMsgWriter), UpgradeStatusMsg, &UpgradeStatusPacket{
				Extension: extensionRaw,
			})
		}()

		var upgradeStatus UpgradeStatusPacket
		go func() {
			msg, err := p.RW.ReadMsg()
			if err != nil {
				errc <- err
				return
			}
			if msg.Code != UpgradeStatusMsg {
				errc <- fmt.Errorf("%w: first msg has code %x (!= %x)", errNoStatusMsg, msg.Code, UpgradeStatusMsg)
				return
			}

			if msg.Size > maxMessageSize {
				errc <- fmt.Errorf("%w: %v > %v", errMsgTooLarge, msg.Size, maxMessageSize)
				return
			}

			// Decode the handshake and make sure everything matches
			if err := msg.Decode(&upgradeStatus); err != nil {
				errc <- fmt.Errorf("%w: message %v: %v", errDecode, msg, err)
				return
			}
			errc <- nil
		}()

		timeout = time.NewTimer(handshakeTimeout)
		defer timeout.Stop()

		for i := 0; i < 2; i++ {
			select {
			case err = <-errc:
				if err != nil {
					return
				}
			case <-timeout.C:
				err = p2p.DiscReadTimeout
				return
			}
		}

		extension, err = upgradeStatus.GetExtension()
		if err != nil {
			return
		}
		p.statusExtension = extension

		if p.statusExtension.DisablePeerTxBroadcast {
			p.CloseTxBroadcast()
		}
	}

	// TD at mainnet block #7753254 is 76 bits. If it becomes 100 million times
	// larger, it will still fit within 100 bits
	if tdlen := p.td.BitLen(); tdlen > 100 {
		err = fmt.Errorf("too large total difficulty: bitlen %d", tdlen)
		return
	}
	ms = MinStatus{TD: status.TD, Head: status.Head}
	return
}

// readStatus reads the remote handshake message.
func (p *Peer) readStatus(network uint64, status *StatusPacket, genesis common.Hash, forkFilter forkid.Filter) error {
	msg, err := p.RW.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != StatusMsg {
		return fmt.Errorf("%w: first msg has code %x (!= %x)", errNoStatusMsg, msg.Code, StatusMsg)
	}
	if msg.Size > maxMessageSize {
		return fmt.Errorf("%w: %v > %v", errMsgTooLarge, msg.Size, maxMessageSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(&status); err != nil {
		return fmt.Errorf("%w: message %v: %v", errDecode, msg, err)
	}
	if status.NetworkID != network {
		return fmt.Errorf("%w: %d (!= %d)", errNetworkIDMismatch, status.NetworkID, network)
	}
	if uint(status.ProtocolVersion) != p.version {
		return fmt.Errorf("%w: %d (!= %d)", errProtocolVersionMismatch, status.ProtocolVersion, p.version)
	}
	if status.Genesis != genesis {
		return fmt.Errorf("%w: %x (!= %x)", errGenesisMismatch, status.Genesis, genesis)
	}
	if err := forkFilter(status.ForkID); err != nil {
		return fmt.Errorf("%w: %v", errForkIDRejected, err)
	}
	return nil
}
