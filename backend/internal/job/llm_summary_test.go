package job

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/llm"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- helpers ---

func requireGemini(t *testing.T) llm.Service {
	t.Helper()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY is not set")
	}
	svc, err := llm.NewGemini(context.Background(), apiKey, "gemini-2.5-flash-lite")
	if err != nil {
		t.Fatalf("failed to create gemini service: %v", err)
	}
	return svc
}

func seedVideoForJob(t *testing.T, pool *pgxpool.Pool, title string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	q := sqlc.New(pool)
	now := time.Now().UTC()

	chPubID := uuid.Must(uuid.NewV7())
	extID := "UC" + chPubID.String()[:22]

	if err := q.ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
		ExternalCustomID: "@ch-" + chPubID.String()[:8],
		ExternalID:       extID,
	}); err != nil {
		t.Fatalf("seedVideoForJob: ClearStaleChannelCustomID: %v", err)
	}

	chRow, err := q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
		ExternalID:                extID,
		ExternalDisplayName:       "Test Channel",
		ExternalCustomID:          "@ch-" + chPubID.String()[:8],
		ExternalIconUrl:           "https://example.com/icon.jpg",
		ExternalDescription:       "A test channel",
		ExternalSubscribersCount:  1000,
		ExternalCreatedAt:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalUploadsPlaylistID: "UU" + chPubID.String()[:22],
		PublicID:                  chPubID,
		RssFetchedAt:              now,
		FetchedAt:                 now,
	})
	if err != nil {
		t.Fatalf("seedVideoForJob: UpsertChannel: %v", err)
	}

	videoPubID := uuid.Must(uuid.NewV7())
	if _, err := q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
		ChannelID:             chRow.PublicID,
		ExternalID:            videoPubID.String()[:16],
		ExternalTitle:         title,
		ExternalDescription:   "desc",
		FetchedAt:             now,
		ExternalCreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalThumbnailUrl:  "https://example.com/thumb.jpg",
		ExternalLengthSeconds: 300,
		ID:                    videoPubID,
	}); err != nil {
		t.Fatalf("seedVideoForJob: UpsertVideo: %v", err)
	}

	return videoPubID
}

func seedWatchRecord(t *testing.T, pool *pgxpool.Pool, userPubID, videoPubID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	q := sqlc.New(pool)
	now := time.Now().UTC()

	if err := q.UpsertWatchHeartbeat(ctx, sqlc.UpsertWatchHeartbeatParams{
		UserPublicID:         userPubID,
		VideoPublicID:        videoPubID,
		WatchPositionSeconds: 60,
		PublicID:             uuid.Must(uuid.NewV7()),
		WatchStartAt:         now,
	}); err != nil {
		t.Fatalf("seedWatchRecord: %v", err)
	}
}

// --- unit tests (no DB) ---

func TestCreateTodaysBits(t *testing.T) {
	t.Parallel()
	bits, err := createTodaysBits()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// int32 * 3 = 12 bytes
	if len(bits) != 12 {
		t.Fatalf("expected 12 bytes, got %d", len(bits))
	}
}

func TestTimeToUUID(t *testing.T) {
	t.Parallel()
	ts := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	u := timeToUUID(ts)

	// version nibble should be 7
	if (u[6] >> 4) != 0x07 {
		t.Fatalf("expected version 7, got %d", u[6]>>4)
	}
	// variant should be 10xx
	if (u[8] >> 6) != 0x02 {
		t.Fatalf("expected variant 10, got %d", u[8]>>6)
	}

	// millisecond timestamp should round-trip
	var ms uint64
	ms |= uint64(u[0]) << 40
	ms |= uint64(u[1]) << 32
	ms |= uint64(u[2]) << 24
	ms |= uint64(u[3]) << 16
	ms |= uint64(u[4]) << 8
	ms |= uint64(u[5])
	if int64(ms) != ts.UnixMilli() {
		t.Fatalf("timestamp mismatch: got %d, want %d", ms, ts.UnixMilli())
	}
}

func TestBuildSummaryPrompt_Japanese(t *testing.T) {
	t.Parallel()
	prompts := buildSummaryPrompt("動画A,動画B", "ja")
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].Role != "user" {
		t.Fatalf("expected role 'user', got %q", prompts[0].Role)
	}
	if !strings.Contains(prompts[0].Message, "動画A,動画B") {
		t.Fatal("expected titles in prompt message")
	}
	if !strings.Contains(prompts[0].Message, "YouTube") {
		t.Fatal("expected Japanese template content")
	}
}

func TestBuildSummaryPrompt_English(t *testing.T) {
	t.Parallel()
	prompts := buildSummaryPrompt("VideoA,VideoB", "en")
	if !strings.Contains(prompts[0].Message, "Video titles watched") {
		t.Fatal("expected English template content")
	}
}

func TestBuildSummaryPrompt_UnknownLanguageFallsBackToEnglish(t *testing.T) {
	t.Parallel()
	prompts := buildSummaryPrompt("titles", "fr")
	if !strings.Contains(prompts[0].Message, "Video titles watched") {
		t.Fatal("expected English fallback template")
	}
}

// --- integration tests (require DB + Gemini API) ---

func TestLLMSummaryJob_Run(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	gemini := requireGemini(t)

	userPubID := testutil.SeedUser(t, pool)

	titles := []string{
		"【初心者向け】猫の飼い方完全ガイド",
		"子猫が初めてお風呂に入った結果…",
		"プロが教える猫じゃらしテクニック",
		"猫カフェ巡り東京編 TOP10",
		"野良猫の保護活動ドキュメンタリー",
	}
	for _, title := range titles {
		videoPubID := seedVideoForJob(t, pool, title)
		seedWatchRecord(t, pool, userPubID, videoPubID)
	}

	job := &llmSummaryJob{db: pool, llmService: gemini}
	if err := job.run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var title, description, model string
	err := pool.QueryRow(ctx,
		"SELECT ai_summary_title, ai_summary_description, ai_model FROM s_monthly_video_watch LIMIT 1",
	).Scan(&title, &description, &model)
	if err != nil {
		t.Fatalf("expected summary row: %v", err)
	}
	if title == "" {
		t.Fatal("expected non-empty title")
	}
	if description == "" {
		t.Fatal("expected non-empty description")
	}
	if model != gemini.ModelID() {
		t.Fatalf("expected model %q, got %q", gemini.ModelID(), model)
	}

	t.Logf("\n=== Generated Summary ===\nModel: %s\nTitle: %s\nDescription: %s\n", model, title, description)
}

func TestLLMSummaryJob_Run_NoWatchRecords(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	gemini := requireGemini(t)

	testutil.SeedUser(t, pool)

	job := &llmSummaryJob{db: pool, llmService: gemini}
	if err := job.run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM s_monthly_video_watch").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 summary rows, got %d", count)
	}
}

func TestLLMSummaryJob_Run_MultipleUsers(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	gemini := requireGemini(t)

	user1 := testutil.SeedUser(t, pool)
	user2 := testutil.SeedUser(t, pool)

	cookingTitles := []string{
		"【簡単レシピ】10分で作れるパスタ",
		"プロの包丁さばきを学ぶ",
		"世界の屋台グルメ TOP5",
	}
	gamingTitles := []string{
		"【マイクラ】巨大建築に挑戦！",
		"スプラトゥーン3 ガチマッチ攻略",
		"ゼルダの伝説 100%クリアRTA",
	}
	for _, title := range cookingTitles {
		v := seedVideoForJob(t, pool, title)
		seedWatchRecord(t, pool, user1, v)
	}
	for _, title := range gamingTitles {
		v := seedVideoForJob(t, pool, title)
		seedWatchRecord(t, pool, user2, v)
	}

	job := &llmSummaryJob{db: pool, llmService: gemini}
	if err := job.run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, err := pool.Query(ctx,
		"SELECT m_user_id, ai_summary_title, ai_summary_description FROM s_monthly_video_watch ORDER BY m_user_id",
	)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var userID int64
		var title, desc string
		if err := rows.Scan(&userID, &title, &desc); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		count++
		t.Logf("\n=== User %d Summary ===\nTitle: %s\nDescription: %s\n", userID, title, desc)
	}
	if count != 2 {
		t.Fatalf("expected 2 summary rows, got %d", count)
	}
}

func TestLLMSummaryJob_Run_UpsertOnSecondRun(t *testing.T) {
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	gemini := requireGemini(t)

	userPubID := testutil.SeedUser(t, pool)
	videoPubID := seedVideoForJob(t, pool, "テスト動画")
	seedWatchRecord(t, pool, userPubID, videoPubID)

	job := &llmSummaryJob{db: pool, llmService: gemini}

	if err := job.run(ctx); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := job.run(ctx); err != nil {
		t.Fatalf("second run: %v", err)
	}

	// upsertなのでレコードは1件のまま
	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM s_monthly_video_watch").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 summary row after upsert, got %d", count)
	}
}
