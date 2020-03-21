package common

import (
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

func ParseTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}
