package domain

import (
	"errors"
)

var ErrNotFound = errors.New("row not found")
var ErrEmptyFilter = errors.New("empty filter")

var ErrInvalidBlock = errors.New("invalid block")
var ErrInvalidTransaction = errors.New("invalid transaction")
var ErrInvalidEvent = errors.New("invalid event")
var ErrInvalidAddress = errors.New("invalid address")
var ErrInvalidHash = errors.New("invalid hash")
var ErrInvalidOffset = errors.New("invalid offset")
var ErrInvalidId = errors.New("invalid id")
var ErrInvalidBlockRange = errors.New("invalid block range")
var ErrInvalidTimeRange = errors.New("invalid time range")
var ErrInvalidLimit = errors.New("invalid limit")
var ErrInvalidTopics = errors.New("invalid topics")

var ErrInvalidSubscription = errors.New("invalid subscription")
