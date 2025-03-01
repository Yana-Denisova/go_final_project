package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	t, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("date не может быть преобразовано в корректную дату: %w", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("не задан формат повторения")
	}

	if repeat == "y" {
		newDate := t.AddDate(1, 0, 0)
		for newDate.Format("20060102") <= now.Format("20060102") {
			newDate = newDate.AddDate(1, 0, 0)
		}
		newDateStr := newDate.Format("20060102")
		return newDateStr, nil
	}

	parts := strings.Split(repeat, " ")
	if len(parts) < 2 {
		return "", fmt.Errorf("некорректный формат повторения")
	}

	letter := parts[0]
	dayNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("ошибка преобразования: %w", err)
	}

	if letter == "d" {
		if dayNumber > 400 {
			return "", fmt.Errorf("максимально допустимое число днй равно 400: %w", err)
		}
		newDate := t.AddDate(0, 0, dayNumber)
		for newDate.Format("20060102") <= now.Format("20060102") {
			newDate = newDate.AddDate(0, 0, dayNumber)
		}
		newDateStr := newDate.Format("20060102")
		return newDateStr, nil
	} else {
		return "", fmt.Errorf("некорректный формат повторения")
	}
}
