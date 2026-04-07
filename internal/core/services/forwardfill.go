package service

import (
	"context"
	"encoding/json"
	"errors"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
)

func (s *IndexerService) forwardfill(ctx context.Context) error {
	c := make(chan uint64)
	errChan := make(chan error, 1)
	
	go func() {
		errChan <- s.fetcher.Subscribe(ctx, c)
	}()

	for {
		select {
		case id := <-c: {
			from, err := s.fetcher.FetchBlockPriority(ctx, id)
			if err != nil {
				slog.Error("forwardfill: failed to fetch block", "blockId", id, "err", err)
				return err
			}
			slog.Debug("forwardfill: block fetched", "blockId", id)
			
			curr, err := s.repo.GetBlockById(ctx, from.Block.Id)
			if err != nil && !errors.Is(domain.ErrNotFound, err) {
				slog.Error("forwardfill: failed to get block id to check reorg", "blockId", id, "err", err)
				return err
			}
			if curr != nil && (curr.Hash != from.Block.Hash) {
				slog.Info("forwardfill: reorg detected", "blockId", from.Block.Id)
				if err := s.repo.Delete(from.Block.Id); err != nil {
					slog.Error("forwardfill: failed to delete block during reorg", "blockId", from.Block.Id, "err", err)
					return err
				}
				reorg, err := json.Marshal(domain.Reorg{BlockId: from.Block.Id})
				if err != nil {
					slog.Error("forwardfill: failed to marshal reorg", "blockId", from.Block.Id, "err", err)
					return err
				}
				if err := s.pubsub.Publish(ctx, "reorg", reorg); err != nil {
					slog.Error("forwardfill: failed to publish reorg", "blockId", from.Block.Id, "err", err)
					return err
				}
			}
			err = s.repo.Create(from.Block, from.Txs, from.Events)
			if err != nil {
				slog.Error("forwardfill: failed to save block", "blockId", from.Block.Id, "err", err)
				return err
			}
			
			s.metrics.SyncedBlock.Add(1)
			s.metrics.ForwardfillLastBlockId.Set(float64(from.Block.Id))
			slog.Debug("forwardfill: block saved DB")

			block, err := json.Marshal(from.Block)
			if err != nil {
				slog.Error("forwardfill: failed to marshal block", "blockId", id, "err", err)
				return err
			}
			if err := s.pubsub.Publish(ctx, "block", block); err != nil {
				slog.Error("forwardfill: failed to publish block", "blockId", id, "err", err)
				return err
			}
			for _, tx := range from.Txs {
				b, err := json.Marshal(tx)
				if err != nil {
					slog.Error("forwardfill: failed to marshal tx", "blockId", id, "err", err)
					return err
				}
				if err := s.pubsub.Publish(ctx, "transaction", b); err != nil {
					slog.Error("forwardfill: failed to publish tx", "blockId", id, "err", err)
					return err
				}
			}
			for _, event := range from.Events {
				b, err := json.Marshal(event)
				if err != nil {
					slog.Error("forwardfill: failed to marshal event", "blockId", id, "err", err)
					return err
				}
				if err := s.pubsub.Publish(ctx, "event", b); err != nil {
					slog.Error("forwardfill: failed to publish event", "blockId", id, "err", err)
					return err
				}
			}
			slog.Debug("forwardfill: data sent to indexer streams")
		}
		case err := <-errChan :{
			slog.Error("forwardfill: error received from subscribe", "err", err)
			return err
		}
		case <-ctx.Done(): {
			slog.Error("forwardfill: cancel from context with error", "err", ctx.Err())
			return ctx.Err()
		}
		}
	}
}