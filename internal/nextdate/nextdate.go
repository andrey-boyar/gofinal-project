package nextdate

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату для задачи в соответствии с правилом повторения.
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if now.IsZero() || dateStr == "" || repeat == "" {
		return "", fmt.Errorf("некорректные входные данные")
	}

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %w", err)
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch {
	case repeat == "y":
		return handleYearlyRepeat(now, date)
	case strings.HasPrefix(repeat, "d "):
		return handleDailyRepeat(now, date, repeat)
	case strings.HasPrefix(repeat, "w "):
		return handleWeeklyRepeat(now, repeat)
	case strings.HasPrefix(repeat, "m "):
		return handleMonthlyRepeat(now, date, repeat)
	default:
		return "", fmt.Errorf("неизвестное правило повторения: %s", repeat)
	}
}

func handleDailyRepeat(now, date time.Time, repeat string) (string, error) {
	daysStr := strings.TrimPrefix(repeat, "d ")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 400 {
		return "", fmt.Errorf("неверный формат правила повторения: %w", err)
	}
	for date.Before(now) {
		date = date.AddDate(0, 0, days)
	}
	return date.Format("20060102"), nil
}

func handleYearlyRepeat(now, date time.Time) (string, error) {
	for date.Before(now) {
		date = date.AddDate(1, 0, 0)
	}
	if date.Month() == time.February && date.Day() == 29 && !isLeapYear(date.Year()) {
		date = date.AddDate(0, 0, 1) // Перейти на 1 марта в невисокосные годы
	}
	return date.Format("20060102"), nil
}

func handleWeeklyRepeat(now time.Time, repeat string) (string, error) {

	daysOfWeekStr := strings.TrimPrefix(repeat, "w ")
	daysOfWeek := strings.Split(daysOfWeekStr, ",")
	nowWeekDay := int(now.Weekday())
	if nowWeekDay == 0 {
		nowWeekDay = 7
	}
	var repeatDays []int
	for _, day := range daysOfWeek {
		dayNumber, err := strconv.Atoi(day)
		if err != nil || dayNumber < 1 || dayNumber > 7 {
			return "", fmt.Errorf("неверный формат правила повторения: %w", err)
		}
		if dayNumber <= nowWeekDay {
			dayNumber += 7
		}
		repeatDays = append(repeatDays, dayNumber)
	}
	slices.Sort(repeatDays)
	shift := repeatDays[0] - nowWeekDay
	nextDate := now.AddDate(0, 0, shift)
	return nextDate.Format("20060102"), nil
}

func handleMonthlyRepeat(now, date time.Time, repeat string) (string, error) {
	mParts := strings.Split(strings.TrimPrefix(repeat, "m "), " ")
	allowDays, err := parseDays(mParts[0])
	if err != nil {
		return "", fmt.Errorf("неверный формат правила повторения: %w", err)
	}
	allowMonths, err := parseMonths(mParts)
	if err != nil {
		return "", fmt.Errorf("неверный формат правила повторения: %w", err)
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
		currentMonth := date.Month()
		for {
			if currentMonth != date.Month() {
				break
			}
			if isSliceHas(allowDaysInMonth, date.Day()) && date.After(now) {
				return date.Format("20060102"), nil
			}
			date = date.AddDate(0, 0, 1)
		}
	}
}

func parseDays(daysStr string) ([]int, error) {
	days := strings.Split(daysStr, ",")
	allowDays := make([]int, 0, len(days))
	for _, day := range days {
		dayInt, err := strconv.Atoi(day)
		if err != nil || dayInt < -2 || dayInt > 31 {
			return nil, fmt.Errorf("неверный формат правила повторения: %w", err)
		}
		allowDays = append(allowDays, dayInt)
	}
	return allowDays, nil
}

func parseMonths(format []string) ([]int, error) {
	if len(format) < 2 {
		return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, nil
	}
	monthsStr := strings.Split(format[1], ",")
	allowMonths := make([]int, 0, len(monthsStr))
	for _, month := range monthsStr {
		monthInt, err := strconv.Atoi(month)
		if err != nil || monthInt < 1 || monthInt > 12 {
			return nil, fmt.Errorf("неверный формат правила повторения: %w", err)
		}
		allowMonths = append(allowMonths, monthInt)
	}
	return allowMonths, nil
}

func isSliceHas(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func makeAllowDaysForMonth(date time.Time, days []int) []int {
	daysInMonth := daysIn(date.Month(), date.Year())
	allowDays := make([]int, 0, len(days))
	for _, day := range days {
		if day > 0 && day <= daysInMonth {
			allowDays = append(allowDays, day)
		} else if day < 0 {
			allowDays = append(allowDays, daysInMonth+day+1)
		}
	}
	return allowDays
}

func daysIn(month time.Month, year int) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func isLeapYear(year int) bool {
	if year%4 == 0 {
		if year%100 == 0 {
			return year%400 == 0
		}
		return true
	}
	return false
}
