package playlist

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

var (
	ErrInvalidPlaylistTitle       = core.NewDomainError("playlist.invalid_title", "invalid playlist title: must be between 1 and 128 characters")
	ErrInvalidPlaylistDescription = core.NewDomainError("playlist.invalid_description", "invalid playlist description: must be at most 255 characters")
	ErrInvalidVisibilityCode      = core.NewDomainError("playlist.invalid_visibility_code", "invalid visibility code")
	ErrInvalidPlaylistCode        = core.NewDomainError("playlist.invalid_playlist_code", "invalid playlist code")
	ErrPlaylistNotModifiable      = core.NewDomainError("playlist.not_modifiable", "this playlist cannot be modified")
)

type VisibilityCode int

var visibilityCodeMap = []struct {
	code VisibilityCode
	str  string
}{
	{code: 0, str: "private"},
	{code: 1, str: "public"},
}

func NewVisibilityCode(s string) (_ VisibilityCode, err error) {
	defer util.Wrap(&err, "playlist.NewVisibilityCode")

	for _, c := range visibilityCodeMap {
		if s == c.str {
			return c.code, nil
		}
	}

	return 0, ErrInvalidVisibilityCode
}

func (v VisibilityCode) String() string {
	for _, c := range visibilityCodeMap {
		if c.code == v {
			return c.str
		}
	}

	return "private"
}

type PlaylistTitle string

func NewPlaylistTitle(s string) (_ PlaylistTitle, err error) {
	defer util.Wrap(&err, "playlist.NewPlaylistTitle")

	if len(s) == 0 || len(s) > 128 {
		return "", ErrInvalidPlaylistTitle
	}
	return PlaylistTitle(s), nil
}

func (p PlaylistTitle) String() string {
	return string(p)
}

type PlaylistDescription string

func NewPlaylistDescription(s string) (_ PlaylistDescription, err error) {
	defer util.Wrap(&err, "playlist.NewPlaylistDescription")

	if len(s) > 255 {
		return "", ErrInvalidPlaylistDescription
	}
	return PlaylistDescription(s), nil
}

func (p PlaylistDescription) String() string {
	return string(p)
}

func (p *Playlist) IsModifiable() bool {
	return p.PlaylistCode == 0 // normal
}

func (p *Playlist) SetTitle(s *string) (err error) {
	if s == nil {
		return nil
	}
	defer util.Wrap(&err, "playlist.(*Playlist).SetTitle")
	t, err := NewPlaylistTitle(*s)
	if err != nil {
		return err
	}
	p.Title = t
	return nil
}

func (p *Playlist) SetDescription(s *string) (err error) {
	if s == nil {
		return nil
	}
	defer util.Wrap(&err, "playlist.(*Playlist).SetDescription")
	d, err := NewPlaylistDescription(*s)
	if err != nil {
		return err
	}
	p.Description = d
	return nil
}

var ErrNegativeVideoCount = core.NewDomainError("playlist.negative_video_count", "video count must not be negative")
var ErrVideoCountUnderflow = core.NewDomainError("playlist.video_count_underflow", "video count is already 0")

func (p *Playlist) SetVideoCount(count int) (err error) {
	defer util.Wrap(&err, "playlist.(*Playlist).SetVideoCount")
	if count < 0 {
		return ErrNegativeVideoCount
	}
	p.VideoCount = count
	return nil
}

func (p *Playlist) IncrementVideoCount() {
	p.VideoCount++
}

func (p *Playlist) DecrementVideoCount() (err error) {
	defer util.Wrap(&err, "playlist.(*Playlist).DecrementVideoCount")
	if p.VideoCount <= 0 {
		return ErrVideoCountUnderflow
	}
	p.VideoCount--
	return nil
}

type PlaylistCode int

var playlistCodeMap = []struct {
	code PlaylistCode
	str  string
}{
	{code: 0, str: "normal"},
	{code: 1, str: "external_auto"},
	{code: 2, str: "watch_later"},
}

func NewPlaylistCode(s string) (_ PlaylistCode, err error) {
	defer util.Wrap(&err, "playlist.NewPlaylistCode")

	for _, c := range playlistCodeMap {
		if s == c.str {
			return c.code, nil
		}
	}

	return 0, ErrInvalidPlaylistCode
}

func (p PlaylistCode) String() string {
	for _, c := range playlistCodeMap {
		if c.code == p {
			return c.str
		}
	}

	return "normal"
}

type Playlist struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	ChannelID      *uuid.UUID
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

func WithPlaylistChannelID(channelID uuid.UUID) PlaylistOption {
	return func(p *Playlist) {
		p.ChannelID = &channelID
	}
}

func NewPlaylist(
	userID uuid.UUID,
	title string,
	description string,
	visibilityStr string,
	playlistTypeStr string,
	opts ...PlaylistOption,
) (_ *Playlist, err error) {
	defer util.Wrap(&err, "playlist.NewPlaylist")

	t, err := NewPlaylistTitle(title)
	if err != nil {
		return nil, err
	}

	d, err := NewPlaylistDescription(description)
	if err != nil {
		return nil, err
	}

	v, err := NewVisibilityCode(visibilityStr)
	if err != nil {
		return nil, err
	}

	pc, err := NewPlaylistCode(playlistTypeStr)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	pl := &Playlist{
		ID:             id,
		UserID:         userID,
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
