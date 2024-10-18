package nextdate

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"final-project/internal/utils"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	log.Printf("NextDate вызван с параметрами: now=%v, date=%s, repeat=%s", now, date, repeat)
	// Парсим исходную дату
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	// Проверяем правило повторения
	if repeat == "" {
		return "", errors.New("правило повтора не указано")
	}

	parts := strings.Split(repeat, " ")
	if len(parts) < 2 {
		return "", errors.New("некорректный формат повтора")
	}

	repeatType := parts[0]
	// var interval int
	// interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("некорректный интервал повтора")
	}
	switch repeatType {
	case "d":
		days, _ := strconv.Atoi(parts[1])
		return startDate.AddDate(0, 0, days).Format("20060102"), nil
	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("неверный формат для еженедельного повтора: должно быть 'w D,D,D,D,D,D,D'")
		}
		days := strings.Split(parts[1], ",")
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err != nil || d < 1 || d > 7 {
				return "", fmt.Errorf("неверный день недели: %s", day)
			}
		}
	case "m":
		days := strings.Split(parts[1], ",")
		var months []string
		if len(parts) > 2 {
			months = strings.Split(parts[2], ",")
		}
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("неверный формат для ежемесячного повтора: должно быть 'm D' или 'm D M'")
		}
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err != nil || (d == 0 || d < -31 || d > 31) {
				return "", fmt.Errorf("неверный день месяца: %s", day)
			}
		}
		if len(parts) == 3 {
			months := strings.Split(parts[2], ",")
			for _, month := range months {
				m, err := strconv.Atoi(month)
				if err != nil || m < 1 || m > 12 {
					return "", fmt.Errorf("неверный месяц: %s", month)
				}
			}
		}
		return calculateNextDateMonthly(now, startDate, days, months)
	case "y":
		if len(parts) == 1 {
			// Если указан только "y", используем дату из параметра date
			return calculateNextDateYearly(now, startDate, startDate.Format("01.02"))
		}
		return calculateNextDateYearly(now, startDate, parts[1])

		// if len(parts) != 2 {
		// 	return "", fmt.Errorf("неверный формат для ежегодного повтора: должно быть 'y MM.DD'")
		// }
		dateParts := strings.Split(parts[1], ".")
		if len(dateParts) != 2 {
			return "", fmt.Errorf("неверный формат даты для ежегодного повтора: должно быть MM.DD")
		}
		monthDay := parts[1]
		nextDate, err := calculateNextDateYearly(now, startDate, monthDay)
		if err != nil {
			return "", err
		}
		return nextDate, nil
		// month, _ := strconv.Atoi(dateParts[0])
		// day, _ := strconv.Atoi(dateParts[1])
		// return calculateNextDateYearly(now, startDate, parts[1])
	default:
		return "", fmt.Errorf("неподдерживаемый тип повтора: %s", repeatType)
	}
	// return "", nil
	return utils.FormatDate(startDate), nil
}

// функция для вычисления следующей даты по дням недели
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

// функция для вычисления следующей даты по месяцам
func calculateNextDateMonthly(now, start time.Time, days []string, months []string) (string, error) {
	// Преобразование дней в числа
	var dayNums []int
	for _, day := range days {
		d, _ := strconv.Atoi(day)
		dayNums = append(dayNums, d)
	}

	// Преобразование месяцев в числа
	var monthNums []int
	if len(months) > 0 {
		for _, month := range months {
			m, _ := strconv.Atoi(month)
			monthNums = append(monthNums, m)
		}
	} else {
		// Если месяцы не указаны, используем все месяцы
		monthNums = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	}

	next := start
	for {
		next = next.AddDate(0, 1, 0)
		if next.After(now) {
			if contains(monthNums, int(next.Month())) {
				for _, day := range dayNums {
					testDate := time.Date(next.Year(), next.Month(), abs(day), 0, 0, 0, 0, next.Location())
					if day < 0 {
						// Для отрицательных дней отсчитываем с конца месяца
						testDate = time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, next.Location()).AddDate(0, 0, day+1)
					}
					if testDate.After(now) {
						return testDate.Format("20060102"), nil
					}
				}
			}
		}
	}
}

// функция для вычисления следующей даты по годам
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// функция для вычисления следующей даты по годам
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

// функция для проверки наличия элемента в массиве
func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// функция для проверки формата повтора
func ValidateRepeatFormat(repeat string) error {
	if repeat == "" {
		return nil // Пустой повтор допустим
	}
	parts := strings.Split(repeat, " ")
	if len(parts) < 1 {
		return fmt.Errorf("неверный формат повтора: недостаточно параметров")
	}

	switch parts[0] {
	case "d": // ежедневный
		if len(parts) != 2 {
			return fmt.Errorf("неверный формат для ежедневного повтора: должно быть 'd N'")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 {
			return fmt.Errorf("неверное количество дней: %s", parts[1])
		}
	case "w": // еженедельный
		if len(parts) != 2 {
			return fmt.Errorf("неверный формат для еженедельного повтора: должно быть 'w D, D, D, D, D, D, D'")
		}
		days := strings.Split(parts[1], ",")
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err != nil || d < 1 || d > 7 {
				return fmt.Errorf("неверный день недели: %s", day)
			}
		}
	case "m": // ежемесячный
		if len(parts) < 2 || len(parts) > 3 {
			return fmt.Errorf("неверный формат для ежемесячного повтора: должно быть 'm D' или 'm D M'")
		}
		if len(parts) != 2 && len(parts) != 3 {
			return fmt.Errorf("неверный формат для ежемесячного повтора: должно быть 'm D' или 'm D M'")
		}
		days := strings.Split(parts[1], ",")
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err != nil || (d == 0 || d < -31 || d > 31) {
				return fmt.Errorf("неверный день месяца: %s", day)
			}
		}
		if len(parts) == 3 {
			months := strings.Split(parts[2], ",")
			for _, month := range months {
				m, err := strconv.Atoi(month)
				if err != nil || m < 1 || m > 12 {
					return fmt.Errorf("неверный месяц: %s", month)
				}
			}
		}
	case "y": // ежегодный
		if len(parts) == 1 {
			return nil // Допускаем формат "y" без даты
		}
		if len(parts) != 2 {
			return fmt.Errorf("неверный формат для ежегодного повтора: должно быть 'y MM.DD'")
		}
		dateParts := strings.Split(parts[1], ".")
		if len(dateParts) != 2 {
			return fmt.Errorf("неверный формат даты для ежегодного повтора: должно быть MM.DD")
		}
		// Добавьте проверку формата MM.DD
		if _, err := time.Parse("01.02", parts[1]); err != nil {
			return fmt.Errorf("неверный формат даты для ежегодного повтора: %s", parts[1])
		}
		month, err := strconv.Atoi(dateParts[0])
		if err != nil || month < 1 || month > 12 {
			return fmt.Errorf("неверный месяц: %s", dateParts[0])
		}
		day, err := strconv.Atoi(dateParts[1])
		if err != nil || day < 1 || day > 31 {
			return fmt.Errorf("неверный день: %s", dateParts[1])
		}
	default:
		return fmt.Errorf("неподдерживаемый тип повтора: %s", parts[0])
	}
	return nil
}

// функция для нормализации формата повтора
func CorrectRepeatFormat(repeat string) string {
	parts := strings.Split(repeat, " ")
	if len(parts) == 1 {
		switch parts[0] {
		case "d": // ежедневный
			return "d 1"
		case "w": // еженедельный
			return "w 1,2,3,4,5,6,7"
		case "m": // ежемесячный
			return "m 1"
		case "y": // ежегодный
			return "y 01.01"
		}
	}
	return repeat
}

// функция для нормализации формата повтора
func NormalizeRepeatFormat(repeat string) string {
	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return repeat
	}

	switch parts[0] {
	case "w": // еженедельный
		if len(parts) == 1 {
			return "w 1,2,3,4,5,6,7"
		}
		days := strings.Split(parts[1], ",")
		uniqueDays := make(map[int]bool)
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err == nil && d >= 1 && d <= 7 {
				uniqueDays[d] = true
			}
		}
		if len(uniqueDays) == 0 {
			return "w 1,2,3,4,5,6,7"
		}
		normalizedDays := make([]string, 0, len(uniqueDays))
		for i := 1; i <= 7; i++ {
			if uniqueDays[i] {
				normalizedDays = append(normalizedDays, strconv.Itoa(i))
			}
		}
		return fmt.Sprintf("w %s", strings.Join(normalizedDays, ","))

	case "m": // ежемесячный
		if len(parts) == 1 {
			return "m 1"
		}
		days := strings.Split(parts[1], ",")
		uniqueDays := make(map[int]bool)
		for _, day := range days {
			d, err := strconv.Atoi(day)
			if err == nil && (d >= -31 && d <= -1 || d >= 1 && d <= 31) {
				uniqueDays[d] = true
			}
		}
		if len(uniqueDays) == 0 {
			return "m 1"
		}
		normalizedDays := make([]string, 0, len(uniqueDays))
		for i := -31; i <= 31; i++ {
			if i != 0 && uniqueDays[i] {
				normalizedDays = append(normalizedDays, strconv.Itoa(i))
			}
		}
		if len(parts) > 2 {
			months := strings.Split(parts[2], ",")
			uniqueMonths := make(map[int]bool)
			for _, month := range months {
				m, err := strconv.Atoi(month)
				if err == nil && m >= 1 && m <= 12 {
					uniqueMonths[m] = true
				}
			}
			if len(uniqueMonths) > 0 {
				normalizedMonths := make([]string, 0, len(uniqueMonths))
				for i := 1; i <= 12; i++ {
					if uniqueMonths[i] {
						normalizedMonths = append(normalizedMonths, strconv.Itoa(i))
					}
				}
				return fmt.Sprintf("m %s %s", strings.Join(normalizedDays, ","), strings.Join(normalizedMonths, ","))
			}
		}
		return fmt.Sprintf("m %s", strings.Join(normalizedDays, ","))

	case "y": // ежегодный
		if len(parts) == 1 {
			return "y 01.01"
		}
		dateParts := strings.Split(parts[1], ".")
		if len(dateParts) != 2 {
			return "y 01.01"
		}
		month, err1 := strconv.Atoi(dateParts[0])
		day, err2 := strconv.Atoi(dateParts[1])
		if err1 != nil || err2 != nil || month < 1 || month > 12 || day < 1 || day > 31 {
			return "y 01.01"
		}
		return fmt.Sprintf("y %02d.%02d", month, day)

	default:
		return repeat
	}
}
