package domain

type IndexerStreams struct {
	Block  chan any
	Txs    chan any
	Events chan any
}
