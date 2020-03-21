package api

import (
	"github.com/nettyrnp/exch-rates/api/sys"
	"github.com/nettyrnp/exch-rates/api/sys/entity"
	"github.com/nettyrnp/exch-rates/config"
)

func LoadModules(api *API) {
	api.Router = api.Router.PathPrefix("/api/v0").Subrouter()
	api.NewExchratesModule(api.Config)
}

func (api *API) NewExchratesModule(conf config.Config) {
	c := sys.NewController(conf, string(entity.KindExchratesService))
	sys.Route(api.Router, c)
}
