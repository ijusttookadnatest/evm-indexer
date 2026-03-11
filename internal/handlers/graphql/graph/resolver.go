package graph

import "github/ijusttookadnatest/evm-indexer/internal/core/ports"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Service ports.QueryService
}
