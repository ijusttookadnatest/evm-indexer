package service

import (
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
)

type EventService struct {
	eventRepo ports.EventRepository
	rangeMax uint
}

func NewEventService(eventRepo ports.EventRepository, rangeMax uint) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		rangeMax: rangeMax,
	}
}

func (service *EventService) GetByEventFilter(filter domain.EventFilter) ([]domain.Event,error) {
	if filter.Emitter == nil && filter.TxHash == nil && filter.FromBlock == nil && len(filter.Topics) == 0 && filter.Limit == nil && filter.ToBlock == nil {
		return nil, domain.ErrEmptyFilter
	}
	if filter.FromBlock != nil || filter.ToBlock != nil {
		if (filter.FromBlock != nil && filter.ToBlock == nil) || (filter.ToBlock != nil && filter.FromBlock == nil) {
			return nil, domain.ErrInvalidBlockRange
		}
		if err := domain.ValidateBlockRange(*filter.FromBlock, *filter.ToBlock, service.rangeMax) ; err != nil {
			return nil, err
		}
	}
	if filter.Limit != nil && *filter.Limit <= 0 {
		return nil, domain.ErrInvalidLimit
	}
	if filter.TxHash != nil {
		if err := domain.ParseHash(*filter.TxHash) ; err != nil {
			return nil, err
		}
	}
	if filter.Emitter != nil {
		if err := domain.ParseAddress(*filter.Emitter) ; err != nil {
			return nil, err
		}
	}
	if len(filter.Topics) > 0 {
		if err := domain.ParseTopics(filter.Topics) ; err != nil {
			return nil, err
		}
	}
	events, err := service.eventRepo.GetByEventFilter(filter)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (service *EventService) GetByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error) {
	if err := domain.ParseHash(hash) ; err != nil {
		return nil, err
	}
	if logIndex < 0 {
		return nil, domain.ErrInvalidId
	}
	event, err := service.eventRepo.GetByTxHashLogIndex(hash, logIndex)
	if err != nil {
		return nil, err
	}
	return event, nil
}