package user

import (
	"sort"
	"strings"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

var (
	ErrDailyScreenTimeOutOfRange = core.NewDomainError("user.invalid_daily_screen_time", "daily screen time limit is out of range")

	ErrDisplayNameTooLong  = core.NewDomainError("user.display_name_too_long", "display name is too long")
	ErrDisplayNameTooShort = core.NewDomainError("user.display_name_too_short", "display name is too short")

	ErrLanguageCodeNotSupported = core.NewDomainError("user.unsupported_language_code", "language code is not supported")

	ErrDailyScreenTimeLimitRangeOrder = core.NewDomainError("user.invalid_screen_time_range_order", "screen time range start must be before end")
	ErrDailyScreenTimeLimitOutOfRange = core.NewDomainError("user.invalid_screen_time_range", "daily screen time limit is out of range")

	ErrInvalidLeaveReasonCode = core.NewDomainError("user.invalid_leave_reason_code", "invalid leave reason code")
)

type LeaveReasonCode int

var leaveReasonCodeMap = []struct {
	code LeaveReasonCode
	str  string
}{
	{code: 0, str: "self"},
}

func NewLeaveReasonCode(s string) (_ LeaveReasonCode, err error) {
	defer util.Wrap(&err, "user.NewLeaveReasonCode")

	for _, c := range leaveReasonCodeMap {
		if s == c.str {
			return c.code, nil
		}
	}

	return 0, ErrInvalidLeaveReasonCode
}

func (l LeaveReasonCode) String() string {
	for _, c := range leaveReasonCodeMap {
		if c.code == l {
			return c.str
		}
	}

	return "self"
}

const unlimitedScreenTimeSentinel = 86401 // 24h + 1s

// IsUnlimitedScreenTimeSeconds はDBから取得した視聴制限秒数が無制限を表すかを判定する。
func IsUnlimitedScreenTimeSeconds(seconds int) bool {
	return seconds >= int((24 * time.Hour).Seconds())
}

type DailyScreenTimeLimit struct {
	duration *time.Duration
}

func NewDailyScreenTimeLimit(seconds *int) (_ DailyScreenTimeLimit, err error) {
	defer util.Wrap(&err, "user.NewDailyScreenTimeLimit")

	if seconds == nil || *seconds >= int((24 * time.Hour).Seconds()) {
		return DailyScreenTimeLimit{duration: nil}, nil
	}

	value := time.Duration(*seconds) * time.Second
	if value < 0 {
		return DailyScreenTimeLimit{}, ErrDailyScreenTimeOutOfRange
	}

	return DailyScreenTimeLimit{duration: &value}, nil
}

func (d DailyScreenTimeLimit) IsUnlimited() bool {
	return d.duration == nil
}

func (d DailyScreenTimeLimit) Seconds() int {
	if d.duration == nil {
		return unlimitedScreenTimeSentinel
	}
	return int(d.duration.Seconds())
}

func (d DailyScreenTimeLimit) ToIntPtr() *int {
	if d.duration == nil {
		return nil
	}
	v := int(d.duration.Seconds())
	return &v
}

type DisplayName string

func NewDisplayName(s string) (_ DisplayName, err error) {
	defer util.Wrap(&err, "user.NewDisplayName")

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

var languageCodeMap = []struct {
	code LanguageCode
	str  string
}{
	{code: "ja", str: "ja"},
	{code: "en", str: "en"},
}

func NewLanguageCode(value string) (_ LanguageCode, err error) {
	defer util.Wrap(&err, "user.NewLanguageCode")

	for _, c := range languageCodeMap {
		if value == c.str {
			return c.code, nil
		}
	}

	return "", ErrLanguageCodeNotSupported
}

func (l LanguageCode) String() string {
	return string(l)
}

type DailyScreenTimeLimitRange struct {
	ID               uuid.UUID
	StartTimeSeconds int
	EndTimeSeconds   int
}

func NewDailyScreenTimeLimitRange(startTimeSeconds, endTimeSeconds int) (_ DailyScreenTimeLimitRange, err error) {
	defer util.Wrap(&err, "user.NewDailyScreenTimeLimitRange")

	if startTimeSeconds >= endTimeSeconds {
		return DailyScreenTimeLimitRange{}, ErrDailyScreenTimeLimitRangeOrder
	}

	if startTimeSeconds < 0 || time.Duration(startTimeSeconds)*time.Second >= 24*time.Hour {
		return DailyScreenTimeLimitRange{}, ErrDailyScreenTimeLimitOutOfRange
	}
	if endTimeSeconds < 0 || time.Duration(endTimeSeconds)*time.Second > 24*time.Hour {
		return DailyScreenTimeLimitRange{}, ErrDailyScreenTimeLimitOutOfRange
	}

	id, err := uuid.NewV7()
	if err != nil {
		return DailyScreenTimeLimitRange{}, err
	}

	return DailyScreenTimeLimitRange{
		ID:               id,
		StartTimeSeconds: startTimeSeconds,
		EndTimeSeconds:   endTimeSeconds,
	}, nil
}

type DailyScreenTimeLimitRangeSet struct {
	Ranges []DailyScreenTimeLimitRange
}

func NewDailyScreenTimeLimitRangeSet(screenLimits []struct{ Start, End int }, loc *time.Location) (_ *DailyScreenTimeLimitRangeSet, err error) {
	defer util.Wrap(&err, "user.NewDailyScreenTimeLimitRangeSet")

	// ローカル時刻の秒数をUTC秒数に変換する。
	// 基準日はタイムゾーンオフセットの計算にのみ使うため任意の日付でよい。
	ref := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
	_, offsetSec := ref.Zone()
	const daySeconds = 24 * 60 * 60

	wrap := func(s int) int { return ((s % daySeconds) + daySeconds) % daySeconds }

	// NOTE: UX的に、時間範囲の重複は許容する
	var ranges []DailyScreenTimeLimitRange
	for _, r := range screenLimits {
		utcStart := wrap(r.Start - offsetSec)
		utcEnd := wrap(r.End - offsetSec)

		if utcStart < utcEnd {
			// 日付を跨がない
			domainRange, err := NewDailyScreenTimeLimitRange(utcStart, utcEnd)
			if err != nil {
				return nil, err
			}
			ranges = append(ranges, domainRange)
		} else {
			// 日付を跨ぐ: [utcStart, 24h] と [0, utcEnd)
			r1, err := NewDailyScreenTimeLimitRange(utcStart, daySeconds)
			if err != nil {
				return nil, err
			}
			ranges = append(ranges, r1)
			if utcEnd > 0 {
				r2, err := NewDailyScreenTimeLimitRange(0, utcEnd)
				if err != nil {
					return nil, err
				}
				ranges = append(ranges, r2)
			}
		}
	}
	// ソートしてマージ
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].StartTimeSeconds < ranges[j].StartTimeSeconds })
	var merged []DailyScreenTimeLimitRange
	for _, r := range ranges {
		if len(merged) > 0 && r.StartTimeSeconds <= merged[len(merged)-1].EndTimeSeconds+1 {
			if r.EndTimeSeconds > merged[len(merged)-1].EndTimeSeconds {
				merged[len(merged)-1].EndTimeSeconds = r.EndTimeSeconds
			}
		} else {
			merged = append(merged, r)
		}
	}

	return &DailyScreenTimeLimitRangeSet{
		Ranges: merged,
	}, nil
}

// BlockedUntil は現在時刻が許可された時間範囲外の場合、次の許可開始時刻を返す。
// 範囲が空の場合は許可時間帯が存在しないため常にブロックする。
func (s *DailyScreenTimeLimitRangeSet) BlockedUntil(now time.Time) *time.Time {
	if len(s.Ranges) == 0 {
		// 許可時間帯が一つもない場合は常にブロック
		sentinel := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
		return &sentinel
	}

	nowSeconds := now.Hour()*3600 + now.Minute()*60 + now.Second()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	for _, r := range s.Ranges {
		if nowSeconds >= r.StartTimeSeconds && nowSeconds < r.EndTimeSeconds {
			return nil
		}
	}

	var best time.Time
	found := false
	for _, r := range s.Ranges {
		if r.StartTimeSeconds > nowSeconds {
			t := today.Add(time.Duration(r.StartTimeSeconds) * time.Second)
			if !found || t.Before(best) {
				best = t
				found = true
			}
		}
	}
	if found {
		return &best
	}

	tomorrow := today.AddDate(0, 0, 1)
	for _, r := range s.Ranges {
		t := tomorrow.Add(time.Duration(r.StartTimeSeconds) * time.Second)
		if !found || t.Before(best) {
			best = t
			found = true
		}
	}
	if found {
		return &best
	}
	return nil
}

type User struct {
	ID              uuid.UUID
	DisplayName     DisplayName
	LanguageCode    LanguageCode
	JoinedAt        time.Time
	ScreenTimeLimit DailyScreenTimeLimit
}

type UserOption func(*User)

func WithUserID(id uuid.UUID) UserOption {
	return func(u *User) {
		u.ID = id
	}
}

func WithUserJoinedAt(joinedAt time.Time) UserOption {
	return func(u *User) {
		u.JoinedAt = joinedAt
	}
}

func (u *User) SetDisplayName(displayName *string) (err error) {
	defer util.Wrap(&err, "user.(*User).SetDisplayName")

	if displayName == nil {
		return nil
	}
	dn, err := NewDisplayName(*displayName)
	if err != nil {
		return err
	}
	u.DisplayName = dn
	return nil
}

func (u *User) SetLanguageCode(languageCode *string) (err error) {
	defer util.Wrap(&err, "user.(*User).SetLanguageCode")

	if languageCode == nil {
		return nil
	}
	lc, err := NewLanguageCode(*languageCode)
	if err != nil {
		return err
	}
	u.LanguageCode = lc
	return nil
}

func (u *User) SetScreenTimeLimit(dailyScreenLimit *int) (err error) {
	defer util.Wrap(&err, "user.(*User).SetScreenTimeLimit")

	if dailyScreenLimit == nil {
		return nil
	}
	stl, err := NewDailyScreenTimeLimit(dailyScreenLimit)
	if err != nil {
		return err
	}
	u.ScreenTimeLimit = stl
	return nil
}

func NewUser(displayName string, languageCode string, dailyScreenLimit *int, opts ...UserOption) (_ *User, err error) {
	defer util.Wrap(&err, "user.NewUser")

	dn, err := NewDisplayName(displayName)
	if err != nil {
		return nil, err
	}
	lc, err := NewLanguageCode(languageCode)
	if err != nil {
		return nil, err
	}
	stl, err := NewDailyScreenTimeLimit(dailyScreenLimit)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:              id,
		DisplayName:     dn,
		LanguageCode:    lc,
		JoinedAt:        time.Now().UTC(),
		ScreenTimeLimit: stl,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u, nil
}
