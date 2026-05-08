package discord_d

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/util"
)

type Client interface {
	SendWebhookMessage(ctx context.Context, message string) (err error)
}

var _ Client = (*clientImpl)(nil)

type clientImpl struct {
	webhookURL string
	httpClient *http.Client
}

type webhookPayload struct {
	Content string `json:"content"`
}

func (w *clientImpl) SendWebhookMessage(ctx context.Context, message string) (err error) {
	defer util.Wrap(&err, "discord_d.(*clientImpl).SendWebhookMessage")

	payload, err := json.Marshal(webhookPayload{Content: message})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

func NewClient(webhookURL string) Client {
	return &clientImpl{
		webhookURL: webhookURL,
		httpClient: new(http.Client{}),
	}
}
