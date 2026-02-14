package domain

var (
	hashLen = 66
	hashLenWithout0x = 64
	addressLen = 42
	topicNumber = 4
)

func ParseHash(hash string) error {
	if  len(hash) != hashLen {
		return ErrInvalidHash
	}
	if hash[0] != '0' && hash[1] != 'x' {
		return ErrInvalidHash
	}
	return nil
}

func ParseAddress(addr string) error {
	if  len(addr) != addressLen {
		return ErrInvalidAddress
	}
	if addr[0] != '0' && addr[1] != 'x' {
		return ErrInvalidAddress
	}
	return nil
}

func ParseBlock(block Block) error {
	if block.Id <= 0 || block.GasLimit <= 0 || block.GasUsed <= 0 || block.Timestamp <= 0 {
		return ErrInvalidBlock
	}
	if err := ParseHash(block.Hash) ; err != nil {
		return ErrInvalidBlock
	}
	if err := ParseHash(block.ParentHash) ; err != nil {
		return ErrInvalidBlock
	}
	if err := ParseAddress(block.Miner) ; err != nil {
		return ErrInvalidBlock
	}
	return nil
}

func ParseTx(tx Transaction) error {
	if tx.BlockId <= 0 || tx.GasUsed <= 0 {
		return ErrInvalidTransaction
	}
	if err := ParseHash(tx.Hash) ; err != nil {
		return ErrInvalidTransaction
	}
	if err := ParseAddress(tx.From) ; err != nil {
		return ErrInvalidTransaction
	}
	if tx.To != nil && *tx.To != "" {
		if err := ParseAddress(*tx.To) ; err != nil {
			return ErrInvalidTransaction
		}
	} 
	return nil
}

func ParseEvent(event Event) error {
	if event.BlockId == 0 {
		return ErrInvalidEvent
	}
	if err := ParseHash(event.TxHash) ; err != nil {
		return ErrInvalidEvent
	}
	if err := ParseAddress(event.Emitter) ; err != nil {
		return ErrInvalidEvent
	}
	if err := ParseTopics(event.Topics) ; err != nil {
		return ErrInvalidEvent
	}
	if (len(event.Datas) - 2) % hashLenWithout0x != 0 {
		return ErrInvalidEvent
	}
	return nil
}

func ParseTopics(topics []string) error {
	for i, topic := range topics {
		if err := ParseHash(topic) ; err != nil {
			return ErrInvalidTopics
		}
		if i >= topicNumber {
			return ErrInvalidTopics
		}
	}
	return nil
}

func ValidateBlockRange(fromBlock, toBlock uint64, rangeMax uint) error {
	if fromBlock == 0 || toBlock == 0 {
		return ErrInvalidBlockRange
	}
	if fromBlock >= toBlock {
		return ErrInvalidBlockRange
	}
	if toBlock - fromBlock > uint64(rangeMax) {
		return ErrInvalidBlockRange	
	}
	return nil
}