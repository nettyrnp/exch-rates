package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nettyrnp/exch-rates/config"
)

type API struct {
	Config config.Config
	Router *mux.Router
	Server *http.Server
}
