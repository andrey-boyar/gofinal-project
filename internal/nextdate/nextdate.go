package nextdate

import (
	"errors"
	"final-project/internal/utils"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату, основываясь на текущей дате, заданной дате и правиле повтора.
func NextDate(now time.Time, date string, repeat string) (string, error) {
	log.Printf("NextDate вызван с параметрами: now=%v, date=%s, repeat=%s", now, date, repeat)
	// Парсим строку с датой в объект времени
	dateTime, err := time.Parse(utils.DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты")
	}
	// Если дата задачи уже в будущем, возвращаем её без изменений
	//if dateTime.After(now) {
	//	return date, nil
	//}

	// Если правило повторения пустое, возвращаем ошибку
	if repeat == "" {
		return "", fmt.Errorf("правило повтора не указано")
	}
	// Определяем следующий период на основе правила повторения
	switch {
	case repeat == "y":
		// Если повторение ежегодное
		for {
			dateTime = dateTime.AddDate(1, 0, 0)
			if dateTime.After(now) {
				break
			}
		}
	case strings.HasPrefix(repeat, "d "):
		// Если повторение через определенное количество дней
		days, err := strconv.Atoi(strings.TrimPrefix(repeat, "d "))
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("неверный 'd' формат повтора")
		}
		for {
			dateTime = dateTime.AddDate(0, 0, days)
			if dateTime.After(now) {
				break
			}
		}
	case strings.HasPrefix(repeat, "w "):
		// Если повторение через определенные дни недели
		repeatDaysStr := strings.Split(strings.TrimPrefix(repeat, "w "), ",")
		repeatDays := make([]int, 0, len(repeatDaysStr))
		for _, day := range repeatDaysStr {
			dayNumber, err := strconv.Atoi(day)
			if err != nil || dayNumber < 1 || dayNumber > 7 {
				return "", fmt.Errorf("неверный день недели: %s", day)
			}
			repeatDays = append(repeatDays, dayNumber)
		}

		// Вычисляем следующую дату
		nextDate, err := calculateNextDateWeekly(now, dateTime, repeatDays)
		if err != nil {
			return "", err
		}
		return nextDate, nil

	case strings.HasPrefix(repeat, "m "):
		// Если повторение через определенные дни и месяцы
		format := strings.Split(strings.TrimPrefix(repeat, "m "), " ")
		allowDays, err := parsDay(format)
		if err != nil {
			return "", fmt.Errorf("неверный формат повтора")
		}
		allowMonths, err := parsMonth(format)
		if err != nil {
			return "", fmt.Errorf("неверный формат повтора")
		}

		for {
			if !isSliceHas(allowMonths, int(dateTime.Month())) {
				dateTime = dateTime.AddDate(0, 1, 0)
				if dateTime.Day() > 1 {
					dateTime = dateTime.AddDate(0, 0, -dateTime.Day()+1)
				}
				continue
			}

			allowDaysInMonth := calculateNextDateMonthly(dateTime, allowDays)
			currentMonth := dateTime.Month()
			for {
				if currentMonth != dateTime.Month() {
					break
				}
				if isSliceHas(allowDaysInMonth, dateTime.Day()) &&
					dateTime.After(now) {
					return dateTime.Format(utils.DateFormat), nil
				}
				dateTime = dateTime.AddDate(0, 0, 1)
			}
		}

	default:
		return "", fmt.Errorf("неверный формат повтора")
	}

	return dateTime.Format(utils.DateFormat), nil
}

// calculateNextDateWeekly вычисляет следующую дату для еженедельного повтора.
func calculateNextDateWeekly(now, start time.Time, days []int) (string, error) {
	if len(days) == 0 {
		return "", fmt.Errorf("не указаны дни недели для повтора")
	}
	sort.Ints(days)

	next := start
	if next.Before(now) {
		next = now
	}

	for {
		nowWeekday := int(next.Weekday())
		if nowWeekday == 0 {
			nowWeekday = 7
		}

		for _, day := range days {
			daysUntil := (day - nowWeekday + 7) % 7
			candidateDate := next.AddDate(0, 0, daysUntil)
			if candidateDate.After(now) {
				return candidateDate.Format(utils.DateFormat), nil
			}
		}
		next = next.AddDate(0, 0, 1)
	}
}

// calculateNextDateMonthly вычисляет следующую дату для ежемесячного повтора.
func calculateNextDateMonthly(date time.Time, days []int) []int {
	daysInMonth := daysInsert(date.Month(), date.Year())
	result := make([]int, 0, len(days))
	for _, d := range days {
		if d > daysInMonth {
			continue
		}
		if d > 0 {
			result = append(result, d)
			continue
		}
		result = append(result, daysInMonth+d+1)
	}
	return result
}

// daysIn возвращает количество дней в указанном месяце и году
func daysInsert(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// parsDay разбирает строку с днями в формате, поддерживаемом повторением
func parsDay(format []string) ([]int, error) {
	daysStr := strings.Split(format[0], ",")
	allowDays := make([]int, 0, len(daysStr))
	for _, dayS := range daysStr {
		if day, err := strconv.ParseInt(dayS, 10, 64); err == nil {
			if day < -2 || day > 31 {
				return []int{}, errors.New("неверный формат повтора")
			}
			allowDays = append(allowDays, int(day))
		}
	}

	return allowDays, nil
}

// parsMonth разбирает строку с месяцами в формате, поддерживаемом повторением
func parsMonth(format []string) ([]int, error) {
	allowMonth := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	if len(format) < 2 {
		return allowMonth, nil
	}

	monthStr := strings.Split(format[1], ",")
	allowMonth = make([]int, 0, len(monthStr))
	for _, ms := range monthStr {
		if month, err := strconv.ParseInt(ms, 10, 64); err == nil {
			if month < 1 || month > 12 {
				return []int{}, errors.New("неверный формат повтора")
			}
			allowMonth = append(allowMonth, int(month))
		}
	}
	return allowMonth, nil
}

// isSliceHas проверяет, содержится ли значение в срезе
func isSliceHas(s []int, v int) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

// contains проверяет, содержится ли значение в срезе
func contains(slice []int, item int) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
