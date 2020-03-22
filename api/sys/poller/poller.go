package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/api/sys/entity"
	"github.com/nettyrnp/exch-rates/api/sys/repository"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Currencies []string
	URL        string
	Timeout    time.Duration
}

type Poller interface {
	Start()
	Stop()
}

type RatesPoller struct {
	Cfg     Config
	Repo    *repository.RDBMSRepository
	Ticker  *time.Ticker
	start   time.Time
	Done    chan bool
	ErrorCh chan error
}

func (a *RatesPoller) Start() {
	a.start = time.Now()

	go a.startRequester()

	go a.watchErrors()
}

func (a *RatesPoller) Stop() {
	a.Ticker.Stop()

	common.LogInfof("Stopped poller. Elapsed time: %v", (time.Since(a.start)))

	a.Done <- true
}

func (a *RatesPoller) watchErrors() {
	for {
		select {
		case <-a.Done:
			return
		case err := <-a.ErrorCh:
			log.Println(err)
			a.Stop()
		}
	}
}

func (a *RatesPoller) startRequester() {
	for {
		select {
		case <-a.Done:
			return
		case <-a.Ticker.C:
			for _, currency := range a.Cfg.Currencies {
				if err := a.makeRequest(context.Background(), currency); err != nil {
					common.LogError(err.Error())
					a.ErrorCh <- errors.WithStack(err)
				}
			}
		}
	}
}

func (a *RatesPoller) makeRequest(ctx context.Context, currency string) error {
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
	fmt.Printf("\npoll result: %v\n", pollResult)

	return toExchrate(pollResult)
}

func toExchrate(p entity.PollResult) (*entity.Exchrate, error) {

	// Commented out temporarily, because the web-site often returns rates for previous days
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

// This method may be used in future (see explanation above)
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
