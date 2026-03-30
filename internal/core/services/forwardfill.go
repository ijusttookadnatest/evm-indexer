package service

import (
	"context"
	"encoding/json"
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
			data, err := s.fetcher.FetchBlockPriority(ctx, id)
			if err != nil {
				slog.Error("forwardfill: failed to fetch block", "blockId", id, "err", err)
				return err
			}
			slog.Debug("forwardfill: block fetched", "blockId", id)

			block, err := json.Marshal(data.Block)
			if err != nil {
				slog.Error("forwardfill: failed to marshal block", "blockId", id, "err", err)
				return err
			}
			if err := s.pubsub.Publish(ctx, "block", block); err != nil {
				slog.Error("forwardfill: failed to publish block", "blockId", id, "err", err)
				return err
			}
			for _, tx := range data.Txs {
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
			for _, event := range data.Events {
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

			err = s.repo.Create(data.Block, data.Txs, data.Events)
			if err != nil {
				slog.Error("forwardfill: failed to save block", "blockId", data.Block.Id, "err", err)
				return err
			}
			
			s.metrics.SyncedBlock.Add(1)
			s.metrics.ForwardfillLastBlockId.Set(float64(data.Block.Id))
			slog.Debug("forwardfill: block saved DB")
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