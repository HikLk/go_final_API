package main

import (
	"errors"
	"fmt"
	"time"
)

// NextDate вычисляет следующую дату для задачи в зависимости от типа повторения
func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Парсим дату из строки
	parsedDate, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", errors.New("Неверный формат даты")
	}

	// Проверяем и вычисляем следующую дату на основе повторений
	var nextDate time.Time
	switch repeat {
	case "daily":
		nextDate = parsedDate.AddDate(0, 0, 1)
	case "weekly":
		nextDate = parsedDate.AddDate(0, 0, 7)
	case "monthly":
		nextDate = parsedDate.AddDate(0, 1, 0)
	default:
		return "", fmt.Errorf("Неизвестный тип повторения: %s", repeat)
	}

	// Возвращаем следующую дату в нужном формате
	return nextDate.Format(DateFormat), nil
}

// Task представляет структуру задачи
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
