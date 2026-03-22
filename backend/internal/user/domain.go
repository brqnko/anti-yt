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

// CalcRemainingSeconds はDBから取得した1日の視聴制限秒数と今日の視聴秒数から残り秒数を計算する。
// limitSeconds が24時間以上（無制限センチネル値）の場合は -1 を返す。
func CalcRemainingSeconds(limitSeconds, watchedSeconds int) int {
	if limitSeconds >= int((24 * time.Hour).Seconds()) {
		return -1
	}
	return max(0, limitSeconds-watchedSeconds)
}

type DailyScreenTimeLimit time.Duration

func NewDailyScreenTimeLimit(seconds *int) (DailyScreenTimeLimit, error) {
	if seconds == nil {
		return DailyScreenTimeLimit(DailyScreenTimeLimitInfinity), nil
	}

	value := time.Duration(*seconds) * time.Second
	if value < 0 || value >= 24*time.Hour {
		return 0, ErrDailyScreenTimeOutOfRange
	}

	return DailyScreenTimeLimit(value), nil
}

func (d DailyScreenTimeLimit) IsInfinity() bool {
	return time.Duration(d) > 24*time.Hour
}

func (d DailyScreenTimeLimit) ToInt() int {
	return int(time.Duration(d).Seconds())
}

type DisplayName string

func NewDisplayName(s string) (DisplayName, error) {
	str := strings.TrimSpace(s)

	length := len([]rune(str))
	if length < 1 {
		return "", ErrDisplayNameTooShort
	}
	if length > 29 {
		return "", ErrDisplayNameTooLong
	}

	return DisplayName(str), nil
}

func (d DisplayName) String() string {
	return string(d)
}

type LanguageCode string

func NewLanguageCode(value string) (LanguageCode, error) {
	if value != "ja" {
		return "", ErrLanguageCodeNotSupported
	}

	return LanguageCode(value), nil
}

func (l LanguageCode) String() string {
	return string(l)
}

type DailyScreenTimeLimitRange struct {
	StartTimeSeconds int
	EndTimeSeconds   int
}

func NewDailyScreenTimeLimitRange(startTimeSeconds, endTimeSeconds int) (DailyScreenTimeLimitRange, error) {
	if startTimeSeconds >= endTimeSeconds {
		return DailyScreenTimeLimitRange{}, ErrDailyScreenTimeLimitRangeOrder
	}

	if startTimeSeconds < 0 || time.Duration(startTimeSeconds)*time.Second >= 24*time.Hour {
		return DailyScreenTimeLimitRange{}, ErrDailyScreenTimeLimitOutOfRange
	}
	if endTimeSeconds < 0 || time.Duration(endTimeSeconds)*time.Second >= 24*time.Hour {
		return DailyScreenTimeLimitRange{}, ErrDailyScreenTimeLimitOutOfRange
	}

	return DailyScreenTimeLimitRange{
		StartTimeSeconds: startTimeSeconds,
		EndTimeSeconds:   endTimeSeconds,
	}, nil
}

type DailyScreenTimeLimitRangeSet []DailyScreenTimeLimitRange

func NewDailyScreenTimeLimitRangeSet(limitRanges []struct{ Start, End int }) (DailyScreenTimeLimitRangeSet, error) {
	// NOTE: UX的に、時間範囲の重複は許容する
	domainRanges := make([]DailyScreenTimeLimitRange, len(limitRanges))
	for i, r := range limitRanges {
		domainRange, err := NewDailyScreenTimeLimitRange(r.Start, r.End)
		if err != nil {
			return nil, err
		}

		domainRanges[i] = domainRange
	}

	return DailyScreenTimeLimitRangeSet(domainRanges), nil
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
	ScreenTimeSeconds int
	RemainingSeconds  int
}

func NewUser(userID uuid.UUID, displayName string, languageCode string, joinedAt time.Time,
	screenTimeLimitRange []struct {
		ID                       uuid.UUID
		StartSeconds, EndSeconds int
	}, screenTimeSeconds int, remainingSeconds int) User {

	return User{
		UserID:               userID,
		DisplayName:          displayName,
		LanguageCode:         languageCode,
		JoinedAt:             joinedAt,
		ScreenTimeLimitRange: screenTimeLimitRange,
		ScreenTimeSeconds:    screenTimeSeconds,
		RemainingSeconds:     remainingSeconds,
	}
}
