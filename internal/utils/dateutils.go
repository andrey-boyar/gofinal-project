package utils

import (
	"time"
)

const (
	DateFormat       = "20060102"
	DefaultTaskLimit = 50
	DateFormatDB     = "02.01.2006"
)

// ParseDate парсит строку с датой в объект времени
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(DateFormat, dateStr)
}

// FormatDate форматирует дату в строку в формате "20060102"
func FormatDate(date time.Time) string {
	return date.Format(DateFormat)
}

// FormatDateDB форматирует дату в строку в формате ""02.01.2006""
func FormatDateDB(date time.Time) string {
	return date.Format(DateFormatDB)
}

// ParseDateDB парсит строку с датой в объект времени в формате ""02.01.2006""
func ParseDateDB(dateStr string) (time.Time, error) {
	return time.Parse(DateFormatDB, dateStr)
}
