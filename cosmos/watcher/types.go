package watcher

import (
	"time"

	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type Block struct {
	ChainID   string           `json:"chain_id"`
	Height    int64            `json:"height"`
	BlockTime time.Time        `json:"block_time"`
	TxsHash   []bytes.HexBytes `json:"txs_hash"`
	Txs       []*tx.Tx         `json:"txs"`
}
