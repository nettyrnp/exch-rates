package http

type historyReq struct {
	Currency string `json:"currency"`
	From     string `json:"from"`
	To       string `json:"to"`
	AggrType string `json:"aggrType"`
	Limit    uint64 `json:"limit"`
	Offset   uint64 `json:"offset"`
}

type historyResp struct {
	Averages []string `json:"averages"`
	Total    int      `json:"total"`
}

type momentalReq struct {
	Currency string `json:"currency"`
	Time     string `json:"time"`
}

type momentalResp struct {
	Rate float64 `json:"rate"`
}

type statusResp struct {
	MostRecent   float64 `json:"most_recent"`
	DayAverage   float64 `json:"day_average"`
	WeekAverage  float64 `json:"week_average"`
	MonthAverage float64 `json:"month_average"`
}
