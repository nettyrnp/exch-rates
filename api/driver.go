package api

import (
	"github.com/gorilla/mux"
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/api/middleware"

	"github.com/nettyrnp/exch-rates/config"
)

func RunHTTP(c config.Config) {
	r := mux.NewRouter()

	r.Use(
		mux.CORSMethodMiddleware(r),
		mux.MiddlewareFunc(middleware.DefaultHeaders(c)),
		//mux.MiddlewareFunc(middleware.Debugger()),
		mux.MiddlewareFunc(middleware.RequestID()),
		mux.MiddlewareFunc(middleware.Logger(common.Logger)),
	)

	s := NewServer(r, c)
	api := &API{
		Config: c,
		Router: r,
		Server: s,
	}

	LoadModules(api)

	common.LogInfof("started HTTP server on %s\n", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		common.LogFatalf("starting HTTP server failed with %s", err)
	}
	// todo: graceful shutdown (with a log message)
}
