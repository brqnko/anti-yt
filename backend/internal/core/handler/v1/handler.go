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
	internalErrorTitle  = "Internal Server Error"
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
	ytService youtube_d.Service,
	rssFetchDuration time.Duration,
) *APIHandler {
	channelService := channel.NewService(db, ytService, rssFetchDuration)

	return &APIHandler{
		db: db,

		authService:     auth.NewService(db, oidcService, serverURL, jwtService, refreshTokenDuration),
		userService:     user.NewService(db, jwtService, serverURL),
		channelService:  channelService,
		videoService:    video.NewService(db, ytService, channelService),
		playlistService: playlist.NewService(db, ytService),
		historyService:  history.NewService(db),

		serverURL:   serverURL,
		frontendURL: frontendURL,
	}
}
