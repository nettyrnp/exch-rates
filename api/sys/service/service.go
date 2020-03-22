package service

import (
	"context"
	"github.com/nettyrnp/exch-rates/api/sys/entity"
	"github.com/nettyrnp/exch-rates/api/sys/poller"
	"github.com/nettyrnp/exch-rates/api/sys/repository"
	"github.com/nettyrnp/exch-rates/config"
	"github.com/pkg/errors"
	"math"
	"strings"
	"time"
)

const minute = uint64(60)

type Service interface {
	StartPolling()
	StopPolling()

	GetStatus(ctx context.Context, currency string) ([]float64, error)
	GetHistory(ctx context.Context, currency string, from, till time.Time, aggrType string, limit, offset uint64) ([]string, int, error)
	GetMomental(ctx context.Context, currency string, moment time.Time) (float64, error)
}

type RatesService struct {
	Name   string
	Repo   repository.Repository
	Poller poller.Poller
	Conf   config.Config
}

func New(conf config.Config, name string, r repository.Repository, p poller.Poller) *RatesService {
	return &RatesService{
		Name:   name,
		Repo:   r,
		Poller: p,
		Conf:   conf,
	}
}

func (s *RatesService) StartPolling() {
	s.Poller.Start()
}

func (s *RatesService) StopPolling() {
	s.Poller.Stop()
}

func (s *RatesService) GetStatus(ctx context.Context, currency string) ([]float64, error) {
	var res []float64

	lastRate, err := s.GetMomental(ctx, currency, time.Now())
	if err != nil {
		return nil, err
	}
	res = append(res, lastRate)

	spans := []int{1, 7, daysLastMonth()}
	for _, span := range spans {
		now := time.Now()
		numDays := time.Duration(span)
		from := now.Add(-24 * time.Hour * numDays)
		avgRate, err := s.Repo.GetAverage(ctx, currency, from, now)
		if err != nil {
			return nil, err
		}
		res = append(res, avgRate)

	}

	return res, nil
}

func (s *RatesService) GetHistory(ctx context.Context, currency string, from, till time.Time, aggrType string, limit, offset uint64) ([]string, int, error) {
	var seconds uint64
	var timeFormat string
	switch strings.ToLower(aggrType) {
	case entity.Aggr1Min:
		seconds, timeFormat = minute*1, "02-01-2006 15:04"
	case entity.Aggr5Min:
		seconds, timeFormat = minute*5, "02-01-2006 15:04"
	case entity.Aggr1Hour:
		seconds, timeFormat = minute*60, "02-01-2006 15"
	case entity.Aggr1Day:
		seconds, timeFormat = minute*60*24, "02-01-2006"
	default:
		return nil, 0, errors.Errorf("unsupported aggrType '%s'", aggrType)
	}

	opts := repository.RatesQueryOpts{
		Currency:          currency,
		From:              from,
		Till:              till,
		Limit:             limit,
		Offset:            offset,
		SecondsInInterval: seconds,
	}

	averages, total, err := s.Repo.GetHistory(ctx, opts)
	if err != nil {
		return nil, 0, errors.New("getting history")
	}

	return toStrings(timeFormat, averages), total, nil
}

func (s *RatesService) GetMomental(ctx context.Context, currency string, moment time.Time) (float64, error) {
	return s.Repo.GetMomental(ctx, currency, moment)
}

func daysLastMonth() int {
	t := time.Now()
	t2 := t.AddDate(0, -1, 0)
	return -int(math.Round(t2.Sub(t).Hours() / 24))
}

func toStrings(timeFormat string, averages []entity.Average) []string {
	var arr []string
	for _, a := range averages {
		arr = append(arr, a.String(timeFormat))
	}
	return arr
}
