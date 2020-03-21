package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/api/sys/entity"
	"github.com/nettyrnp/exch-rates/api/sys/repository"
	"os"

	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

type Config struct {
	Interval   time.Duration
	Currencies []string
	URL        string
	Timeout    time.Duration
}

type Poller interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type RatesPoller struct {
	Cfg  Config
	Repo *repository.RDBMSRepository
}

func (a *RatesPoller) Start(ctx context.Context) error {
	ticker := time.NewTicker(a.Cfg.Interval)

	for _, currency := range a.Cfg.Currencies {

		go func(currency string) {
			ctx := context.Background()

			defer ticker.Stop()

			err := a.getPollResponse(ctx, time.Now(), currency)
			if err != nil {
				common.LogError(err.Error())
				return
			}
			for t := range ticker.C {
				err := a.getPollResponse(ctx, t, currency)
				if err != nil {
					common.LogError(err.Error())
					return
				}
			}
		}(currency)
	}

	return nil
}

func (a *RatesPoller) Stop(ctx context.Context) error {
	// todo
	// ...

	common.LogInfof(">> Exiting...")
	os.Exit(0)
	return nil
}

func (a *RatesPoller) getPollResponse(ctx context.Context, t time.Time, currency string) error {
	ctx, cancel := context.WithTimeout(ctx, a.Cfg.Timeout)
	defer cancel()

	res, err := doRequest(ctx, a.Cfg.URL+currency)
	if err != nil {
		return errors.Wrapf(err, "getting poll response")
	}

	return a.Repo.AddExchrate(ctx, res)
}

func doRequest(ctx context.Context, url string) (*entity.Exchrate, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "doing http request")
	}
	defer resp.Body.Close()

	var pollResult entity.PollResult
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "reading http request")
	}
	if err := json.Unmarshal(data, &pollResult); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling http request")
	}
	fmt.Printf(">> pollResult: %v\n", pollResult)

	return toExchrate(pollResult)
}

func toExchrate(p entity.PollResult) (*entity.Exchrate, error) {

	//t, err := toTime(p.Date)
	//if err != nil {
	//	return nil, err
	//}

	t := time.Now().UTC()
	return &entity.Exchrate{
		Time:     t,
		Currency: p.Base,
		Rate:     p.Rates.RUB,
	}, nil
}

// Deprecated, because the site often returns rates for previous days
func toTime(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, errors.New("parsing date")
	}
	now := time.Now().UTC()

	if t.Year() != now.Year() || t.Month() != now.Month() || t.Day() != now.Day() {
		return time.Time{}, errors.New("date is not fresh")
	}
	return now, nil
}
