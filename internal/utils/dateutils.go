package utils

import (
	"time"
)

const (
	DateFormat = "20060102"
)

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(DateFormat, dateStr)
}

func FormatDate(date time.Time) string {
	return date.Format(DateFormat)
}
