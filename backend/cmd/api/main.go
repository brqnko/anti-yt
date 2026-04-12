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
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/admin"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/middleware_d"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/llm"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/core/report"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/job"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
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
		// uuid.UUID型の属性値をbase64(RawURL)に変換して出力する
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if id, ok := a.Value.Any().(uuid.UUID); ok {
					return slog.String(a.Key, base64.RawURLEncoding.EncodeToString(id[:]))
				}
				return a
			},
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
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
	jtiBlacklistRepo := database_d.NewJtiBlacklistRepository(redisClient)
	ratelimitRepo := database_d.NewRatelimitRepository(redisClient, 24*time.Hour)
	feedRepo := database_d.NewFeedRepository(redisClient, 1000)

	job.NewPurgeLeftUserJob(db, feedRepo, discord_d.NewDiscordClient(cfg.discordWebhookURL)).Run()
	job.NewRefillFeedJob(db, feedRepo, discord_d.NewDiscordClient(cfg.discordWebhookURL)).Run()

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	oidcService, err := oidc.NewGoogleOIDCService(initCtx, cfg.oidcGoogleClientID, cfg.oidcGoogleClientSecret, fmt.Sprintf("%s/v1/auth/google/callback", cfg.serverURL))
	if err != nil {
		slog.Error("failed to create oidc service", slog.Any("error", err))
		return 1
	}

	jwtService := jwt_d.NewService(jwtPrivate, jwtPublic, 30*time.Minute, cfg.serverURL)

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ytService, err := youtube_d.NewService(initCtx, cfg.youtubeDataAPIKey, cfg.oidcGoogleClientID, cfg.oidcGoogleClientSecret, fmt.Sprintf("%s/v1/auth/oauth/youtube/callback", cfg.serverURL))
	if err != nil {
		slog.Error("failed to create youtube service", slog.Any("error", err))
		return 1
	}
	slog.Info("youtube api ok")

	initCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	llmService, err := llm.NewGemini(initCtx, cfg.geminiAPIKey, "gemini-2.5-flash-lite")
	if err != nil {
		slog.Error("failed to create llm service", slog.Any("error", err))
		return 1
	}
	slog.Info("gemini api ok")

	var reportOpts []report.Option
	if cfg.discordWebhookURL != "" {
		reportOpts = append(reportOpts, report.WithDiscord(discord_d.NewDiscordClient(cfg.discordWebhookURL)))
	}
	reportService := report.NewService(reportOpts...)

	scheduler := scheduler.NewService()
	if err := scheduler.AddFunc("0 0 * * *", job.NewLLMSummaryJob(db, llmService, discord_d.NewDiscordClient(cfg.discordWebhookURL))); err != nil {
		slog.Error("failed to setup llm summary job", slog.Any("error", err))
		return 1
	}
	if err := scheduler.AddFunc("0 0 * * *", job.NewPurgeLeftUserJob(db, feedRepo, discord_d.NewDiscordClient(cfg.discordWebhookURL))); err != nil {
		slog.Error("failed to setup purge left user job", slog.Any("error", err))
		return 1
	}
	if err := scheduler.AddFunc("50 6 * * *", job.NewExhaustQuotaJob(db, ytService, feedRepo, discord_d.NewDiscordClient(cfg.discordWebhookURL))); err != nil { // 夏
		slog.Error("failed to setup exhaust quota job (PDT)", slog.Any("error", err))
		return 1
	}
	if err := scheduler.AddFunc("50 7 * * *", job.NewExhaustQuotaJob(db, ytService, feedRepo, discord_d.NewDiscordClient(cfg.discordWebhookURL))); err != nil { // 冬
		slog.Error("failed to setup exhaust quota job (PST)", slog.Any("error", err))
		return 1
	}
	if err := scheduler.AddFunc("0 * * * *", job.NewRefillFeedJob(db, feedRepo, discord_d.NewDiscordClient(cfg.discordWebhookURL))); err != nil {
		slog.Error("failed to setup refill feed job", slog.Any("error", err))
		return 1
	}

	if cfg.discordWebhookURL != "" {
		if err := scheduler.AddFunc("0 0 * * *", job.NewAuthorizationReportJob(db, discord_d.NewDiscordClient(cfg.discordWebhookURL))); err != nil {
			slog.Error("failed to setup authorization report job", slog.Any("error", err))
			return 1
		}
	}

	r := chi.NewRouter()
	// NOTE: RateLimit, gzip, loggerはLB(pingora, nginx等)で行う。各リクエストへのレートリミットはmiddlewareが行う
	// r.Useは上に書いた順から実行される
	r.Use(middleware.Recoverer)
	r.Use(middleware_d.SecureHeaders)
	if cfg.env != "production" {
		r.Use(middleware_d.RandomLag)
		r.Use(v1.SwaggerMiddleware())
	}
	admin.HandleAdminEndpoints(r, db, ytService, feedRepo, cfg.adminAPIKey)
	v1.HandlerFromMux(v1.NewStrictHandler(
		v1.NewAPIHandler(db, oidcService, cfg.serverURL, cfg.frontendURL, jwtService, 30*24*time.Hour, ytService, 1*time.Hour, scheduler, jtiBlacklistRepo, feedRepo),
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
			middleware_d.WrapErrorMiddleware(reportService),
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
			middleware_d.UserRatelimitMiddleware(ratelimitRepo, 2000, map[string]int{
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
			middleware_d.AccessTokenMiddleware(jwtService, jtiBlacklistRepo, map[string]struct{}{
				"GetAuthGoogle":               {},
				"GetAuthGoogleCallback":       {},
				"PostAuthRefresh":             {},
				"GetAuthOauthYoutubeCallback": {},
				"PostUsersMe":                 {},
				"PostAuthReactivate":          {},
				"GetChannelsChannelId":          {},
				"GetChannelsChannelIdVideos":    {},
				"GetChannelsChannelIdPlaylists": {},
				"GetVideosVideoId":              {},
				"GetPlaylistsPlaylistId":        {},
				"GetFeed":                       {},
			}),
			middleware_d.RequestIDMiddleware,
		}), r)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("shutting down the server")
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

func loadPrivateKey(path string) (ed25519.PrivateKey, error) {
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

func loadPublicKey(path string) (ed25519.PublicKey, error) {
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
