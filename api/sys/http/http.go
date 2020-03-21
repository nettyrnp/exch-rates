package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nettyrnp/exch-rates/api/common"
	"github.com/nettyrnp/exch-rates/api/sys/dto"
	"github.com/nettyrnp/exch-rates/api/sys/service"
	"github.com/nettyrnp/exch-rates/config"
	"github.com/pkg/errors"
	"net/http"
)

type Controller struct {
	Kind    string
	Service service.Service
	Conf    config.Config
}

func New(s service.Service, conf config.Config, kind string) *Controller {
	return &Controller{
		Kind:    kind,
		Service: s,
		Conf:    conf,
	}
}

func (c *Controller) Version(w http.ResponseWriter, r *http.Request) {
	v := "0.0.1-1"
	w.Write([]byte(fmt.Sprintf("%s Service, version %s", c.Kind, v)))
}

func (c *Controller) Logs(w http.ResponseWriter, r *http.Request) {
	if c.Conf.AppEnv != config.AppEnvDev {
		common.LogError("Attempt to access logs in non-development mode")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	svcResp := dto.NewServiceResponse()
	log, err := common.GetLog(c.Conf)
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, err.Error())
		return
	}
	max := 5000
	if len(log) > max {
		log = "     <........... truncated ............>     \n" + log[len(log)-max:]
	}
	w.Write([]byte("Backend latest log: \n" + log))
}

func (c *Controller) StartPolling(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()

	if err := c.Service.StartPolling(r.Context()); err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "polling exchange rates data from %v", c.Conf.PollerURL).Error())
		return
	}

	common.LogInfof("Started polling")
	respondOK(w, svcResp, "Started polling")
}

func (c *Controller) StopPolling(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()

	if err := c.Service.StopPolling(r.Context()); err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "stopping polling").Error())
		return
	}

	common.LogInfof("Stopped polling")
	respondOK(w, svcResp, "Stopped polling")
}

func (c *Controller) Status(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()
	currencyName := mux.Vars(r)["name"]

	rates, err := c.Service.GetStatus(r.Context(), currencyName)
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "getting status for currency '%v'", currencyName).Error())
		return
	}

	svcResp.Body = &statusResp{
		MostRecent:   rates[0],
		DayAverage:   rates[1],
		WeekAverage:  rates[2],
		MonthAverage: rates[3],
	}
	respondOK(w, svcResp, "")
}

func (c *Controller) History(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()

	var req historyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errors.Wrap(err, "parsing request body").Error())
		return
	}

	from, err := common.ParseTime(req.From)
	if err != nil {
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errors.Wrapf(err, "parsing param From '%v'", req.From).Error())
		return
	}
	till, err := common.ParseTime(req.To)
	if err != nil {
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errors.Wrapf(err, "parsing param To '%v'", req.To).Error())
		return
	}

	averages, total, err := c.Service.GetHistory(r.Context(), req.Currency, from, till, req.AggrType, req.Limit, req.Offset)
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "finding averages").Error())
		return
	}

	svcResp.Body = &historyResp{
		Averages: averages,
		Total:    total,
	}
	respondOK(w, svcResp, "")
}

func (c *Controller) Momental(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()

	var req momentalReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errors.Wrap(err, "parsing request body").Error())
		return
	}

	moment, err := common.ParseTime(req.Time)
	if err != nil {
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errors.Wrapf(err, "parsing %v param", req.Time).Error())
		return
	}

	rate, err := c.Service.GetMomental(r.Context(), req.Currency, moment)
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "finding exchange rate for moment %v", moment).Error())
		return
	}

	svcResp.Body = &momentalResp{
		Rate: rate,
	}
	respondOK(w, svcResp, "")
}

func (c *Controller) respondNotOK(w http.ResponseWriter, statusCode int, response *dto.ServiceResponse, errorMsg string) {
	if c.Conf.AppEnv == config.AppEnvDev {
		respondNotOKWithError(w, statusCode, response, errorMsg)
		return
	}
	common.LogError(errorMsg)
	respond(w, statusCode, response, "")
}

func respondNotOKWithError(w http.ResponseWriter, statusCode int, response *dto.ServiceResponse, errorMsg string) {
	common.LogError(errorMsg)
	respond(w, statusCode, response, errorMsg)
}

func respondOK(w http.ResponseWriter, response *dto.ServiceResponse, msg string) {
	if msg != "" {
		common.LogInfo(msg)
	}
	statusCode := http.StatusOK
	respond(w, statusCode, response, msg)
}

func respond(w http.ResponseWriter, statusCode int, response *dto.ServiceResponse, msg string) {
	response.Status.Code = statusCode
	response.Status.Text = msg
	jsonResponse, _ := json.Marshal(*response)
	w.WriteHeader(statusCode)
	w.Write(jsonResponse)
}
