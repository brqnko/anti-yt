package v1

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ StrictServerInterface = (*APIHandler)(nil)

const (
	internalErrorTitle  = "internal Server Error"
	internalErrorDetail = "Something went wrong!"
)

type APIHandler struct {
	db *pgxpool.Pool

	authService     *auth.Service
	userService     *user.Service
	channelService  *channel.Service
	videoService    *video.Service
	playlistService *playlist.Service
	historyService  *history.Service

	serverURL   string
	frontendURL string
}

func NewAPIHandler(
	db *pgxpool.Pool,
	oidcService oidc.GoogleOIDCService,
	serverURL, frontendURL string,
	jwtService jwt_d.JWTService,
	refreshTokenDuration time.Duration,
	ytService youtube_d.YouTubeAPIService,
	rssFetchDuration time.Duration,
) (*APIHandler, error) {
	authService, err := auth.NewService(db, oidcService, serverURL, jwtService, refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	userService, err := user.NewService(db, jwtService, serverURL)
	if err != nil {
		return nil, err
	}

	channelService, err := channel.NewService(db, ytService, rssFetchDuration)
	if err != nil {
		return nil, err
	}

	videoService, err := video.NewService(db, ytService)
	if err != nil {
		return nil, err
	}

	playlistService, err := playlist.NewService(db, ytService)
	if err != nil {
		return nil, err
	}

	historyService, err := history.NewService(db)
	if err != nil {
		return nil, err
	}

	return &APIHandler{
		db: db,

		authService:     authService,
		userService:     userService,
		channelService:  channelService,
		videoService:    videoService,
		playlistService: playlistService,
		historyService:  historyService,

		serverURL:   serverURL,
		frontendURL: frontendURL,
	}, nil
}
