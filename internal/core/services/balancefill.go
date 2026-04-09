package service

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
	"time"
)

var ERC20 = 20
var ERC721 = 2
var ERC1155_SINGLE = 3
var ERC1155_BATCH = 4

func findStandard(event domain.Log) int {
	if len(event.Topics) == 0 {
		return -1
	}
	switch event.Topics[0] {
	case "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62":
		return ERC1155_SINGLE
	case "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb":
		return ERC1155_BATCH
	}
	if len(event.Topics) <= 3 {
		return ERC20
	}
	return ERC721
}

func extractBalanceEntriesFromLog(event domain.Log) []domain.BalanceEntry {
	std := findStandard(event)

	switch std {
	case ERC20:
		from, to := event.Topics[1], event.Topics[2]
		delta := new(big.Int).SetBytes(common.FromHex(event.Datas))
		return []domain.BalanceEntry{
			{WalletAddress: from, TokenAddress: event.Emitter, Amount: new(big.Int).Neg(delta)},
			{WalletAddress: to, TokenAddress: event.Emitter, Amount: delta},
		}

	case ERC721:
		from, to, tokenId := event.Topics[1], event.Topics[2], event.Topics[3]
		return []domain.BalanceEntry{
			{WalletAddress: from, TokenAddress: event.Emitter, TokenId: tokenId, Amount: big.NewInt(-1)},
			{WalletAddress: to, TokenAddress: event.Emitter, TokenId: tokenId, Amount: big.NewInt(1)},
		}

	case ERC1155_SINGLE:
		from, to := event.Topics[2], event.Topics[3]
		data := common.FromHex(event.Datas)
		if len(data) < 64 {
			return nil
		}
		tokenId := new(big.Int).SetBytes(data[0:32])
		value := new(big.Int).SetBytes(data[32:64])
		return []domain.BalanceEntry{
			{WalletAddress: from, TokenAddress: event.Emitter, TokenId: common.BigToHash(tokenId).Hex(), Amount: new(big.Int).Neg(value)},
			{WalletAddress: to, TokenAddress: event.Emitter, TokenId: common.BigToHash(tokenId).Hex(), Amount: value},
		}

	case ERC1155_BATCH:
		from, to := event.Topics[2], event.Topics[3]
		uint256Arr, _ := abi.NewType("uint256[]", "", nil)
		args := abi.Arguments{{Type: uint256Arr}, {Type: uint256Arr}}
		vals, err := args.Unpack(common.FromHex(event.Datas))
		if err != nil || len(vals) < 2 {
			return nil
		}
		ids := vals[0].([]*big.Int)
		values := vals[1].([]*big.Int)
		entries := make([]domain.BalanceEntry, 0, len(ids)*2)
		for i := range ids {
			if i >= len(values) {
				break
			}
			entries = append(entries,
				domain.BalanceEntry{WalletAddress: from, TokenAddress: event.Emitter, TokenId: common.BigToHash(ids[i]).Hex(), Amount: new(big.Int).Neg(values[i])},
				domain.BalanceEntry{WalletAddress: to, TokenAddress: event.Emitter, TokenId: common.BigToHash(ids[i]).Hex(), Amount: values[i]},
			)
		}
		return entries
	}

	return nil
}

var BlockTimeValidation = 12 * time.Second
var transferSignatures = []string{
	"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // ERC20/ERC721 Transfer
	"0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62", // ERC1155 TransferSingle
	"0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb", // ERC1155 TransferBatch
}

type balanceKey struct {
	wallet string
	token string
	id string
}

func aggregateSameEntries(entries []domain.BalanceEntry) []domain.BalanceEntry {
	m := make(map[balanceKey]*big.Int)
	
	for _, entry := range entries {
		key := balanceKey{
			entry.WalletAddress,
			entry.TokenAddress,
			entry.TokenId,
		}
		if m[key] != nil {
			m[key].Add(m[key], entry.Amount)
			} else {
			m[key] = new(big.Int).Set(entry.Amount)
		}
	}
	
	newEntries := make([]domain.BalanceEntry, len(m))
	i := 0
	for balanceKey, amount := range m {
		newEntries[i] = domain.BalanceEntry{
			WalletAddress: balanceKey.wallet,
			TokenAddress: balanceKey.token,
			TokenId: balanceKey.id,
			Amount: amount,
		}
		i++
	}
	return newEntries
}

func (s *IndexerService) balancefill(ctx context.Context, batchSize uint64, lagFinalized int) error {
	cursor, err := s.repo.GetBalancefillCursor(ctx)
	if err != nil {
		return err
	}
	
	for {
		select {
		case <-ctx.Done(): {
			slog.Error("balancefill: context CANCELLED", "Reason", ctx.Err())
			return ctx.Err()
		}
		default: {
			var fullBalanceEntries []domain.BalanceEntry
			var balanceEntries []domain.BalanceEntry
	
			maxBlock, err := s.repo.GetMaxIndexedBlock(ctx)
			if err != nil {
				return err
			}
			if cursor >= maxBlock - uint64(lagFinalized) {
				select {                                                                         
				case <-time.After(BlockTimeValidation):
				case <-ctx.Done(): {
					slog.Error("balancefill: context CANCELLED", "Reason", ctx.Err())                                                               
					return ctx.Err()
				}
				}
				continue
			}
			
			events, err := s.repo.GetLogsByTopic(ctx, domain.LogFilter{
				Topics: transferSignatures,
				From:   cursor,
				Limit:  batchSize,
			})
			if err != nil {
				slog.Error("balancefill: failed to read batch logs", "cursor", cursor, "err", err)
				return err
			}
			slog.Debug("balancefill: logs fetched", "cursor", cursor)
			if len(events) == 0 {
				continue
			}
	
			for _, event := range events {
				fullBalanceEntries = append(fullBalanceEntries, extractBalanceEntriesFromLog(event)...)
			}
			balanceEntries = aggregateSameEntries(fullBalanceEntries)
			err = s.repo.BatchUpsertBalance(ctx, balanceEntries)
			if err != nil {
				slog.Error("balancefill: failed to batch upsert balance update", "cursor", cursor, "err", err)
				return err
			}
	
			cursor = events[len(events) - 1].Id
			err = s.repo.UpdateBalancefillCursor(ctx, cursor)
			if err != nil {
				slog.Error("balancefill: failed to update balancefill cursor", "cursor", cursor, "err", err)
				return err
			}
		}
		}
	}
}