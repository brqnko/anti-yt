package playlist

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidPlaylistTitle       = errors.New("invalid playlist title: must be between 1 and 128 characters")
	ErrInvalidPlaylistDescription = errors.New("invalid playlist description: must be at most 255 characters")
	ErrInvalidVisibilityCode      = errors.New("invalid visibility code")
	ErrInvalidPlaylistCode = errors.New("invalid playlist code")
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
		return 0, fmt.Errorf("failed to create visibility code(newVisibilityCode): %w", ErrInvalidVisibilityCode)
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
		return "", fmt.Errorf("failed to create playlist title(newPlaylistTitle): %w", ErrInvalidPlaylistTitle)
	}
	return PlaylistTitle(s), nil
}

func (p PlaylistTitle) String() string {
	return string(p)
}

type PlaylistDescription string

func NewPlaylistDescription(s string) (PlaylistDescription, error) {
	if len(s) > 255 {
		return "", fmt.Errorf("failed to create playlist description(newPlaylistDescription): %w", ErrInvalidPlaylistDescription)
	}
	return PlaylistDescription(s), nil
}

func (p PlaylistDescription) String() string {
	return string(p)
}

func (p *Playlist) SetTitle(s *string) error {
	if s == nil {
		return nil
	}
	t, err := NewPlaylistTitle(*s)
	if err != nil {
		return err
	}
	p.Title = t
	return nil
}

func (p *Playlist) SetDescription(s *string) error {
	if s == nil {
		return nil
	}
	d, err := NewPlaylistDescription(*s)
	if err != nil {
		return err
	}
	p.Description = d
	return nil
}

var ErrNegativeVideoCount = errors.New("video count must not be negative")
var ErrVideoCountUnderflow = errors.New("video count is already 0")

func (p *Playlist) SetVideoCount(count int) error {
	if count < 0 {
		return ErrNegativeVideoCount
	}
	p.VideoCount = count
	return nil
}

func (p *Playlist) IncrementVideoCount() {
	p.VideoCount++
}

func (p *Playlist) DecrementVideoCount() error {
	if p.VideoCount <= 0 {
		return ErrVideoCountUnderflow
	}
	p.VideoCount--
	return nil
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
		return 0, fmt.Errorf("failed to create playlist code(newPlaylistCode): %w", ErrInvalidPlaylistCode)
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
	ID             uuid.UUID
	Title          PlaylistTitle
	Description    PlaylistDescription
	VisibilityCode VisibilityCode
	PlaylistCode   PlaylistCode
	VideoCount     int
	RegisteredAt   time.Time
}

type PlaylistOption func(*Playlist)

func WithPlaylistID(id uuid.UUID) PlaylistOption {
	return func(p *Playlist) {
		p.ID = id
	}
}

func WithPlaylistRegisteredAt(registeredAt time.Time) PlaylistOption {
	return func(p *Playlist) {
		p.RegisteredAt = registeredAt
	}
}

func WithPlaylistVideoCount(count int) PlaylistOption {
	return func(p *Playlist) {
		p.VideoCount = count
	}
}

func NewPlaylist(
	title string,
	description string,
	visibilityStr string,
	playlistTypeStr string,
	opts ...PlaylistOption,
) (*Playlist, error) {
	t, err := NewPlaylistTitle(title)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist title(newPlaylist): %w", err)
	}

	d, err := NewPlaylistDescription(description)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist description(newPlaylist): %w", err)
	}

	v, err := NewVisibilityCode(visibilityStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create visibility code(newPlaylist): %w", err)
	}

	pc, err := NewPlaylistCode(playlistTypeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create playlist code(newPlaylist): %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid v7(newPlaylist): %w", err)
	}

	pl := &Playlist{
		ID:             id,
		Title:          t,
		Description:    d,
		VisibilityCode: v,
		PlaylistCode:   pc,
		VideoCount:     0,
		RegisteredAt:   time.Now().UTC(),
	}

	for _, opt := range opts {
		opt(pl)
	}

	return pl, nil
}
