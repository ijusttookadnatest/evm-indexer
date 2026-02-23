package ws

import "fmt"

func validateSubscription(sub SubscribeMessage) error {
	if sub.Type != "subscribe" {
		return fmt.Errorf("invalid subscription")
	}
	if sub.Topic != "events" && sub.Topic != "transactions" && sub.Topic != "blocks" {
		return fmt.Errorf("invalid subscription")
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
		if payload.Topic[0] != subscription.Topic0 {
			return false
		}
	}
	return true
}