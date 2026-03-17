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
			txs, err := json.Marshal(data.Txs)
			if err != nil {
				slog.Error("forwardfill: failed to marshal txs", "blockId", id, "err", err)
				return err
			}
			events, err := json.Marshal(data.Events)
			if err != nil {
				slog.Error("forwardfill: failed to marshal events", "blockId", id, "err", err)
				return err
			}

			if err := s.pubsub.Publish(ctx, "block", block); err != nil {
				slog.Error("forwardfill: failed to publish block", "blockId", id, "err", err)
				return err
			}
			if err := s.pubsub.Publish(ctx, "transaction", txs); err != nil {
				slog.Error("forwardfill: failed to publish txs", "blockId", id, "err", err)
				return err
			}
			if err := s.pubsub.Publish(ctx, "event", events); err != nil {
				slog.Error("forwardfill: failed to publish events", "blockId", id, "err", err)
				return err
			}

			slog.Debug("forwardfill: data sent to indexer streams")

			err = s.repo.Create(data.Block, data.Txs, data.Events)
			if err != nil {
				slog.Error("forwardfill: failed to save block", "blockId", data.Block.Id, "err", err)
				return err
			}
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