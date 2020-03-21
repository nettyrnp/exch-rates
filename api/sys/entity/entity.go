package entity

import (
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/pkg/errors"
	"time"
)

const (
	Aggr1Min  = "1min"
	Aggr5Min  = "5min"
	Aggr1Hour = "1hour"
	Aggr1Day  = "1day"
)

var ErrInvalidLine = errors.New("invalid line")

type ServiceKind string

const (
	KindExchratesService ServiceKind = "Exchrates"
)

type PollResult struct {
	Rates struct {
		RUB float64 `json:"RUB"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
}

type Average struct {
	Time time.Time
	Rate float64
}

type Exchrate struct {
	ID        int       `json:"-",db:"id"`
	Time      time.Time `json:"time",db:"time"`
	Currency  string    `json:"currency",db:"currency"`
	Rate      float64   `json:"rate",db:"rate"`
	CreatedAt time.Time `json:"createdAt",db:"created_at"`
}

func (c *Exchrate) Validate() error {
	var errs []error
	if c.Time.IsZero() {
		errs = append(errs, errors.New("Time cannot be zero"))
	}
	if c.Currency == "" {
		errs = append(errs, errors.New("Currency cannot be empty"))
	}
	if c.Rate <= 0 {
		errs = append(errs, errors.New("Rate should be positive"))
	}
	if len(errs) > 0 {
		return common.JoinErrors(errs)
	}
	return nil
}
