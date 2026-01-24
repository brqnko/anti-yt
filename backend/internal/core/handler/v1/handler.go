package v1

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/brqnko/anti-yt/backend/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/pressly/goose/v3"
)

var _ ServerInterface = (*Handler)(nil)

type Handler struct {
	db *sql.DB
}

func (h *Handler) GetAuthGoogle(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostAuthLogout(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetChannelsChannelIdVideos(ctx echo.Context, channelId openapi_types.UUID, params GetChannelsChannelIdVideosParams) error {
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

func (h *Handler) DeletePlaylistsPlaylistId(ctx echo.Context, playlistId openapi_types.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GetPlaylistsPlaylistId(ctx echo.Context, playlistId openapi_types.UUID, params GetPlaylistsPlaylistIdParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PatchPlaylistsPlaylistId(ctx echo.Context, playlistId openapi_types.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DeletePlaylistsPlaylistIdVideos(ctx echo.Context, playlistId openapi_types.UUID, params DeletePlaylistsPlaylistIdVideosParams) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) PostPlaylistsPlaylistIdVideos(ctx echo.Context, playlistId openapi_types.UUID) error {
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

func (h *Handler) DeleteSubscriptionsChannelId(ctx echo.Context, channelId openapi_types.UUID) error {
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

func (h *Handler) DeleteUsersMeSessionsSessionId(ctx echo.Context, sessionId openapi_types.UUID) error {
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

func RunMigration(db *sql.DB) error {
	fmt.Println("running migration")
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// "migrations" は go:embed で指定したディレクトリ名
	if err := goose.Up(db, "."); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Printf("Error: %s, Detail: %s, Hint: %s\n", pgErr.Message, pgErr.Detail, pgErr.Hint)
		}
		return err
	}

	fmt.Println("migration ok")
	return nil
}

func NewHandler() *Handler {
	data, err := os.ReadFile("/run/secrets/db-password")
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://postgres:%s@db:5432/example?sslmode=disable", strings.TrimSpace(string(data))))
	if err != nil {
		log.Fatal(err)
	}
	if err = RunMigration(db); err != nil {
		log.Fatal(err)
	}

	return &Handler{db: db}
}

func (h *Handler) GetHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "OK")
}
