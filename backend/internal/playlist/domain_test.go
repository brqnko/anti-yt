package playlist_test

import (
	"strings"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewVisibilityCode(t *testing.T) {
	t.Parallel()

	type arg struct {
		str string
	}

	type want struct {
		code playlist.VisibilityCode
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"private": {
			arg:     arg{str: "private"},
			want:    &want{code: 0},
			wantErr: nil,
		},
		"public": {
			arg:     arg{str: "public"},
			want:    &want{code: 1},
			wantErr: nil,
		},
		"invalid": {
			arg:     arg{str: "invalid"},
			want:    nil,
			wantErr: playlist.ErrInvalidVisibilityCode,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := playlist.NewVisibilityCode(c.arg.str)

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

func TestNewPlaylistTitle(t *testing.T) {
	t.Parallel()

	type arg struct {
		title string
	}

	type want struct {
		title playlist.PlaylistTitle
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{title: "my playlist"},
			want:    &want{title: playlist.PlaylistTitle("my playlist")},
			wantErr: nil,
		},
		"128 chars": {
			arg:     arg{title: strings.Repeat("a", 128)},
			want:    &want{title: playlist.PlaylistTitle(strings.Repeat("a", 128))},
			wantErr: nil,
		},
		"empty": {
			arg:     arg{title: ""},
			want:    nil,
			wantErr: playlist.ErrInvalidPlaylistTitle,
		},
		"129 chars": {
			arg:     arg{title: strings.Repeat("a", 129)},
			want:    nil,
			wantErr: playlist.ErrInvalidPlaylistTitle,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := playlist.NewPlaylistTitle(c.arg.title)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.title, got)
			}
		})
	}
}

func TestNewPlaylistDescription(t *testing.T) {
	t.Parallel()

	type arg struct {
		description string
	}

	type want struct {
		description playlist.PlaylistDescription
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{description: "my description"},
			want:    &want{description: playlist.PlaylistDescription("my description")},
			wantErr: nil,
		},
		"empty": {
			arg:     arg{description: ""},
			want:    &want{description: playlist.PlaylistDescription("")},
			wantErr: nil,
		},
		"255 chars": {
			arg:     arg{description: strings.Repeat("a", 255)},
			want:    &want{description: playlist.PlaylistDescription(strings.Repeat("a", 255))},
			wantErr: nil,
		},
		"256 chars": {
			arg:     arg{description: strings.Repeat("a", 256)},
			want:    nil,
			wantErr: playlist.ErrInvalidPlaylistDescription,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := playlist.NewPlaylistDescription(c.arg.description)

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

func TestPlaylist_IsModifiable(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		playlistCode string
		want         bool
	}{
		"normal is modifiable": {
			playlistCode: "normal",
			want:         true,
		},
		"external_auto is not modifiable": {
			playlistCode: "external_auto",
			want:         false,
		},
		"watch_later is not modifiable": {
			playlistCode: "watch_later",
			want:         false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange
			p, _ := playlist.NewPlaylist(uuid.Must(uuid.NewV7()), "title", "", "private", c.playlistCode)

			// act
			got := p.IsModifiable()

			// assert
			assert.Equal(t, c.want, got)
		})
	}
}

func TestPlaylist_SetVideoCount(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		count   int
		want    int
		wantErr error
	}{
		"success": {
			count:   5,
			want:    5,
			wantErr: nil,
		},
		"zero": {
			count:   0,
			want:    0,
			wantErr: nil,
		},
		"negative": {
			count:   -1,
			wantErr: playlist.ErrNegativeVideoCount,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange
			p, _ := playlist.NewPlaylist(uuid.Must(uuid.NewV7()), "title", "", "private", "normal")

			// act
			err := p.SetVideoCount(c.count)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want, p.VideoCount)
			}
		})
	}
}

func TestPlaylist_DecrementVideoCount(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		initialCount int
		want         int
		wantErr      error
	}{
		"success": {
			initialCount: 1,
			want:         0,
			wantErr:      nil,
		},
		"underflow": {
			initialCount: 0,
			wantErr:      playlist.ErrVideoCountUnderflow,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange
			p, _ := playlist.NewPlaylist(uuid.Must(uuid.NewV7()), "title", "", "private", "normal")
			_ = p.SetVideoCount(c.initialCount)

			// act
			err := p.DecrementVideoCount()

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want, p.VideoCount)
			}
		})
	}
}

func TestNewPlaylist(t *testing.T) {
	t.Parallel()

	userID := uuid.Must(uuid.NewV7())

	type arg struct {
		userID        uuid.UUID
		title         string
		description   string
		visibilityStr string
		playlistType  string
	}

	type want struct {
		userID         uuid.UUID
		title          playlist.PlaylistTitle
		description    playlist.PlaylistDescription
		visibilityCode playlist.VisibilityCode
		videoCount     int
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:  arg{userID: userID, title: "my playlist", description: "desc", visibilityStr: "public", playlistType: "normal"},
			want: &want{userID: userID, title: "my playlist", description: "desc", visibilityCode: 1, videoCount: 0},
		},
		"invalid title": {
			arg:     arg{userID: userID, title: "", description: "desc", visibilityStr: "public", playlistType: "normal"},
			wantErr: playlist.ErrInvalidPlaylistTitle,
		},
		"invalid description": {
			arg:     arg{userID: userID, title: "title", description: strings.Repeat("a", 256), visibilityStr: "public", playlistType: "normal"},
			wantErr: playlist.ErrInvalidPlaylistDescription,
		},
		"invalid visibility": {
			arg:     arg{userID: userID, title: "title", description: "", visibilityStr: "invalid", playlistType: "normal"},
			wantErr: playlist.ErrInvalidVisibilityCode,
		},
		"invalid playlist type": {
			arg:     arg{userID: userID, title: "title", description: "", visibilityStr: "public", playlistType: "invalid"},
			wantErr: playlist.ErrInvalidPlaylistCode,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := playlist.NewPlaylist(c.arg.userID, c.arg.title, c.arg.description, c.arg.visibilityStr, c.arg.playlistType)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.userID, got.UserID)
				assert.Equal(t, c.want.title, got.Title)
				assert.Equal(t, c.want.description, got.Description)
				assert.Equal(t, c.want.visibilityCode, got.VisibilityCode)
				assert.Equal(t, c.want.videoCount, got.VideoCount)
			}
		})
	}
}

func TestNewPlaylistCode(t *testing.T) {
	t.Parallel()

	type arg struct {
		str string
	}

	type want struct {
		code playlist.PlaylistCode
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"normal": {
			arg:     arg{str: "normal"},
			want:    &want{code: 0},
			wantErr: nil,
		},
		"external_auto": {
			arg:     arg{str: "external_auto"},
			want:    &want{code: 1},
			wantErr: nil,
		},
		"watch_later": {
			arg:     arg{str: "watch_later"},
			want:    &want{code: 2},
			wantErr: nil,
		},
		"invalid": {
			arg:     arg{str: "invalid"},
			want:    nil,
			wantErr: playlist.ErrInvalidPlaylistCode,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := playlist.NewPlaylistCode(c.arg.str)

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
