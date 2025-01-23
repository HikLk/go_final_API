package main

import (
	"errors"
	"fmt"
	"time"
)

// Формат даты: ГГГГММДД
const DateFormat = "20060102"

// NextDate вычисляет следующую дату задачи на основе текущей даты, стартовой даты и правила повторения
func NextDate(now time.Time, startDate, repeat string) (string, error) {
	// Парсим стартовую дату
	parsedStartDate, err := time.Parse(DateFormat, startDate)
	if err != nil {
		return "", errors.New("Неверный формат стартовой даты")
	}

	// Инициализируем переменные для хранения единицы и интервала повторения
	var unit string
	var interval int

	// Считываем правило повторения в формате "<единица> <интервал>"
	_, err = fmt.Sscanf(repeat, "%s %d", &unit, &interval)
	if err != nil || interval <= 0 {
		return "", errors.New("Неверное правило повторения")
	}

	// Вычисляем следующую дату в зависимости от единицы времени
	switch unit {
	case "d": // дни
		parsedStartDate = parsedStartDate.AddDate(0, 0, interval)
	case "w": // недели
		parsedStartDate = parsedStartDate.AddDate(0, 0, 7*interval)
	case "m": // месяцы
		parsedStartDate = parsedStartDate.AddDate(0, interval, 0)
	case "y": // годы
		parsedStartDate = parsedStartDate.AddDate(interval, 0, 0)
	default:
		return "", errors.New("Неподдерживаемая единица повторения")
	}

	// Если следующая дата раньше текущей, увеличиваем её до тех пор, пока она не станет позже текущей
	for parsedStartDate.Before(now) {
		switch unit {
		case "d": // дни
			parsedStartDate = parsedStartDate.AddDate(0, 0, interval)
		case "w": // недели
			parsedStartDate = parsedStartDate.AddDate(0, 0, 7*interval)
		case "m": // месяцы
			parsedStartDate = parsedStartDate.AddDate(0, interval, 0)
		case "y": // годы
			parsedStartDate = parsedStartDate.AddDate(interval, 0, 0)
		}
	}

	// Возвращаем вычисленную дату в формате ГГГГММДД
	return parsedStartDate.Format(DateFormat), nil
}
