package v1

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/database"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

var _ ServerInterface = (*Handler)(nil)

type Handler struct {
	db *sql.DB
}

func NewHandler() (*Handler, error) {
	db, err := database.ConnectDB()
	if err != nil {
		return nil, err
	}
	if err = database.RunMigration(db); err != nil {
		return nil, err
	}

	return &Handler{db: db}, nil
}

func (h *Handler) Close(ctx context.Context) error {
	if err := h.db.Close(); err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetAuthGoogle(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthLogout(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetChannelsChannelIdVideos(ctx echo.Context, channelId openapitypes.UUID, params GetChannelsChannelIdVideosParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetFeed(ctx echo.Context, params GetFeedParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetFeedChannels(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetHistory(ctx echo.Context, params GetHistoryParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetPlaylists(ctx echo.Context, params GetPlaylistsParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostPlaylists(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeletePlaylistsPlaylistId(ctx echo.Context, playlistId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetPlaylistsPlaylistId(ctx echo.Context, playlistId openapitypes.UUID, params GetPlaylistsPlaylistIdParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PatchPlaylistsPlaylistId(ctx echo.Context, playlistId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeletePlaylistsPlaylistIdVideos(ctx echo.Context, playlistId openapitypes.UUID, params DeletePlaylistsPlaylistIdVideosParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostPlaylistsPlaylistIdVideos(ctx echo.Context, playlistId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetSearch(ctx echo.Context, params GetSearchParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetStatisticsDaily(ctx echo.Context, params GetStatisticsDailyParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetStatisticsMonthly(ctx echo.Context, params GetStatisticsMonthlyParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetSubscriptions(ctx echo.Context, params GetSubscriptionsParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostSubscriptions(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeleteSubscriptionsChannelId(ctx echo.Context, channelId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeleteUsersMe(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetUsersMeStatus(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PatchUsersMeStatus(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostUsersMe(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetUsersMeLimits(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetUsersMeSessions(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeleteUsersMeSessionsSessionId(ctx echo.Context, sessionId openapitypes.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetVideosVideoId(ctx echo.Context, externalVideoId string) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostVideosVideoIdHeartbeats(ctx echo.Context, externalVideoId string) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthGoogle(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetAuthGoogleCallback(ctx echo.Context, params GetAuthGoogleCallbackParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthRefresh(ctx echo.Context, params PostAuthRefreshParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "OK")
}
