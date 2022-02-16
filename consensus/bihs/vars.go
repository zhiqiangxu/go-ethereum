package bihs

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

var (
	errUnknownBlock      = errors.New("unknown block")
	errInvalidCoinbase   = errors.New("invalid coinbase")
	errInvalidNonce      = errors.New("invalid nonce")
	errInvalidDigest     = errors.New("invalid digest")
	errInvalidTime       = errors.New("invalid time")
	errInvalidUncleHash  = errors.New("invalid uncleHash")
	errInvalidDifficulty = errors.New("invalid difficulty")
	errUnclesNotAllowed  = errors.New("uncles not allowed")
)

var (
	defaultUncleHash  = types.CalcUncleHash(nil)
	defaultDifficulty = big.NewInt(0)
	deltaSeconds      = int64(5)
	defaultNonce      = types.BlockNonce{}
)
