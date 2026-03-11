package ws

import (
	"encoding/json"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func marshalWSMessage(msgType string, payload any) ([]byte, error) {
	return json.Marshal(WSMessage{Type: msgType, Payload: payload})
}

func validateSubscription(sub SubscribeMessage) error {
	if sub.Type != "subscribe" {
		return domain.ErrInvalidSubscription
	}
	if sub.Topic != "events" && sub.Topic != "transactions" && sub.Topic != "blocks" {
		return domain.ErrInvalidSubscription
	}
	return nil
}

func extractFilter(sub SubscribeMessage) SubscriptionFilter {
	var address, topic0 string
	if sub.Address != "" {
		address = sub.Address
	}
	if sub.Topic0 != "" {
		topic0 = sub.Topic0
	}
	return SubscriptionFilter{
		Address: address,
		Topic0: topic0,
	}
}

func matchesFilter(subscription SubscriptionFilter, payload PayloadFilter) bool {
	if subscription.Address != "" {
		if payload.To != subscription.Address && payload.From != subscription.Address && payload.Emitter != subscription.Address {
			return false
		}
	}
	if subscription.Topic0 != "" {
		if len(payload.Topic) == 0 || payload.Topic[0] != subscription.Topic0 {
			return false
		}
	}
	return true
}