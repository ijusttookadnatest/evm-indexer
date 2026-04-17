package domain

import "math/big"

type Reorg struct {
	BlockId uint64 `json:"block_id"`
}

type Event struct {
	BlockId  uint64
	TxHash   string
	LogIndex uint64
	Emitter  string
	Datas    string
	Topics   []string
}

type Log struct {
	Id      uint64
	BlockId uint64
	Emitter string
	Datas   string
	Topics  []string
}

type BlockTxsEvents struct {
	Block  Block
	Txs    []Transaction
	Events []Event
}

type Block struct {
	Hash       string
	Id         uint64
	ParentHash string
	GasLimit   uint64
	GasUsed    uint64
	Miner      string
	Timestamp  uint64
}

type BlockTxs struct {
	Block Block
	Txs   []Transaction
}

type Transaction struct {
	BlockId uint64
	Hash    string
	From    string
	To      *string
	GasUsed uint64
	Status  uint64
}

type TransactionFilter struct {
	BlockId   *uint64
	Hash      *string
	From      *string
	To        *string
	FromBlock *uint64
	ToBlock   *uint64
	Limit     *int
}

type EventFilter struct {
	TxHash    *string
	Emitter   *string
	Topics    []string
	FromBlock *uint64
	ToBlock   *uint64
	Limit     *int
}

type LogFilter struct {
	Topics    []string
	FromBlock uint64
	ToBlock   uint64
}

// type WalletBalance struct {
//     WalletAddress string
//     TokenAddress  string
//     TokenId       string // "" for ERC20, tokenId hex for ERC721/1155
//     Amount        *big.Int
// }

// type BalanceUpdate struct {
//     From string
//     To string
//     TokenAddress string
//     TokenId string
//     Delta *big.Int
// }

type BalanceEntry struct {
	WalletAddress string
	TokenAddress  string
	TokenId       string
	Amount        *big.Int
}
