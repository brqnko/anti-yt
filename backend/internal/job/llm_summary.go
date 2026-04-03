package job

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/llm"
	"github.com/brqnko/anti-yt/backend/internal/core/scheduler"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/genai"
)

type llmSummaryJob struct {
	db *pgxpool.Pool

	llmService llm.Service
	mx         *sync.Mutex
}

type summaryResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

var summarySchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"title":       {Type: genai.TypeString, Description: "short summary title (max 50 characters)"},
		"description": {Type: genai.TypeString, Description: "detailed description of viewing patterns and interests (max 500 characters)"},
	},
	Required: []string{"title", "description"},
}

var summaryPromptTemplates = map[string]string{
	"ja": `あなたはYouTubeの視聴履歴を分析するAIアシスタントです。
ユーザーが視聴した動画タイトルの一覧から、視聴傾向を簡潔にまとめてください。

titleは50文字以内の短い要約タイトル、descriptionは500文字以内の視聴傾向や興味関心の詳細な説明です。

視聴した動画タイトル:
%s`,

	"en": `You are an AI assistant that analyzes YouTube viewing history.
Given a list of video titles a user watched, provide a brief summary of their viewing habits.

title should be a short summary title (max 50 characters), description should be a detailed description of the viewing patterns and interests (max 500 characters).

Video titles watched:
%s`,
}

// UTC時刻のYMDで[]byteを構築します
func createTodaysBits() (_ []byte, err error) {
	defer util.Wrap(&err, "job.createTodaysBits")

	y, m, d := time.Now().UTC().Date()
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int32(y)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, int32(m)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, int32(d)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// uuidv7をtime.Timeから実装する。右のbitは0で埋め尽くす
func timeToUUID(t time.Time) uuid.UUID {
	var u uuid.UUID

	ms := uint64(t.UnixMilli())

	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	u[6] = 0x70

	u[8] = 0x80

	return u
}

func (j *llmSummaryJob) run(ctx context.Context) (err error) {
	defer util.Wrap(&err, "job.(*llmSummaryJob).run")

	tx, err := j.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback in llmSummaryJob.run", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// ad lock
	key, err := createTodaysBits()
	if err != nil {
		return err
	}
	if err := database_d.TryAdLock(ctx, q, key); err != nil {
		return err
	}

	startedAt := time.Now().UTC()
	y, m, d := startedAt.Date()
	rows, err := q.GetVideoWatchTitlesByUser(
		ctx,
		timeToUUID(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -7)),
	)
	if err != nil {
		return err
	}

	targetMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)

	for _, row := range rows {
		titles := string(row.TitleConcat)
		if titles == "" {
			continue
		}

		tmpl, ok := summaryPromptTemplates[row.LanguageCode]
		if !ok {
			tmpl = summaryPromptTemplates["en"]
		}
		prompts := []llm.Prompt{
			{
				Role:    "user",
				Message: fmt.Sprintf(tmpl, titles),
			},
		}
		resp, err := j.llmService.Completion(ctx, prompts, llm.WithJSONSchema(summarySchema))
		if err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "llm completion failed in summary job", slog.Int64("user_id", row.UserID), slog.Any("error", err))
			continue
		}

		var summary summaryResponse
		if err := json.Unmarshal([]byte(resp), &summary); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "unmarshal summary response failed", slog.Int64("user_id", row.UserID), slog.Any("error", err))
			continue
		}

		// VARCHAR(128) / VARCHAR(4096) はcharacter数なのでruneで切る
		if runes := []rune(summary.Title); len(runes) > 128 {
			summary.Title = string(runes[:128])
		}
		if runes := []rune(summary.Description); len(runes) > 4096 {
			summary.Description = string(runes[:4096])
		}

		if err := q.UpsertMonthlyVideoWatchSummary(ctx, sqlc.UpsertMonthlyVideoWatchSummaryParams{
			UserID:               row.UserID,
			AiSummaryTitle:       summary.Title,
			AiSummaryDescription: summary.Description,
			AiModel:              j.llmService.ModelID(),
			GeneratedAt:          startedAt,
			TargetMonth:          targetMonth,
		}); err != nil {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "upsert summary failed", slog.Int64("user_id", row.UserID), slog.Any("error", err))
			continue
		}
	}

	if err := tx.Commit(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return err
	}

	return nil
}

func (j *llmSummaryJob) Run() {
	// ad lockするけど一応
	j.mx.Lock()
	defer j.mx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := j.run(ctx); err != nil {
		util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to run llm summary job", slog.Any("error", err))
	}
}

func NewLLMSummaryJob(db *pgxpool.Pool, llmService llm.Service) scheduler.Job {
	return &llmSummaryJob{
		llmService: llmService,
		db:         db,
		mx:         &sync.Mutex{},
	}
}
