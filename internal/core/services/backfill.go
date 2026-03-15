package service

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"log/slog"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

func worker(ctx context.Context, chanJobs <-chan uint64, chanResults chan domain.BlockTxsEvents, backfiller ports.Fetcher) error {
	for {
		select {
		case <-ctx.Done():
			{
				slog.Error("worker: context cancelled", "reason", ctx.Err())
				return ctx.Err()
			}
		case id, ok := <-chanJobs:
			{
				if !ok {
					slog.Error("worker: jobs channel closed, exiting")
					return nil
				}
				blockData, err := backfiller.FetchBlock(ctx, id)
				if err != nil {
					slog.Error("worker: failed to fetch block", "blockId", id, "err", err)
					return err
				}
				slog.Info("worker: block fetched", "blockId", id)
				select {
				case chanResults <- blockData:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

func loader(ctx context.Context, chanLastIndex chan uint64, chanResults <-chan domain.BlockTxsEvents, repo ports.IndexerRepository, numJobs int, curr uint64) error {
	m := make(map[uint64]domain.BlockTxsEvents, numJobs)
	defer close(chanLastIndex)
	// defer clean map
	for {
		select {
		case <-ctx.Done():
			{
				slog.Error("loader: context cancelled", "reason", ctx.Err())
				return ctx.Err()
			}
		case result, ok := <-chanResults:
			{
				if !ok { return nil }
				if result.Block.Id != curr { slog.Info("loader: block buffered (out-of-order)", "blockId", result.Block.Id, "waitingFor", curr, "buffered", len(m)+1) }
				m[result.Block.Id] = result
				
				prevCurr := curr
				for {
					data, ok := m[curr]
					if !ok { break }
					slog.Info("loader: saving block", "blockId", data.Block.Id, "txs", len(data.Txs), "events", len(data.Events))
					err := repo.Create(data.Block, data.Txs, data.Events)
					if err != nil {
						slog.Error("loader: failed to save block", "blockId", data.Block.Id, "err", err)
						return err
					}
					delete(m, curr)
					curr++
				}
				if curr > prevCurr {
					if err := repo.UpdateBackfillCursor(curr - 1) ; err != nil {
						return err
					}
					select {
					case chanLastIndex <- curr - 1:
					case <-ctx.Done():{
						return ctx.Err()
					}
					default: {                                                                            
						<-chanLastIndex
						chanLastIndex <- curr - 1
					}
				}
				}
			}
		}
	}
}

func (service *IndexerService) backfill(ctx context.Context, from uint64, targetId uint64, concurrencyF int) error {
	var err error
	var ok bool
	var curr uint64
	var i uint64

	cursor, err := service.repo.GetBackfillCursor()
	if err != nil {
		return err
	}

	if cursor == 0 {
		curr = from
	} else {
		curr = cursor + 1
	}

	if curr > targetId {
		slog.Info("backfill: already up to date")
		return nil
	}

	numWorkers := runtime.NumCPU()*concurrencyF + 1
	numJobs := numWorkers

	gWorkers, ctxWorkers := errgroup.WithContext(ctx)
	gLoader, ctxLoader := errgroup.WithContext(ctx)

	chanResults := make(chan domain.BlockTxsEvents, numJobs)
	chanJobs := make(chan uint64, numJobs)
	chanLastIndex := make(chan uint64, 1)

	gLoader.Go(func() error {
		return loader(ctxLoader, chanLastIndex, chanResults, service.repo, numJobs, curr)
	})

	for range numWorkers {
		gWorkers.Go(func() error {
			return worker(ctxWorkers, chanJobs, chanResults, service.fetcher)
		})
	}

	sent := curr

curr:
	for curr <= targetId {
		loop:
		for i = sent; i < sent+uint64(numJobs) && i <= targetId; i++ {
			time.Sleep(1 * time.Millisecond)
			select {
			case chanJobs <- i:
			case <-ctxWorkers.Done():
          		break loop
			case <-ctx.Done():
				break loop
      }
		}
		sent = i

		select {
		case curr, ok = <-chanLastIndex:
			{
				if !ok {
					slog.Error("backfill: chanLastIndex closed unexpectedly")
					break curr
				}
				if curr == targetId {
					break curr
				}
				slog.Info("backfill: progress", "curr", curr, "targetId", targetId, "remaining", targetId-curr)
			}
		case <-ctxWorkers.Done():
			{
				slog.Error("backfill: error in workers, break loop")
				break curr
			}
		case <-ctxLoader.Done():
			{
				slog.Error("backfill: error in loader, break loop")
				break curr
			}
		case <-ctx.Done():
			{
				slog.Error("backfill: context cancelled", "reason", ctx.Err())
				break curr
			}
		}
	}

	slog.Info("backfill: waiting for workers and loader to finish")
	close(chanJobs)
	workerErr := gWorkers.Wait()

	close(chanResults)
	loaderErr := gLoader.Wait()

	if workerErr != nil {
		slog.Error("backfill: worker error", "err", workerErr)
		return workerErr
	}
	if loaderErr != nil {
		slog.Error("backfill: loader error", "err", loaderErr)
		return loaderErr
	}
	if ctx.Err() != nil {
		slog.Error("backfill: context cancelled", "err", ctx.Err())
		return ctx.Err()
	}
	slog.Info("backfill: completed successfully", "lastIndexed", targetId)
	return nil
}

// func (service *IndexerService) Create(block *domain.Block, txs []domain.Transaction, events []domain.Event) error {
// 	if err := domain.ParseBlock(*block); err != nil {
// 		return err
// 	}
// 	for _, tx := range txs {
// 		if err := domain.ParseTx(tx); err != nil {
// 			return err
// 		}
// 	}
// 	for _, event := range events {
// 		if err := domain.ParseEvent(event); err != nil {
// 			return err
// 		}
// 	}
// 	err := service.repo.Create(*block, txs, events)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (service *IndexerService) Delete(blockId int) error {
	if blockId <= 0 {
		return domain.ErrInvalidId
	}
	if err := service.repo.Delete(blockId); err != nil {
		return err
	}
	return nil
}
