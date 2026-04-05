package user_test

import (
	"strings"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestNewLeaveReasonCode(t *testing.T) {
	t.Parallel()

	type arg struct {
		str string
	}

	type want struct {
		code user.LeaveReasonCode
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"self": {
			arg:     arg{str: "self"},
			want:    &want{code: 0},
			wantErr: nil,
		},
		"invalid": {
			arg:     arg{str: "invalid"},
			want:    nil,
			wantErr: user.ErrInvalidLeaveReasonCode,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := user.NewLeaveReasonCode(c.arg.str)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.code, got)
			}
		})
	}
}

func TestNewDailyScreenTimeLimit(t *testing.T) {
	t.Parallel()

	ptr := func(v int) *int { return &v }

	type arg struct {
		seconds *int
	}

	type want struct {
		isUnlimited bool
		seconds     int
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"nil is unlimited": {
			arg:  arg{seconds: nil},
			want: &want{isUnlimited: true},
		},
		"86400 is unlimited": {
			arg:  arg{seconds: ptr(86400)},
			want: &want{isUnlimited: true},
		},
		"success": {
			arg:  arg{seconds: ptr(3600)},
			want: &want{isUnlimited: false, seconds: 3600},
		},
		"negative": {
			arg:     arg{seconds: ptr(-1)},
			wantErr: user.ErrDailyScreenTimeOutOfRange,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := user.NewDailyScreenTimeLimit(c.arg.seconds)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.isUnlimited, got.IsUnlimited())
				if !c.want.isUnlimited {
					assert.Equal(t, c.want.seconds, got.Seconds())
				}
			}
		})
	}
}

func TestIsUnlimitedScreenTimeSeconds(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		seconds int
		want    bool
	}{
		"86400 is unlimited": {
			seconds: 86400,
			want:    true,
		},
		"86399 is not unlimited": {
			seconds: 86399,
			want:    false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got := user.IsUnlimitedScreenTimeSeconds(c.seconds)

			// assert
			assert.Equal(t, c.want, got)
		})
	}
}

func TestNewDisplayName(t *testing.T) {
	t.Parallel()

	type arg struct {
		name string
	}

	type want struct {
		name user.DisplayName
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{name: "testuser"},
			want:    &want{name: user.DisplayName("testuser")},
			wantErr: nil,
		},
		"29 chars": {
			arg:     arg{name: strings.Repeat("a", 29)},
			want:    &want{name: user.DisplayName(strings.Repeat("a", 29))},
			wantErr: nil,
		},
		"30 chars": {
			arg:     arg{name: strings.Repeat("a", 30)},
			want:    nil,
			wantErr: user.ErrDisplayNameTooLong,
		},
		"empty": {
			arg:     arg{name: ""},
			want:    nil,
			wantErr: user.ErrDisplayNameTooShort,
		},
		"spaces only": {
			arg:     arg{name: "   "},
			want:    nil,
			wantErr: user.ErrDisplayNameTooShort,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := user.NewDisplayName(c.arg.name)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.name, got)
			}
		})
	}
}

func TestNewLanguageCode(t *testing.T) {
	t.Parallel()

	type arg struct {
		str string
	}

	type want struct {
		code user.LanguageCode
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"ja": {
			arg:     arg{str: "ja"},
			want:    &want{code: "ja"},
			wantErr: nil,
		},
		"en": {
			arg:     arg{str: "en"},
			want:    &want{code: "en"},
			wantErr: nil,
		},
		"invalid": {
			arg:     arg{str: "invalid"},
			want:    nil,
			wantErr: user.ErrLanguageCodeNotSupported,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := user.NewLanguageCode(c.arg.str)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.code, got)
			}
		})
	}
}

func TestNewDailyScreenTimeLimitRange(t *testing.T) {
	t.Parallel()

	type arg struct {
		start int
		end   int
	}

	type want struct {
		start int
		end   int
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{start: 3600, end: 7200},
			want:    &want{start: 3600, end: 7200},
			wantErr: nil,
		},
		"start equals end": {
			arg:     arg{start: 3600, end: 3600},
			wantErr: user.ErrDailyScreenTimeLimitRangeOrder,
		},
		"start after end": {
			arg:     arg{start: 7200, end: 3600},
			wantErr: user.ErrDailyScreenTimeLimitRangeOrder,
		},
		"start is negative": {
			arg:     arg{start: -1, end: 3600},
			wantErr: user.ErrDailyScreenTimeLimitOutOfRange,
		},
		"start is 86400": {
			arg:     arg{start: 86400, end: 86401},
			wantErr: user.ErrDailyScreenTimeLimitOutOfRange,
		},
		"end exceeds 86400": {
			arg:     arg{start: 3600, end: 86401},
			wantErr: user.ErrDailyScreenTimeLimitOutOfRange,
		},
		"end is 86400": {
			arg:     arg{start: 3600, end: 86400},
			want:    &want{start: 3600, end: 86400},
			wantErr: nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := user.NewDailyScreenTimeLimitRange(c.arg.start, c.arg.end)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.start, got.StartTimeSeconds)
				assert.Equal(t, c.want.end, got.EndTimeSeconds)
			}
		})
	}
}

func TestNewDailyScreenTimeLimitRangeSet(t *testing.T) {
	t.Parallel()

	type rangeInput struct {
		Start, End int
	}

	type arg struct {
		ranges []rangeInput
		loc    *time.Location
	}

	type want struct {
		ranges []user.DailyScreenTimeLimitRange
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"empty": {
			arg:  arg{ranges: []rangeInput{}, loc: time.UTC},
			want: &want{ranges: nil},
		},
		"single range": {
			arg:  arg{ranges: []rangeInput{{Start: 3600, End: 7200}}, loc: time.UTC},
			want: &want{ranges: []user.DailyScreenTimeLimitRange{{StartTimeSeconds: 3600, EndTimeSeconds: 7200}}},
		},
		"overlapping ranges are merged": {
			arg: arg{
				ranges: []rangeInput{{Start: 3600, End: 7200}, {Start: 5400, End: 10800}},
				loc:    time.UTC,
			},
			want: &want{ranges: []user.DailyScreenTimeLimitRange{{StartTimeSeconds: 3600, EndTimeSeconds: 10800}}},
		},
		"non-overlapping ranges": {
			arg: arg{
				ranges: []rangeInput{{Start: 3600, End: 7200}, {Start: 10800, End: 14400}},
				loc:    time.UTC,
			},
			want: &want{ranges: []user.DailyScreenTimeLimitRange{
				{StartTimeSeconds: 3600, EndTimeSeconds: 7200},
				{StartTimeSeconds: 10800, EndTimeSeconds: 14400},
			}},
		},
		"crosses midnight (JST UTC+9)": {
			// JST 23:00-01:00 → UTC 14:00-16:00 (no wrap)
			arg: arg{
				ranges: []rangeInput{{Start: 23 * 3600, End: 25 * 3600}},
				loc:    time.FixedZone("JST", 9*3600),
			},
			want: &want{ranges: []user.DailyScreenTimeLimitRange{{StartTimeSeconds: 14 * 3600, EndTimeSeconds: 16 * 3600}}},
		},
		"start equals end is ignored": {
			arg:  arg{ranges: []rangeInput{{Start: 3600, End: 3600}}, loc: time.UTC},
			want: &want{ranges: nil},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ranges := make([]struct{ Start, End int }, len(c.arg.ranges))
			for i, r := range c.arg.ranges {
				ranges[i] = struct{ Start, End int }{Start: r.Start, End: r.End}
			}

			// act
			got, err := user.NewDailyScreenTimeLimitRangeSet(ranges, c.arg.loc)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.ranges, got.Ranges)
			}
		})
	}
}

func TestDailyScreenTimeLimitRangeSet_BlockedUntil(t *testing.T) {
	t.Parallel()

	// now = 2025-01-01 10:00:00 UTC (= 36000 seconds)
	now := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	today := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	rangeSet := func(ranges ...user.DailyScreenTimeLimitRange) *user.DailyScreenTimeLimitRangeSet {
		return &user.DailyScreenTimeLimitRangeSet{Ranges: ranges}
	}
	ptr := func(t time.Time) *time.Time { return &t }

	cases := map[string]struct {
		rangeSet *user.DailyScreenTimeLimitRangeSet
		want     *time.Time
	}{
		"no ranges: always blocked": {
			rangeSet: rangeSet(),
			want:     ptr(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)),
		},
		"within range: not blocked": {
			rangeSet: rangeSet(user.DailyScreenTimeLimitRange{StartTimeSeconds: 9 * 3600, EndTimeSeconds: 11 * 3600}),
			want:     nil,
		},
		"before range today: blocked until range start": {
			rangeSet: rangeSet(user.DailyScreenTimeLimitRange{StartTimeSeconds: 12 * 3600, EndTimeSeconds: 14 * 3600}),
			want:     ptr(today.Add(12 * time.Hour)),
		},
		"after all ranges today: blocked until tomorrow": {
			rangeSet: rangeSet(user.DailyScreenTimeLimitRange{StartTimeSeconds: 8 * 3600, EndTimeSeconds: 9 * 3600}),
			want:     ptr(tomorrow.Add(8 * time.Hour)),
		},
		"at exact start of range: not blocked": {
			rangeSet: rangeSet(user.DailyScreenTimeLimitRange{StartTimeSeconds: 10 * 3600, EndTimeSeconds: 12 * 3600}),
			want:     nil,
		},
		"at exact end of range: blocked": {
			rangeSet: rangeSet(user.DailyScreenTimeLimitRange{StartTimeSeconds: 8 * 3600, EndTimeSeconds: 10 * 3600}),
			want:     ptr(tomorrow.Add(8 * time.Hour)),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got := c.rangeSet.BlockedUntil(now)

			// assert
			if c.want == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, *c.want, *got)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	t.Parallel()

	ptr := func(v int) *int { return &v }

	type arg struct {
		displayName     string
		languageCode    string
		screenTimeLimit *int
	}

	type want struct {
		displayName     user.DisplayName
		languageCode    user.LanguageCode
		screenUnlimited bool
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:  arg{displayName: "testuser", languageCode: "ja", screenTimeLimit: ptr(3600)},
			want: &want{displayName: "testuser", languageCode: "ja", screenUnlimited: false},
		},
		"success unlimited": {
			arg:  arg{displayName: "testuser", languageCode: "ja", screenTimeLimit: nil},
			want: &want{displayName: "testuser", languageCode: "ja", screenUnlimited: true},
		},
		"invalid display name": {
			arg:     arg{displayName: "", languageCode: "ja", screenTimeLimit: nil},
			wantErr: user.ErrDisplayNameTooShort,
		},
		"invalid language code": {
			arg:     arg{displayName: "testuser", languageCode: "invalid", screenTimeLimit: nil},
			wantErr: user.ErrLanguageCodeNotSupported,
		},
		"invalid screen time limit": {
			arg:     arg{displayName: "testuser", languageCode: "ja", screenTimeLimit: ptr(-1)},
			wantErr: user.ErrDailyScreenTimeOutOfRange,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := user.NewUser(c.arg.displayName, c.arg.languageCode, c.arg.screenTimeLimit)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.displayName, got.DisplayName)
				assert.Equal(t, c.want.languageCode, got.LanguageCode)
				assert.Equal(t, c.want.screenUnlimited, got.ScreenTimeLimit.IsUnlimited())
			}
		})
	}
}
