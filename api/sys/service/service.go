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
	StartPolling(ctx context.Context) error
	StopPolling(ctx context.Context) error

	GetStatus(ctx context.Context, currency string) ([]float64, error)
	GetHistory(ctx context.Context, currency string, from, till time.Time, aggrType string, limit, offset uint64) ([]entity.Average, int, error)
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

func (s *RatesService) StartPolling(ctx context.Context) error {
	return s.Poller.Start(ctx)
}

func (s *RatesService) StopPolling(ctx context.Context) error {
	return s.Poller.Stop(ctx)
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

func (s *RatesService) GetHistory(ctx context.Context, currency string, from, till time.Time, aggrType string, limit, offset uint64) ([]entity.Average, int, error) {
	var seconds uint64
	switch strings.ToLower(aggrType) {
	case entity.Aggr1Min:
		seconds = minute * 1
	case entity.Aggr5Min:
		seconds = minute * 5
	case entity.Aggr1Hour:
		seconds = minute * 60
	case entity.Aggr1Day:
		seconds = minute * 60 * 24
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
	return s.Repo.GetHistory(ctx, opts)
}

func (s *RatesService) GetMomental(ctx context.Context, currency string, moment time.Time) (float64, error) {
	return s.Repo.GetMomental(ctx, currency, moment)
}

func daysLastMonth() int {
	t := time.Now()
	t2 := t.AddDate(0, -1, 0)
	return -int(math.Round(t2.Sub(t).Hours() / 24))
}
