package main

import (
	"errors"
	"fmt"
	"time"
)

const DateFormat = "20060102"

func NextDate(now time.Time, startDate, repeat string) (string, error) {
	parsedStartDate, err := time.Parse(DateFormat, startDate)
	if err != nil {
		return "", errors.New("Invalid start date format")
	}

	var unit string
	var interval int
	_, err = fmt.Sscanf(repeat, "%s %d", &unit, &interval)
	if err != nil || interval <= 0 {
		return "", errors.New("Invalid repetition rule")
	}

	switch unit {
	case "d":
		parsedStartDate = parsedStartDate.AddDate(0, 0, interval)
	case "w":
		parsedStartDate = parsedStartDate.AddDate(0, 0, 7*interval)
	case "m":
		parsedStartDate = parsedStartDate.AddDate(0, interval, 0)
	case "y":
		parsedStartDate = parsedStartDate.AddDate(interval, 0, 0)
	default:
		return "", errors.New("Unsupported repetition unit")
	}

	for parsedStartDate.Before(now) {
		switch unit {
		case "d":
			parsedStartDate = parsedStartDate.AddDate(0, 0, interval)
		case "w":
			parsedStartDate = parsedStartDate.AddDate(0, 0, 7*interval)
		case "m":
			parsedStartDate = parsedStartDate.AddDate(0, interval, 0)
		case "y":
			parsedStartDate = parsedStartDate.AddDate(interval, 0, 0)
		}
	}

	return parsedStartDate.Format(DateFormat), nil
}
