package repository

import (
	"time"
)

type RatesQueryOpts struct {
	Currency          string
	From              time.Time
	Till              time.Time
	Limit             uint64
	Offset            uint64
	SecondsInInterval uint64
}
