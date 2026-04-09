package service

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

func (s *IndexerService) backfill(ctx context.Context, from uint64, targetId uint64, concurrencyF int, c chan struct{}) error {
	cursor, err := s.repo.GetBackfillCursor(ctx)
	if err != nil {
		return err
	}

	curr := from
	if cursor > 0 {
		curr = cursor + 1
	}

	if curr > targetId {
		slog.Info("backfill: already up to date")
		return nil
	}

	batchSize := uint64(runtime.NumCPU()*concurrencyF + 1)
	slog.Info("backfill: starting", "from", curr, "to", targetId, "total", targetId-curr+1, "batchSize", batchSize)

	for curr <= targetId {
		end := curr + batchSize - 1
		if end > targetId {
			end = targetId
		}

		size := int(end - curr + 1)
		results := make([]domain.BlockTxsEvents, size)

		slog.Info("backfill: fetching batch", "from", curr, "to", end, "size", size)
		processStart := time.Now()
		g, gCtx := errgroup.WithContext(ctx)
		for i := range size {
			blockId := curr + uint64(i)
			g.Go(func() error {
				fetchStart := time.Now()
				data, err := s.fetcher.FetchBlock(gCtx, blockId)
				if err != nil {
					slog.Error("backfill: fetch failed", "blockId", blockId, "err", err)
					return err
				}
				s.metrics.DurationFetchingBlock.Observe(time.Since(fetchStart).Seconds())
				results[i] = data
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		writeStart := time.Now()
		if err := s.repo.BulkCreate(ctx, results); err != nil {
			slog.Error("backfill: results save failed")
			return err
		}
		s.metrics.DurationWritingBlockDB.Observe(time.Since(writeStart).Seconds())
		s.metrics.DurationProcessingBlock.Observe(time.Since(processStart).Seconds())
		if err := s.repo.UpdateBackfillCursor(ctx, end); err != nil {
			return err
		}

		s.metrics.BackfillLastBlockId.Set(float64(end))
		s.metrics.SyncedBlock.Add(float64(len(results)))
		slog.Info("backfill: progress", "curr", end, "targetId", targetId, "remaining", targetId-end)
		
		curr = end + 1
	}

	slog.Info("backfill: completed successfully", "lastIndexed", targetId)
	c <- struct{}{}
	
	return nil
}