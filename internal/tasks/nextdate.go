package tasks

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const DP = "20060102"

// NextDate вычисляет следующую дату для задачи в соответствии с правилом повторения.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	date, err := time.Parse(DP, dateStr)
	if err != nil {
		return "", errors.New("invalid date format")
	}
	if repeat == "" {
		return "", errors.New("repeat rule is not specified")
	}

	return handleRepeat(now, date, repeat)
}

// handleRepeat обрабатывает правила повторения.
func handleRepeat(now, date time.Time, repeat string) (string, error) {
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
		return "", fmt.Errorf("repeat rule is not specified: %s", repeat)
	}
}

// handleYearlyRepeat вычисляет следующую дату для ежегодного повтора.
func handleYearlyRepeat(now, date time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) {
			break
		}
	}
	if date.Month() == time.February && date.Day() == 29 && !isLeapYear(date.Year()+1) {
		date = time.Date(date.Year()+1, 3, 1, 0, 0, 0, 0, date.Location())
	}
	return date.Format(DP), nil
}

// handleDailyRepeat вычисляет следующую дату для ежедневного повтора.
func handleDailyRepeat(now, date time.Time, repeat string) (string, error) {
	days, err := strconv.Atoi(strings.TrimPrefix(repeat, "d "))
	if err != nil || days < 1 || days > 400 {
		return "", errors.New("invalid 'd' repeat format")
	}
	for {
		date = date.AddDate(0, 0, days)
		if date.After(now) {
			break
		}
	}
	return date.Format(DP), nil
}

// handleWeeklyRepeat вычисляет следующую дату для недельного повтора.
func handleWeeklyRepeat(now, date time.Time, repeat string) (string, error) {
	repeatDaysStr := strings.Split(strings.TrimPrefix(repeat, "w "), ",")
	repeatDays := make([]int, 0, len(repeatDaysStr))

	// Извлечение дней недели и проверка их корректности
	for _, day := range repeatDaysStr {
		dayNumber, parseErr := strconv.Atoi(day)
		if parseErr != nil || dayNumber < 1 || dayNumber > 7 {
			return "", errors.New("invalid repeat format: days of the week should be from 1 to 7")
		}
		repeatDays = append(repeatDays, dayNumber)
	}

	if len(repeatDays) == 0 {
		return "", errors.New("invalid repeat format: no days of the week are present")
	}

	// Сортировка дней недели
	slices.Sort(repeatDays)

	// Поиск следующей даты, соответствующей указанным дням недели
	for date.Before(now) || date.Equal(now) {
		date = date.AddDate(0, 0, 1) // Переход к следующему дню
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7 // Преобразование воскресенья в 7
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

// handleMonthlyRepeat вычисляет следующую дату для месячного повтора.
func handleMonthlyRepeat(now, date time.Time, repeat string) (string, error) {
	// Извлечение числа из строки повтора
	repeatParts := strings.Split(strings.TrimPrefix(repeat, "m "), " ")
	if len(repeatParts) < 1 {
		return "", errors.New("неверный формат 'm' повтора")
	}

	// Парсинг дней
	repeatDaysStr := strings.Split(repeatParts[0], ",")
	repeatDays := make([]int, 0, len(repeatDaysStr))
	for _, day := range repeatDaysStr {
		dayNumber, err := strconv.Atoi(day)
		if err != nil || dayNumber < -31 || dayNumber > 31 {
			return "", errors.New("неверный формат 'm' повтора: допустимые значения от -31 до 31")
		}
		repeatDays = append(repeatDays, dayNumber)
	}

	// Логика для нахождения следующей даты
	for {
		date = date.AddDate(0, 1, 0) // Переход к следующему месяцу
		for _, day := range repeatDays {
			var nextDate time.Time
			if day == -1 { // последний день месяца
				nextDate = time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
			} else if day == -2 { // предпоследний день месяца
				nextDate = time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).AddDate(0, 0, -1)
			} else if day < 0 { // другие отрицательные дни
				nextDate = time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).AddDate(0, 0, day+1)
			} else { // положительные дни
				nextDate = time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, date.Location())
			}
			if nextDate.After(now) {
				return nextDate.Format(DP), nil
			}
		}
	}
}

// isLeapYear проверяет, является ли год високосным.
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
