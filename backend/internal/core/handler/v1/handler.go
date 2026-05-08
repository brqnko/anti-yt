package v1

import (
	"time"

	"github.com/google/uuid"

	"github.com/brqnko/anti-yt/backend/internal/auth"
	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/feed"
	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/jackc/pgx/v5/pgxpool"
)

func cursorToUUID(b *util.Base64UUID) *uuid.UUID {
	if b == nil {
		return nil
	}
	u := b.UUID()
	return &u
}

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
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	youtubeClient youtube_d.Client,
	rssFetchDuration time.Duration,
	scheduler scheduler.Service,
	jtiBlacklistRepo database_d.JtiBlacklistRepository,
	feedRepo database_d.FeedRepository,
) *APIHandler {

	return new(APIHandler{
		db: db,

		channelService:  channel.NewService(db, youtubeClient, feedRepo, rssFetchDuration),
		videoService:    video.NewService(db),
		playlistService: playlist.NewService(db, youtubeClient, feedRepo),
		authService:     auth.NewService(db, oidcService, youtubeClient, channel.NewService(db, youtubeClient, feedRepo, rssFetchDuration), playlist.NewService(db, youtubeClient, feedRepo), serverURL, jwtService, accessTokenDuration, refreshTokenDuration, jtiBlacklistRepo),
		userService:     user.NewService(db, jwtService, serverURL, accessTokenDuration, jtiBlacklistRepo),
		historyService:  history.NewService(db, feedRepo),
		feedService:     feed.NewService(db, youtubeClient, feedRepo, rssFetchDuration),

		serverURL:   serverURL,
		frontendURL: frontendURL,
	})
}
