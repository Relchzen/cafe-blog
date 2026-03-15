package utils

import (
	"fmt"
	"time"
)

func ValidateAndParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be in format YYYY-MM-DD")
	}

	return parsedDate, nil // Return the parsed time.Time object
}
