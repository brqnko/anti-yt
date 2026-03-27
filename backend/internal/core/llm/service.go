package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type Prompt struct {
	Role    string
	Message string
}

type Option func(*genai.GenerateContentConfig)

// WithJSONSchema は構造化出力を有効にし、レスポンスを指定したJSONスキーマに従わせる
func WithJSONSchema(schema *genai.Schema) Option {
	return func(c *genai.GenerateContentConfig) {
		c.ResponseMIMEType = "application/json"
		c.ResponseSchema = schema
	}
}

type Service interface {
	Completion(ctx context.Context, prompts []Prompt, opts ...Option) (_ string, err error)
	ModelID() string
}

var _ Service = (*geminiImpl)(nil)

type geminiImpl struct {
	client  *genai.Client
	modelID string
}

func (g *geminiImpl) Completion(ctx context.Context, prompts []Prompt, opts ...Option) (_ string, err error) {
	contents := make([]*genai.Content, len(prompts))
	for i, p := range prompts {
		contents[i] = &genai.Content{
			Role:  p.Role,
			Parts: []*genai.Part{genai.NewPartFromText(p.Message)},
		}
	}

	var config *genai.GenerateContentConfig
	if len(opts) > 0 {
		config = &genai.GenerateContentConfig{}
		for _, opt := range opts {
			opt(config)
		}
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.modelID, contents, config)
	if err != nil {
		return "", fmt.Errorf("gemini completion: %w", err)
	}

	text := resp.Text()
	if text == "" {
		return "", fmt.Errorf("gemini completion: empty response")
	}

	return text, nil
}

func (g *geminiImpl) ModelID() string {
	return g.modelID
}

func NewGemini(ctx context.Context, apiKey, modelID string) (Service, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("creating gemini client: %w", err)
	}

	return &geminiImpl{
		client:  client,
		modelID: modelID,
	}, nil
}
