package sys

import (
	"os"

	"github.com/gorilla/mux"

	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/api/sys/http"
	"github.com/nettyrnp/exch-rates/api/sys/poller"
	"github.com/nettyrnp/exch-rates/api/sys/repository"
	"github.com/nettyrnp/exch-rates/api/sys/service"
	"github.com/nettyrnp/exch-rates/config"
)

func NewRepository(conf config.Config, kind string) *repository.RDBMSRepository {
	repo := &repository.RDBMSRepository{
		Name: kind,
		Cfg: repository.Config{
			Driver: conf.RepositoryDriver,
			DSN:    conf.RepositoryDSN,
		},
	}

	if initErr := repo.Init(); initErr != nil {
		common.LogError(initErr.Error())
		os.Exit(1)
	}

	return repo
}

func NewPoller(conf config.Config, repo *repository.RDBMSRepository) *poller.RatesPoller {
	a := &poller.RatesPoller{
		Cfg: poller.Config{
			Interval:   conf.PollerInterval,
			Currencies: conf.PollerBaseCurrencies,
			URL:        conf.PollerURL,
			Timeout:    conf.PollerTimeout,
		},
		Repo: repo,
	}
	return a
}

// todo: remove kind
func NewController(conf config.Config, kind string) *http.Controller {
	repo := NewRepository(conf, kind)

	pollr := NewPoller(conf, repo)

	svc := service.New(conf, kind, repo, pollr)

	return http.New(svc, conf, kind)
}

func Route(mux *mux.Router, c *http.Controller) {
	mux.HandleFunc("/exchrates/admin/version", c.Version).Methods("GET")
	mux.HandleFunc("/exchrates/admin/logs", c.Logs).Methods("GET")
	mux.HandleFunc("/exchrates/start_poll", c.StartPolling).Methods("POST")
	mux.HandleFunc("/exchrates/stop_poll", c.StopPolling).Methods("POST")

	mux.HandleFunc("/exchrates/status/{name}", c.Status).Methods("GET", "OPTIONS")
	mux.HandleFunc("/exchrates/history", c.History).Methods("POST", "OPTIONS")
	mux.HandleFunc("/exchrates/momental", c.Momental).Methods("POST", "OPTIONS")
}
