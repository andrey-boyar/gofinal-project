package nextdate

import (
	"errors"
<<<<<<< HEAD
	"final-project/internal/utils"
=======
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"final-project/internal/utils"
)

// NextDate вычисляет следующую дату, основываясь на текущей дате, заданной дате и правиле повтора.
func NextDate(now time.Time, date string, repeat string) (string, error) {
	log.Printf("NextDate вызван с параметрами: now=%v, date=%s, repeat=%s", now, date, repeat)
	// Парсим строку с датой в объект времени
	dateTime, err := time.Parse(utils.DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты")
	}
	// Если правило повторения пустое, возвращаем ошибку
	if repeat == "" {
		return "", errors.New("правило повтора не указано")
	}
<<<<<<< HEAD
	// Определяем следующий период на основе правила повторения
	switch {
	case repeat == "y":
		// Если повторение ежегодное
		for {
			dateTime = dateTime.AddDate(1, 0, 0)
			if dateTime.After(now) {
				break
=======

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
		if days > 400 {
			return "", fmt.Errorf("интервал повтора не может быть больше 400 дней")
		}
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
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
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
		nowWeekDay := int(now.Weekday())
		if nowWeekDay == 0 {
			nowWeekDay = 7
		}
		repeatDaysStr := strings.Split(strings.TrimPrefix(repeat, "w "), ",")
		repeatDays := make([]int, 0, len(repeatDaysStr))
		for _, day := range repeatDaysStr {
			if dayNumber, parseErr := strconv.ParseInt(day, 10, 64); parseErr == nil {
				if dayNumber < 1 || dayNumber > 7 {
					return "", fmt.Errorf("неверный день недели: %d", dayNumber)
				}
				if int(dayNumber) <= nowWeekDay {
					dayNumber += 7
				}
				repeatDays = append(repeatDays, int(dayNumber))
			}
		}
<<<<<<< HEAD
		sort.Ints(repeatDays)
		shift := repeatDays[0] - nowWeekDay
		date = now.AddDate(0, 0, shift).Format(utils.DateFormat)

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

=======
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
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
	default:
		return "", fmt.Errorf("неверный формат повтора")
	}
<<<<<<< HEAD

	return dateTime.Format(utils.DateFormat), nil
}

// calculateNextDateWeekly вычисляет следующую дату для еженедельного повтора.
=======
	// return "", nil
	return utils.FormatDate(startDate), nil
}

// функция для вычисления следующей даты по дням недели
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
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
	if next.Before(now) {
		next = now
	}
<<<<<<< HEAD
=======
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
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
	for {
		for _, wd := range weekdays {
			candidateDate := next
			daysUntilWeekday := (int(wd) - int(candidateDate.Weekday()) + 7) % 7
			candidateDate = candidateDate.AddDate(0, 0, daysUntilWeekday)
			if candidateDate.After(now) {
				return candidateDate.Format(utils.DateFormat), nil
			}
		}
		next = next.AddDate(0, 0, 7)
	}
}

<<<<<<< HEAD
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
=======
// функция для вычисления следующей даты по годам
func abs(n int) int {
	if n < 0 {
		return -n
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
	}
	return result
}

<<<<<<< HEAD
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
=======
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
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
			return true
		}
	}
	return false
}

<<<<<<< HEAD
// contains проверяет, содержится ли значение в срезе
func contains(slice []int, item int) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
=======
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
		if err != nil || days < 1 || days > 400 {
			return fmt.Errorf("неверное количество дней: %s (должно быть от 1 до 400)", parts[1])
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
>>>>>>> 49cbd60aebcbc098619db0606202eea4fda9a289
	}
	return false
}
