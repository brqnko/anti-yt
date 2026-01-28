package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDailyScreenTimeOutOfRange = errors.New("!(0 <= value <= 24 * Hour)")

	ErrDisplayNameTooLong  = errors.New("display name is too long")
	ErrDisplayNameTooShort = errors.New("display name is too short")
)

type DailyScreenTimeLimit *time.Duration

func NewDailyScreenTimeLimit(seconds int) (DailyScreenTimeLimit, error) {
	value := time.Duration(seconds) * time.Second
	if value < 0 || value >= 24*time.Hour {
		return nil, ErrDailyScreenTimeOutOfRange
	}

	return DailyScreenTimeLimit(&value), nil
}

type DisplayName *string

func NewDisplayName(s *string) (DisplayName, error) {
	if s == nil {
		return nil, nil
	}

	str := strings.TrimSpace(*s)

	length := len(str)
	if length < 1 {
		return nil, ErrDisplayNameTooShort
	}
	if length > 50 {
		return nil, ErrDisplayNameTooLong
	}

	return DisplayName(s), nil
}

type DailyScreenTimeLimitRange struct {
	StartTimeSeconds int
	EndTimeSeconds   int
}

func NewDailyScreenTimeLimitRange(startTimeSeconds, endTimeSeconds int) (DailyScreenTimeLimitRange, error) {
	return DailyScreenTimeLimitRange{
		StartTimeSeconds: startTimeSeconds,
		EndTimeSeconds:   endTimeSeconds,
	}, nil
}

type DailyScreenTimeLimitRangeSet struct {
	ranges []DailyScreenTimeLimitRange
}

func NewDailyScreenTimeLimitRangeSet(ranges []DailyScreenTimeLimitRange) (DailyScreenTimeLimitRangeSet, error) {
	return DailyScreenTimeLimitRangeSet{
		ranges: ranges,
	}, nil
}

type User struct {
	PublicID                     uuid.UUID
	JoinedAt                     time.Time
	DailyScreenTimeLimit         DailyScreenTimeLimit
	LanguageCode                 string
	DisplayName                  DisplayName
	DailyScreenTimeLimitRangeSet DailyScreenTimeLimitRangeSet
}

type CreateUserInput struct {
	DailyScreenTimeLimit         DailyScreenTimeLimit
	LanguageCode                 string
	DisplayName                  DisplayName
	DailyScreenTimeLimitRangeSet DailyScreenTimeLimitRangeSet
}

func NewCreateUserInput(displayName *string, languageCode string, dailyScreenLimit *time.Duration, dailyScreenTimeLimitRangeSet []struct{ start, end int }) (*CreateUserInput, error) {
	return nil, nil
}
