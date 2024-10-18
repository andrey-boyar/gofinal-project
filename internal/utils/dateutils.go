package utils

import (
	"time"
)

const (
	DateFormat = "20060102"
	// DefaultTaskLimit определяет максимальное количество задач, возвращаемых при поиске
	DefaultTaskLimit = 50
	DateFormatDB     = "2006-01-02"
)

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(DateFormat, dateStr)
}

func FormatDate(date time.Time) string {
	return date.Format(DateFormat)
}

func FormatDateDB(date time.Time) string {
	return date.Format(DateFormatDB)
}

func ParseDateDB(dateStr string) (time.Time, error) {
	return time.Parse(DateFormatDB, dateStr)
}
