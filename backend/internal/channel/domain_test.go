package channel_test

import (
	"strings"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/stretchr/testify/assert"
)

func TestNewChannel(t *testing.T) {
	t.Parallel()

	type arg struct {
		fetchedAt    time.Time
		rssFetchedAt time.Time
		channel      youtube_d.Channel
		opts         []channel.ChannelOption
	}

	type want struct {
		fetchedAt    time.Time
		rssFetchedAt time.Time
		channel      youtube_d.Channel
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg: arg{
				fetchedAt:    time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
				rssFetchedAt: time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC),
				channel: youtube_d.Channel{
					ID:                "UCrtfvygbhunjkcrtfvygbhu",
					DisplayName:       "crtfvygbuhnijmok",
					CustomID:          "vtfygbhjknlm",
					Description:       "frtygioubgytvfrvygbuijmo",
					IconURL:           "rtfvygbuhinjmonhubgyvftcrvgybhujk",
					SubscribersCount:  3232,
					UploadsPlaylistID: "ctrvfygbhunij",
					CreatedAt:         time.Time{},
				},
				opts: []channel.ChannelOption{},
			},
			want: &want{
				fetchedAt:    time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
				rssFetchedAt: time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC),
				channel: youtube_d.Channel{
					ID:                "UCrtfvygbhunjkcrtfvygbhu",
					DisplayName:       "crtfvygbuhnijmok",
					CustomID:          "vtfygbhjknlm",
					Description:       "frtygioubgytvfrvygbuijmo",
					IconURL:           "rtfvygbuhinjmonhubgyvftcrvgybhujk",
					SubscribersCount:  3232,
					UploadsPlaylistID: "ctrvfygbhunij",
					CreatedAt:         time.Time{},
				},
			},
			wantErr: nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange

			// act
			got, err := channel.NewChannel(c.arg.fetchedAt, c.arg.rssFetchedAt, c.arg.channel, c.arg.opts...)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.channel, got.Channel)
				assert.Equal(t, c.want.fetchedAt, got.FetchedAt)
				assert.Equal(t, c.want.rssFetchedAt, got.RSSFetchedAt)
			}
		})
	}
}

func TestChannel_ShouldRSSFetchFeed(t *testing.T) {
	t.Parallel()

	type arg struct {
		rssFetchedAt  time.Time
		fetchDuration time.Duration
	}

	cases := map[string]struct {
		arg  arg
		want bool
	}{
		"should fetch": {
			arg: arg{
				rssFetchedAt:  time.Now().UTC().Add(-2 * time.Hour),
				fetchDuration: time.Hour,
			},
			want: true,
		},
		"should not fetch": {
			arg: arg{
				rssFetchedAt:  time.Now().UTC().Add(-30 * time.Minute),
				fetchDuration: time.Hour,
			},
			want: false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ch := new(channel.Channel{RSSFetchedAt: c.arg.rssFetchedAt})

			// act
			got := ch.ShouldFetchRSSFeed(c.arg.fetchDuration)

			// assert
			assert.Equal(t, c.want, got)
		})
	}
}

func TestNewValuableCategoryCode(t *testing.T) {
	t.Parallel()

	type arg struct {
		str string
	}

	type want struct {
		code channel.ValuableCategoryCode
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"unknown": {
			arg:     arg{str: "unknown"},
			want:    &want{code: 0},
			wantErr: nil,
		},
		"learn_deepen": {
			arg:     arg{str: "learn_deepen"},
			want:    &want{code: 1},
			wantErr: nil,
		},
		"invalid": {
			arg:     arg{str: "invalid"},
			want:    nil,
			wantErr: channel.ErrInvalidValuableCategoryCode,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := channel.NewValuableCategoryCode(c.arg.str)

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

func TestNewValuableDescription(t *testing.T) {
	t.Parallel()

	type arg struct {
		description string
	}

	type want struct {
		description channel.ValuableDescription
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{description: "valid description"},
			want:    &want{description: channel.ValuableDescription("valid description")},
			wantErr: nil,
		},
		"too long": {
			arg:     arg{description: strings.Repeat("a", 256)},
			want:    nil,
			wantErr: channel.ErrInvalidValuableDescription,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := channel.NewValuableDescription(c.arg.description)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.description, got)
			}
		})
	}
}
