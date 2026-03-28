package v1

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/feed"
	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ StrictServerInterface = (*APIHandler)(nil)

type APIHandler struct {
	db *pgxpool.Pool

	authService     *auth.Service
	userService     *user.Service
	channelService  *channel.Service
	videoService    *video.Service
	playlistService *playlist.Service
	historyService  *history.Service
	feedService     *feed.Service

	serverURL   string
	frontendURL string
}

func NewAPIHandler(
	db *pgxpool.Pool,
	oidcService oidc.GoogleOIDCService,
	serverURL, frontendURL string,
	jwtService jwt_d.Service,
	refreshTokenDuration time.Duration,
	ytService youtube_d.Service,
	rssFetchDuration time.Duration,
	scheduler scheduler.Service,
) *APIHandler {

	return &APIHandler{
		db: db,

		channelService:  channel.NewService(db, ytService, rssFetchDuration),
		videoService:    video.NewService(db),
		playlistService: playlist.NewService(db, ytService),
		authService:     auth.NewService(db, oidcService, ytService, channel.NewService(db, ytService, rssFetchDuration), playlist.NewService(db, ytService), serverURL, jwtService, refreshTokenDuration),
		userService:     user.NewService(db, jwtService, serverURL),
		historyService:  history.NewService(db),
		feedService:     feed.NewService(db, ytService, rssFetchDuration),

		serverURL:   serverURL,
		frontendURL: frontendURL,
	}
}
