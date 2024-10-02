package nextdate

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
	//"fmt"
)

const DP = "20060102"

// NextDate вычисляет следующую дату для задачи в соответствии с правилом повторения.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	// Парсим дату
	date, err := time.Parse(DP, dateStr)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}
	// Если правило повторения пустое, возвращаем ошибку
	if repeat == "" {
		return "", errors.New("правило повтора не указано")
	}

	// Определяем следующий период на основе правила повторения
	switch {
	case repeat == "y":
		return handleYearlyRepeat(now, date)
	case strings.HasPrefix(repeat, "d "):
		return handleDailyRepeat(now, date, repeat)
	case strings.HasPrefix(repeat, "w "):
		return handleWeeklyRepeat(now, date, repeat)
	case strings.HasPrefix(repeat, "m "):
		return handleMonthlyRepeat(now, date, repeat)
	default:
		return "", errors.New("неверный формат повтора")
	}
}

// функция для вычисления следующей даты для задачи в соответствии с правилом повторения
func handleYearlyRepeat(now, date time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) {
			break
		}
	}
	// Проверка на 29 февраля
	if date.Month() == time.February && date.Day() == 29 && !isLeapYear(date.Year()+1) {
		date = time.Date(date.Year()+1, 3, 1, 0, 0, 0, 0, date.Location())
	}
	return date.Format(DP), nil
}

// функция для вычисления следующей даты для задачи в соответствии с правилом повторения
func handleDailyRepeat(now, date time.Time, repeat string) (string, error) {
	days, err := strconv.Atoi(strings.TrimPrefix(repeat, "d "))
	if err != nil || days < 1 || days > 400 {
		return "", errors.New("неверный 'd' формат повтора")
	}
	for {
		date = date.AddDate(0, 0, days)
		if date.After(now) {
			break
		}
	}
	return date.Format(DP), nil
}

// функция для вычисления следующей даты для задачи в соответствии с правилом повторения
func handleWeeklyRepeat(now, date time.Time, repeat string) (string, error) {
	repeatDaysStr := strings.Split(strings.TrimPrefix(repeat, "w "), ",")
	repeatDays := make([]int, 0, len(repeatDaysStr))
	for _, day := range repeatDaysStr {
		if dayNumber, parseErr := strconv.ParseInt(day, 10, 64); parseErr == nil {
			if dayNumber < 1 || dayNumber > 7 {
				return "", errors.New("неверный формат повтора")
			}
			repeatDays = append(repeatDays, int(dayNumber))
		}
	}
	if len(repeatDays) == 0 {
		return "", errors.New("неверный формат повтора")
	}
	slices.Sort(repeatDays)

	// Используем date вместо now для начальной точки отсчета
	for date.Before(now) || date.Equal(now) {
		date = date.AddDate(0, 0, 1)
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if slices.Contains(repeatDays, weekday) {
			return date.Format(DP), nil
		}
	}

	// Если не нашли подходящую дату, продолжаем поиск
	for {
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if slices.Contains(repeatDays, weekday) {
			return date.Format(DP), nil
		}
		date = date.AddDate(0, 0, 1)
	}
}

// функция для вычисления следующей даты для задачи в соответствии с правилом повторения
func handleMonthlyRepeat(now, date time.Time, repeat string) (string, error) {
	format := strings.Split(strings.TrimPrefix(repeat, "m "), " ")
	allowDays, err := parseDays(format)
	if err != nil {
		return "", errors.New("неверный формат повтора")
	}
	allowMonths, err := parseMonths(format)
	if err != nil {
		return "", errors.New("неверный формат повтора")
	}
	// Проверяем, все ли дни отрицательные
	allNegative := true
	for _, day := range allowDays {
		if day > 0 {
			allNegative = false
			break
		}
	}

	for {
		if !isSliceHas(allowMonths, int(date.Month())) {
			date = date.AddDate(0, 1, 0)
			if date.Day() > 1 {
				date = date.AddDate(0, 0, -date.Day()+1)
			}
			continue
		}

		allowDaysInMonth := makeAllowDaysForMonth(date, allowDays)
		//if len(allowDaysInMonth) == 0 {
		// Если нет допустимых дней в месяце, возвращаем пустую строку
		//return "", nil
		//}
		if len(allowDaysInMonth) == 0 {
			if allNegative {
				// Если все дни отрицательные и их нет в текущем месяце, возвращаем пустую строку
				return "", nil
			}
			date = date.AddDate(0, 1, 0)
			continue
		}
		currentMonth := date.Month()
		for {
			if currentMonth != date.Month() {
				break
			}
			if isSliceHas(allowDaysInMonth, date.Day()) && date.After(now) {
				return date.Format(DP), nil
			}
			date = date.AddDate(0, 0, 1)
		}
	}
}

// функция для парсинга дней
func parseDays(format []string) ([]int, error) {
	daysStr := strings.Split(format[0], ",")
	allowDays := make([]int, 0, len(daysStr))
	for _, dayS := range daysStr {
		if day, err := strconv.ParseInt(dayS, 10, 64); err == nil {
			if day < -31 || day > 31 || day == 0 {
				return []int{}, errors.New("неверный формат повтора")
			}
			allowDays = append(allowDays, int(day))
		}
	}
	return allowDays, nil
}

// функция для парсинга месяцев
func parseMonths(format []string) ([]int, error) {
	if len(format) < 2 {
		return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, nil
	}
	monthStr := strings.Split(format[1], ",")
	allowMonths := make([]int, 0, len(monthStr))
	for _, ms := range monthStr {
		if month, err := strconv.ParseInt(ms, 10, 64); err == nil {
			if month < 1 || month > 12 {
				return []int{}, errors.New("неверный формат повтора")
			}
			allowMonths = append(allowMonths, int(month))
		}
	}
	if len(allowMonths) == 0 {
		return []int{}, errors.New("неверный формат повтора")
	}
	return allowMonths, nil
}

// функция для проверки наличия элемента в массиве
func isSliceHas(s []int, v int) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

// функция для создания допустимых дней для месяца
func makeAllowDaysForMonth(date time.Time, days []int) []int {
	daysInMonth := daysIn(date.Month(), date.Year())
	result := make([]int, 0, len(days))
	for _, d := range days {
		if d > 0 && d <= daysInMonth {
			result = append(result, d)
		} else if d < 0 {
			actualDay := daysInMonth + d + 1
			if actualDay > 0 && actualDay <= daysInMonth {
				result = append(result, actualDay)
			}
		}
	}
	slices.Sort(result)
	return result
}

// функция для вычисления количества дней в месяце
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// функция для проверки високосного года
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
