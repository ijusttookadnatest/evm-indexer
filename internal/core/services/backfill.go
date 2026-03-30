package service

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

func (service *IndexerService) backfill(ctx context.Context, from uint64, targetId uint64, concurrencyF int) error {
	cursor, err := service.repo.GetBackfillCursor()
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
				data, err := service.fetcher.FetchBlock(gCtx, blockId)
				if err != nil {
					slog.Error("backfill: fetch failed", "blockId", blockId, "err", err)
					return err
				}
				service.metrics.DurationFetchingBlock.Observe(time.Since(fetchStart).Seconds())
				results[i] = data
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		writeStart := time.Now()
		if err := service.repo.BulkCreate(results); err != nil {
			slog.Error("backfill: results save failed")
			return err
		}
		service.metrics.DurationWritingBlockDB.Observe(time.Since(writeStart).Seconds())
		service.metrics.DurationProcessingBlock.Observe(time.Since(processStart).Seconds())
		if err := service.repo.UpdateBackfillCursor(end); err != nil {
			return err
		}

		service.metrics.BackfillLastBlockId.Set(float64(end))
		service.metrics.SyncedBlock.Add(float64(len(results)))
		slog.Info("backfill: progress", "curr", end, "targetId", targetId, "remaining", targetId-end)
		
		curr = end + 1
	}

	slog.Info("backfill: completed successfully", "lastIndexed", targetId)
	return nil
}

func (service *IndexerService) Delete(blockId int) error {
	if blockId <= 0 {
		return domain.ErrInvalidId
	}
	if err := service.repo.Delete(blockId); err != nil {
		return err
	}
	return nil
}
