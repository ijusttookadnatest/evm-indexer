package service

import (
	"context"
	"log/slog"
)

func (s *IndexerService) forwardfill(ctx context.Context) error {
	c := make(chan uint64)
	e := make(chan error, 1)
	go s.fetcher.Subscribe(ctx, c, e)

	for {
		select {
		case id := <-c: {
			data, err := s.fetcher.FetchBlock(id)
			if err != nil {
				slog.Error("forwardfill: failed to fetch block", "blockId", id, "err", err)
				return err
			}
			slog.Info("forwardfill: block fetched", "blockId", id)
			err = s.repo.Create(data.Block, data.Txs, data.Events)
			if err != nil {
				slog.Error("forwardfill: failed to save block", "blockId", data.Block.Id, "err", err)
				return err
			}
		}
		case err := <-e :{
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