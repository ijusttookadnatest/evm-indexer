package service

import "github/ijusttookadnatest/indexer-evm/internal/core/ports"

type IndexerService struct {
	repo ports.IndexerRepository
	fetcher ports.Fetcher
}

func NewIndexerService(repo ports.IndexerRepository, fetcher ports.Fetcher) *IndexerService {
	return &IndexerService{repo: repo, fetcher:fetcher}
}