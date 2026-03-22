package playlist

import (
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
)

var (
	ErrInvalidPlaylistTitle       = errors.New("invalid playlist title: must be between 1 and 128 characters")
	ErrInvalidPlaylistDescription = errors.New("invalid playlist description: must be at most 255 characters")
	ErrInvalidVisibilityCode      = errors.New("invalid visibility code")
	ErrInvalidPlaylistCode        = errors.New("invalid playlist code")
	ErrPlaylistNotFound           = errors.New("playlist not found")
	ErrVideoNotInPlaylist         = errors.New("video not in playlist")
)

type VisibilityCode int

const (
	VisibilityPrivate VisibilityCode = 0
)

func NewVisibilityCode(s string) (VisibilityCode, error) {
	switch s {
	case "private":
		return VisibilityPrivate, nil
	default:
		return 0, ErrInvalidVisibilityCode
	}
}

func (v VisibilityCode) String() string {
	switch v {
	case VisibilityPrivate:
		return "private"
	default:
		return "private"
	}
}

type PlaylistTitle string

func NewPlaylistTitle(s string) (PlaylistTitle, error) {
	if len(s) == 0 || len(s) > 128 {
		return "", ErrInvalidPlaylistTitle
	}
	return PlaylistTitle(s), nil
}

type PlaylistDescription string

func NewPlaylistDescription(s string) (PlaylistDescription, error) {
	if len(s) > 255 {
		return "", ErrInvalidPlaylistDescription
	}
	return PlaylistDescription(s), nil
}

type PlaylistCode int

const (
	PlaylistCodeNormal PlaylistCode = 0
)

func NewPlaylistCode(s string) (PlaylistCode, error) {
	switch s {
	case "normal":
		return PlaylistCodeNormal, nil
	default:
		return 0, ErrInvalidPlaylistCode
	}
}

func (p PlaylistCode) String() string {
	switch p {
	case PlaylistCodeNormal:
		return "normal"
	default:
		return "normal"
	}
}

type Playlist struct {
	ID                   uuid.UUID
	Title                PlaylistTitle
	Description          PlaylistDescription
	VisibilityCode       VisibilityCode
	PlaylistCode         PlaylistCode
	VideoCount           int
	CreatedAt            time.Time
	TopVideoThumbnailURL string
	Videos               []video.Video
}

func (p Playlist) WithGeneratedFields(id uuid.UUID, createdAt time.Time) Playlist {
	p.ID = id
	p.CreatedAt = createdAt
	return p
}

func NewPlaylist(
	id uuid.UUID,
	title string,
	description string,
	visibilityStr string,
	playlistTypeStr string,
	videoCount int,
	createdAt time.Time,
	topVideoThumbnailUrl string,
	videos []video.Video,
) (Playlist, error) {
	t, err := NewPlaylistTitle(title)
	if err != nil {
		return Playlist{}, err
	}

	d, err := NewPlaylistDescription(description)
	if err != nil {
		return Playlist{}, err
	}

	v, err := NewVisibilityCode(visibilityStr)
	if err != nil {
		return Playlist{}, err
	}

	p, err := NewPlaylistCode(playlistTypeStr)
	if err != nil {
		return Playlist{}, err
	}

	return Playlist{
		ID:                   id,
		Title:                t,
		Description:          d,
		VisibilityCode:       v,
		PlaylistCode:         p,
		VideoCount:           videoCount,
		CreatedAt:            createdAt,
		TopVideoThumbnailURL: topVideoThumbnailUrl,
		Videos:               videos,
	}, nil
}
