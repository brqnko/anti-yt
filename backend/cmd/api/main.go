package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/admin"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/llm"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/core/otel_d"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/job"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	os.Exit(run(context.Background()))
}

type config struct {
	env                    string
	oidcGoogleClientID     string
	oidcGoogleClientSecret string
	serverURL              string
	frontendURL            string
	youtubeDataAPIKey      string
	adminAPIKey            string
	geminiAPIKey           string
	discordWebhookURL      string

	port int

	databaseURL string
	redisURL    string
}

func run(ctx context.Context) int {
	jwtPrivate, err := loadPrivateKey("/run/secrets/jwt-private")
	if err != nil {
		fmt.Printf("failed to load jwt-private: %v\n", err)
		return 1
	}
	jwtPublic, err := loadPublicKey("/run/secrets/jwt-public")
	if err != nil {
		fmt.Printf("failed to load jwt-public: %v\n", err)
		return 1
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Printf("invalid or missing PORT: %v\n", err)
		return 1
	}
	cfg := config{
		env:                    os.Getenv("ENV"),
		oidcGoogleClientID:     os.Getenv("OIDC_GOOGLE_CLIENT_ID"),
		oidcGoogleClientSecret: os.Getenv("OIDC_GOOGLE_CLIENT_SECRET"),
		databaseURL:            os.Getenv("DATABASE_URL"),
		serverURL:              os.Getenv("SERVER_URL"),
		frontendURL:            os.Getenv("FRONTEND_URL"),
		youtubeDataAPIKey:      os.Getenv("YOUTUBE_DATA_API_KEY"),
		adminAPIKey:            os.Getenv("ADMIN_API_KEY"),
		geminiAPIKey:           os.Getenv("GEMINI_API_KEY"),
		discordWebhookURL:      os.Getenv("DISCORD_WEBHOOK_URL"),
		redisURL:               os.Getenv("REDIS_URL"),
		port:                   port,
	}

	var handler slog.Handler
	if cfg.env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, new(slog.HandlerOptions{
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if id, ok := a.Value.Any().(uuid.UUID); ok {
					return slog.String(a.Key, base64.RawURLEncoding.EncodeToString(id[:]))
				}
				return a
			},
		}))
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	otelShutdown, err := otel_d.Setup(initCtx, "anti-yt-backend", cfg.env)
	if err != nil {
		slog.Error("failed to setup otel", slog.Any("error", err))
		return 1
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown otel", slog.Any("error", err))
		}
	}()

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	if err = database_d.RunMigration(initCtx, cfg.databaseURL); err != nil {
		slog.Error("failed to run migration", slog.Any("error", err))
		return 1
	}
	slog.Info("migration ok")

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	db, err := database_d.ConnectPostgres(initCtx, cfg.databaseURL)
	if err != nil {
		slog.Error("failed to connect db", slog.Any("error", err))
		return 1
	}
	defer db.Close()

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	redisClient, err := database_d.ConnectRedis(initCtx, cfg.redisURL)
	if err != nil {
		slog.Error("failed to connect redis", slog.Any("error", err))
		return 1
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("failed to close redis", slog.Any("error", err))
		}
	}()
	jtiBlacklistRepo := database_d.NewInMemoryJtiBlacklistRepository()

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	oidcClient, err := oidc.NewGoogleClient(initCtx, cfg.oidcGoogleClientID, cfg.oidcGoogleClientSecret, fmt.Sprintf("%s/v1/auth/google/callback", cfg.serverURL))
	if err != nil {
		slog.Error("failed to create oidc client", slog.Any("error", err))
		return 1
	}

	accessTokenDuration := 30 * time.Minute
	jwtService := jwt_d.NewService(jwtPrivate, jwtPublic, accessTokenDuration, cfg.serverURL)

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	youtubeClient, err := youtube_d.NewClient(initCtx, cfg.youtubeDataAPIKey, cfg.oidcGoogleClientID, cfg.oidcGoogleClientSecret, fmt.Sprintf("%s/v1/auth/oauth/youtube/callback", cfg.serverURL))
	if err != nil {
		slog.Error("failed to create youtube client", slog.Any("error", err))
		return 1
	}
	slog.Info("youtube api ok")

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	llmClient, err := llm.NewGemini(initCtx, cfg.geminiAPIKey, "gemini-2.5-flash-lite")
	if err != nil {
		slog.Error("failed to create llm client", slog.Any("error", err))
		return 1
	}
	slog.Info("gemini api ok")

	scheduler := scheduler.NewService()
	if err := scheduler.AddFunc("0 0 * * *", job.NewLLMSummaryJob(db, llmClient, discord_d.NewClient(cfg.discordWebhookURL))); err != nil {
		slog.Error("failed to setup llm summary job", slog.Any("error", err))
		return 1
	}
	if err := scheduler.AddFunc("0 0 * * *", job.NewPurgeLeftUserJob(db, database_d.NewFeedRepository(redisClient, 1000, sqlc.New(db)), discord_d.NewClient(cfg.discordWebhookURL))); err != nil {
		slog.Error("failed to setup purge left user job", slog.Any("error", err))
		return 1
	}
	exhaustQuotaJob := job.NewExhaustQuotaJob(db, youtubeClient, database_d.NewFeedRepository(redisClient, 1000, sqlc.New(db)), discord_d.NewClient(cfg.discordWebhookURL))
	if err := scheduler.AddFunc("0 6 * * *", exhaustQuotaJob); err != nil { // 夏
		slog.Error("failed to setup exhaust quota job (PDT)", slog.Any("error", err))
		return 1
	}
	if err := scheduler.AddFunc("0 7 * * *", exhaustQuotaJob); err != nil { // 冬
		slog.Error("failed to setup exhaust quota job (PST)", slog.Any("error", err))
		return 1
	}
	go exhaustQuotaJob.Run()
	if err := scheduler.AddFunc("0 * * * *", job.NewRefillFeedJob(db, database_d.NewFeedRepository(redisClient, 1000, sqlc.New(db)))); err != nil {
		slog.Error("failed to setup refill feed job", slog.Any("error", err))
		return 1
	}

	if err := scheduler.AddFunc("0 0 * * *", job.NewAuthorizationReportJob(db, discord_d.NewClient(cfg.discordWebhookURL))); err != nil {
		slog.Error("failed to setup authorization report job", slog.Any("error", err))
		return 1
	}

	r := chi.NewRouter()
	// NOTE: RateLimit, gzip, loggerはLB(pingora, nginx等)で行う。各リクエストへのレートリミットはmiddlewareが行う
	// r.Useは上に書いた順から実行される
	r.Use(otelchi.Middleware("anti-yt-backend", otelchi.WithChiRoutes(r)))
	r.Use(middleware.Recoverer)
	if cfg.env != "production" {
		r.Use(v1.SwaggerMiddleware())
	}
	admin.HandleAdminEndpoints(r, db, youtubeClient, database_d.NewFeedRepository(redisClient, 1000, sqlc.New(db)), cfg.adminAPIKey)
	v1.HandlerFromMux(v1.NewStrictHandler(
		v1.NewAPIHandler(db, oidcClient, cfg.serverURL, cfg.frontendURL, jwtService, accessTokenDuration, 30*24*time.Hour, youtubeClient, 1*time.Hour, scheduler, jtiBlacklistRepo, database_d.NewFeedRepository(redisClient, 1000, sqlc.New(db))),
		// StrictMiddlewareは下に書いた順から実行される
		[]v1.StrictMiddlewareFunc{
			middleware_d.ResponseCookieMiddleware,
			middleware_d.ScreenTimeMiddleware(db, map[string]struct{}{
				"GetAuthGoogle":         {},
				"GetAuthGoogleCallback": {},
				"PostAuthLogout":        {},
				"PostAuthRefresh":       {},
				"PostUsersMe":           {},
				"GetUsersMeStatus":      {},
				"PatchUsersMeStatus":    {},
				"GetHealth":             {},
			}),
			// WrapErrorMiddlewareはScreenTimeMiddlewareが返すDomainErrorを
			// 捕捉する必要があるため、ScreenTimeより外側(=下)に置く。
			// またSlogMiddlewareがctxに入れるloggerを使うため、Slogより内側(=上)に置く。
			middleware_d.WrapErrorMiddleware(discord_d.NewClient(cfg.discordWebhookURL)),
			middleware_d.CsrfMiddleware(map[string]struct{}{
				"GetAuthGoogle":         {},
				"GetAuthGoogleCallback": {},
			}),
			middleware_d.AuthTokensMiddleware(map[string]struct{}{
				"PostAuthLogout":     {},
				"PostAuthRefresh":    {},
				"PostAuthReactivate": {},
				"PostUsersMe":        {},
			}),
			middleware_d.UserRatelimitMiddleware(database_d.NewRatelimitRepository(redisClient, 24*time.Hour), 2000, map[string]int{
				"PostChannelsSubscribe":         3,
				"GetSearch":                     100,
				"PostPlaylists":                 100,
				"PostPlaylistsPlaylistIdVideos": 3,
				"GetAuthOauthYoutubeCallback":   250,
			}),
			middleware_d.TimezoneMiddleware(map[string]struct{}{
				"GetAuthGoogle":               {},
				"GetAuthGoogleCallback":       {},
				"PostAuthLogout":              {},
				"PostAuthRefresh":             {},
				"GetHealth":                   {},
				"GetAuthOauthYoutube":         {},
				"GetAuthOauthYoutubeCallback": {},
			}),
			// slogはuser_id(AccessTokenMiddlewareで付与), request_idをcontextに必要としているためここに配置
			middleware_d.SlogMiddleware,
			middleware_d.AccessTokenMiddleware(jwtService, jtiBlacklistRepo,
				// optional: クッキーなしかつrefresh_tokenなしは匿名通過、失効時は401（フロント自動リフレッシュ起動）
				map[string]struct{}{
					"GetChannelsChannelId":          {},
					"GetChannelsChannelIdVideos":    {},
					"GetChannelsChannelIdPlaylists": {},
					"GetVideosVideoId":              {},
					"GetPlaylistsPlaylistId":        {},
					"GetPlaylistsPlaylistIdVideos":  {},
					"GetFeed":                       {},
				},
				// public: ブラウザが直接遷移するリダイレクト型の認証フロー。
				// access_tokenの有無/失効に関わらず常に匿名通過（JSON 401を返さない）
				map[string]struct{}{
					"GetAuthGoogle":               {},
					"GetAuthGoogleCallback":       {},
					"PostAuthRefresh":             {},
					"GetAuthOauthYoutubeCallback": {},
				},
				// bypass: register tokenを受け取るため、検証失敗でも常に通過
				map[string]struct{}{
					"PostUsersMe":        {},
					"PostAuthReactivate": {},
				},
			),
			middleware_d.RequestIDMiddleware,
		}), r)

	srv := new(http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: otelhttp.NewHandler(r, "http.server"),
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("shutting down the server", slog.Any("error", err))
		}
	}()

	slog.Info("listening")

	<-ctx.Done()

	slog.Info("graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", slog.Any("error", err))
		return 1
	}

	scheduler.Stop()

	slog.Info("graceful shutdown ok")

	return 0
}

func loadPrivateKey(path string) (_ ed25519.PrivateKey, err error) {
	defer util.Wrap(&err, "loadPrivateKey(path=%s)", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to find block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("key is not ed25519 private key")
	}
	return privateKey, nil
}

func loadPublicKey(path string) (_ ed25519.PublicKey, err error) {
	defer util.Wrap(&err, "loadPublicKey(path=%s)", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to find block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("key is not ed25519 public key")
	}
	return privateKey, nil
}
