package client

import (
	"github.com/ethereum/go-ethereum"
)

type Sync interface {
	ethereum.LogFilterer
	ethereum.ChainReader
}
