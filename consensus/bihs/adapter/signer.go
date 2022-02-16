package adapter

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zhiqiangxu/bihs"
)

type Signer struct {
	address    common.Address
	privateKey *ecdsa.PrivateKey
}

func NewSigner(privateKey *ecdsa.PrivateKey) *Signer {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &Signer{
		address:    address,
		privateKey: privateKey,
	}
}

func (s *Signer) Address() common.Address {
	return s.address
}

func (s *Signer) Sign(data []byte) (sig []byte) {
	hashData := crypto.Keccak256(data)
	sig, err := crypto.Sign(hashData, s.privateKey)
	if err != nil {
		panic(fmt.Sprintf("crypto.Sign failed:%v", err))
	}
	return
}

func (s *Signer) Verify(data []byte, sig []byte) (id bihs.ID, err error) {
	pubkey, err := crypto.Ecrecover(data, sig)
	if err != nil {
		return
	}

	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
	id = signer[:]
	return
}

func (s *Signer) TCombine(_ []byte, sigs [][]byte) []byte {
	var aggregatedSigs []byte
	for _, sig := range sigs {
		aggregatedSigs = append(aggregatedSigs, sig...)
	}
	return aggregatedSigs
}

func (s *Signer) TVerify(data []byte, sigs []byte, ids []bihs.ID, quorum int32) bool {
	if len(sigs)%crypto.SignatureLength != 0 {
		return false
	}

	addrMap := make(map[common.Address]bool)
	for _, id := range ids {
		addr := common.BytesToAddress(id)
		if addrMap[addr] {
			return false
		}
		addrMap[addr] = true
	}
	sigCount := int32(0)
	for len(sigs) >= crypto.SignatureLength {
		signer, err := s.Verify(data, sigs[0:crypto.SignatureLength])
		if err != nil {
			return false
		}
		addr := common.BytesToAddress(signer)
		if !addrMap[addr] {
			return false
		}
		delete(addrMap, addr)
		sigCount++
	}

	return sigCount >= quorum
}
