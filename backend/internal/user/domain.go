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

	ErrLanguageCodeNotSupported = errors.New("given language code is not supported yet")

	ErrDailyScreenTimeLimitRangeOrder = errors.New("start >= end")
	ErrDailyScreenTimeLimitOutOfRange = errors.New("!(0 <= value <= 24 * Hour)")
)

const DailyScreenTimeLimitInfinity = 24*time.Hour + 1*time.Second

type DailyScreenTimeLimit time.Duration

func NewDailyScreenTimeLimit(seconds *int) (*DailyScreenTimeLimit, error) {
	if seconds == nil {
		domain := DailyScreenTimeLimit(DailyScreenTimeLimitInfinity)
		return &domain, nil
	}

	value := time.Duration(*seconds) * time.Second
	if value < 0 || value >= 24*time.Hour {
		return nil, ErrDailyScreenTimeOutOfRange
	}

	domain := DailyScreenTimeLimit(value)
	return &domain, nil
}

func (d *DailyScreenTimeLimit) IsInfinity() bool {
	if d == nil {
		return true
	}

	duration := (time.Duration)(*d)
	return duration > 24*time.Hour
}

func (d *DailyScreenTimeLimit) ToInt() *int {
	if d == nil {
		return nil
	}
	s := (int)(time.Duration(*d).Seconds())
	return &s
}

type DisplayName string

func NewDisplayName(s string) (*DisplayName, error) {
	str := strings.TrimSpace(s)

	length := len([]rune(str))
	if length < 1 {
		return nil, ErrDisplayNameTooShort
	}
	if length > 29 {
		return nil, ErrDisplayNameTooLong
	}

	domain := DisplayName(str)
	return &domain, nil
}

func (d *DisplayName) String() string {
	if d == nil {
		return ""
	}
	return (string)(*d)
}

type LanguageCode string

func NewLanguageCode(value string) (*LanguageCode, error) {
	if value != "ja" {
		return nil, ErrLanguageCodeNotSupported
	}

	domain := LanguageCode(value)
	return &domain, nil
}

func (l *LanguageCode) String() string {
	if l == nil {
		return ""
	}
	return (string)(*l)
}

type DailyScreenTimeLimitRange struct {
	StartTimeSeconds int
	EndTimeSeconds   int
}

func NewDailyScreenTimeLimitRange(startTimeSeconds, endTimeSeconds int) (*DailyScreenTimeLimitRange, error) {
	if startTimeSeconds >= endTimeSeconds {
		return nil, ErrDailyScreenTimeLimitRangeOrder
	}

	if startTimeSeconds < 0 || time.Duration(startTimeSeconds)*time.Second >= 24*time.Hour {
		return nil, ErrDailyScreenTimeLimitOutOfRange
	}
	if endTimeSeconds < 0 || time.Duration(endTimeSeconds)*time.Second >= 24*time.Hour {
		return nil, ErrDailyScreenTimeLimitOutOfRange
	}

	return &DailyScreenTimeLimitRange{
		StartTimeSeconds: startTimeSeconds,
		EndTimeSeconds:   endTimeSeconds,
	}, nil
}

type DailyScreenTimeLimitRangeSet []*DailyScreenTimeLimitRange

func NewDailyScreenTimeLimitRangeSet(limitRanges []struct{ Start, End int }) (*DailyScreenTimeLimitRangeSet, error) {
	// NOTE: UX的に、時間範囲の重複は許容する
	domainRanges := make([]*DailyScreenTimeLimitRange, len(limitRanges))
	for i, r := range limitRanges {
		domainRange, err := NewDailyScreenTimeLimitRange(r.Start, r.End)
		if err != nil {
			return nil, err
		}

		domainRanges[i] = domainRange
	}

	d := (DailyScreenTimeLimitRangeSet)(domainRanges)
	return &d, nil
}

type User struct {
	UserID               uuid.UUID
	DisplayName          string
	LanguageCode         string
	JoinedAt             time.Time
	ScreenTimeLimitRange []struct {
		ID                       uuid.UUID
		StartSeconds, EndSeconds int
	}
	ScreenTimeSeconds *int
	RemainingSeconds  *int
}

func NewUser(userID uuid.UUID, displayName string, languageCode string, joinedAt time.Time,
	screenTimeLimitRange []struct {
		ID                       uuid.UUID
		StartSeconds, EndSeconds int
	}, screenTimeSeconds *int, remainingSeconds *int) *User {

	return &User{
		UserID:               userID,
		DisplayName:          displayName,
		LanguageCode:         languageCode,
		JoinedAt:             joinedAt,
		ScreenTimeLimitRange: screenTimeLimitRange,
		ScreenTimeSeconds:    screenTimeSeconds,
		RemainingSeconds:     remainingSeconds,
	}
}
