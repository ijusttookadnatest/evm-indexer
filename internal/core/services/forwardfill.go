package service

import (
	"context"
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

			s.indexerStreams.Block <- data.Block
			s.indexerStreams.Txs <- data.Txs
			s.indexerStreams.Events <- data.Events
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