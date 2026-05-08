package util

import "fmt"

func IntToHHmm(minutes int) (string, error) {
	if minutes < 0 || minutes > 1440 {
		return "", fmt.Errorf("minutes must be between 0 and 1440, got %d", minutes)
	}
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60), nil
}

func Truncate(s string, maxLen int) string {
	count := 0
	for i := range s {
		if count == maxLen {
			return s[:i]
		}
		count++
	}
	return s
}

func HHmmToInt(s string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, fmt.Errorf("invalid HH:mm format: %s", s)
	}
	minutes := h*60 + m
	if minutes < 0 || minutes > 1440 {
		return 0, fmt.Errorf("minutes must be between 0 and 1440, got %d", minutes)
	}
	return minutes, nil
}
