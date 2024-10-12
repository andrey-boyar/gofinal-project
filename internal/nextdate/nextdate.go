package nextdate

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Парсим исходную дату
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	// Проверяем правило повторения
	if repeat == "" {
		return "", fmt.Errorf("правило повтора не указано")
	}

	parts := strings.Split(repeat, " ")
	if len(parts) < 1 {
		return "", fmt.Errorf("некорректный формат повтора")
	}

	repeatType := parts[0]
	//var interval int
	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("некорректный интервал повтора")
	}

	switch repeatType {
	case "d":
		if interval < 1 || interval > 400 {
			return "", fmt.Errorf("некорректный интервал для ежедневного повтора")
		}
		next := startDate
		for next.Before(now) || next.Equal(now) {
			next = next.AddDate(0, 0, interval)
		}
		//return next.Format("20060102"), nil
		// Возвращаем исходную дату, если она больше текущей
		if startDate.After(now) {
			return date, nil
		}
		days := int(now.Sub(startDate).Hours() / 24)
		nextDate := startDate.AddDate(0, 0, ((days/interval)+1)*interval)
		// Иначе вычисляем следующую дату
		/*nextDate := startDate
		for nextDate.Before(now) || nextDate.Equal(now) {
			nextDate = nextDate.AddDate(0, 0, interval)*/
		//}
		return nextDate.Format("20060102"), nil

	/*if len(parts) != 2 {
	return "", fmt.Errorf("некорректный формат для ежедневного повтора")
	}
	interval, err = strconv.Atoi(parts[1])
	if err != nil || interval < 1 || interval > 400 {
		return "", fmt.Errorf("некорректный интервал для ежедневного повтора")
	}
	return calculateNextDate(now, startDate, func(t time.Time) time.Time {
		return t.AddDate(0, 0, interval)
	})*/
	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат для еженедельного повтора")
		}
		days := strings.Split(parts[1], ",")
		return calculateNextDateWeekly(now, startDate, days)
	case "m":
		if len(parts) < 2 {
			return "", fmt.Errorf("некорректный формат для ежемесячного повтора")
		}
		return calculateNextDateMonthly(now, startDate, parts[1:])
	case "y":
		/*if len(parts) != 1 {
			return "", fmt.Errorf("некорректный формат для ежегодного повтора")
		}
		return calculateNextDate(now, startDate, func(t time.Time) time.Time {
			return t.AddDate(1, 0, 0)
		})*/if len(parts) != 2 {
			return "", fmt.Errorf("некорректный формат для ежегодного повтора")
		}
		return calculateNextDateYearly(now, startDate, parts[1])
	default:
		return "", fmt.Errorf("неподдерживаемый тип повтора: %s", repeatType)
	}
}

//func calculateNextDate(now, start time.Time, incrementFunc func(time.Time) time.Time) (string, error) {
//	next := start
//for next.Before(now) || next.Equal(now) {
//	next = incrementFunc(next)	}
//return next.Format("20060102"), nil}

func calculateNextDateWeekly(now, start time.Time, days []string) (string, error) {
	if len(days) == 0 {
		return "", fmt.Errorf("не указаны дни недели для повтора")
	}
	sort.Strings(days)
	weekdays := make([]time.Weekday, 0, len(days))
	for _, day := range days {
		d, err := strconv.Atoi(day)
		if err != nil || d < 1 || d > 7 {
			return "", fmt.Errorf("некорректный день недели: %s", day)
		}
		weekdays = append(weekdays, time.Weekday((d-1)%7))
	}

	next := start
	for next.Before(now) || next.Equal(now) {
		next = next.AddDate(0, 0, 1)
		for _, wd := range weekdays {
			if next.Weekday() == wd && next.After(now) {
				return next.Format("20060102"), nil
			}
		}
	}
	return "", fmt.Errorf("не удалось найти следующую дату")
}

func calculateNextDateMonthly(now, start time.Time, parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("не указаны дни месяца для повтора")
	}
	sort.Strings(parts)
	days := strings.Split(parts[0], ",")
	var months []int
	if len(parts) > 1 {
		for _, m := range strings.Split(parts[1], ",") {
			month, err := strconv.Atoi(m)
			if err != nil || month < 1 || month > 12 {
				return "", fmt.Errorf("некорректный месяц: %s", m)
			}
			months = append(months, month)
		}
	}

	next := start
	for next.Before(now) || next.Equal(now) {
		next = next.AddDate(0, 1, 0)
		if len(months) > 0 && !contains(months, int(next.Month())) {
			next = next.AddDate(0, 1, 0)
			continue
		}
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err != nil {
				return "", fmt.Errorf("некорректный день месяца: %s", day)
			}
			candidate := time.Date(next.Year(), next.Month(), d, 0, 0, 0, 0, next.Location())
			if candidate.Month() == next.Month() && candidate.After(now) {
				return candidate.Format("20060102"), nil
			}
		}
		next = next.AddDate(0, 1, 0)
	}
	return "", fmt.Errorf("не удалось найти следующую дату")
}
func calculateNextDateYearly(now, start time.Time, monthDay string) (string, error) {
	parts := strings.Split(monthDay, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("некорректный формат месяца и дня для ежегодного повтора")
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return "", fmt.Errorf("некорректный месяц для ежегодного повтора")
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil || day < 1 || day > 31 {
		return "", fmt.Errorf("некорректный день для ежегодного повтора")
	}

	nextDate := time.Date(start.Year(), time.Month(month), day, 0, 0, 0, 0, start.Location())
	for nextDate.Before(now) || nextDate.Equal(now) {
		nextDate = nextDate.AddDate(1, 0, 0)
	}

	return nextDate.Format("20060102"), nil
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// ValidateRepeatFormat проверяет формат повтора
func ValidateRepeatFormat(repeat string) error {
	if repeat == "" {
		return nil
	}
	parts := strings.Split(repeat, " ")
	if len(parts) != 2 {
		return fmt.Errorf("неверный формат повтора")
	}
	period := parts[0]
	interval, err := strconv.Atoi(parts[1])
	if err != nil || interval <= 0 {
		return fmt.Errorf("неверный интервал повтора")
	}
	switch period {
	case "d", "w", "m", "y":
		return nil
	default:
		return fmt.Errorf("неверный тип повтора")
	}
}

//func isLeapYear(year int) bool {
//	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
//}
