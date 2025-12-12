package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/ocr"
)

type client struct {
	apiKey     string
	httpClient *http.Client
}

// NewFormatterClient は OpenAI API を利用する整形クライアントを生成します。
func NewFormatterClient() (ocr.Formatter, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, ErrMissingAPIKey
	}

	return &client{
		apiKey: key,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ErrMissingAPIKey は API キーが未設定のときに返されます。
var ErrMissingAPIKey = errors.New("OPENAI_API_KEY is not set")

// chatMessage は Chat Completions API のメッセージを表します。
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest は Chat Completions API のリクエストボディです。
type chatRequest struct {
	Model       string        `json:"model"`
	Temperature float32       `json:"temperature"`
	Messages    []chatMessage `json:"messages"`
}

// chatChoice は Chat Completions API のレスポンスの一部です。
type chatChoice struct {
	Message chatMessage `json:"message"`
}

// chatResponse は Chat Completions API のレスポンスボディです。
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

func (c *client) Format(ctx context.Context, rawText string) (receipt.FormattedReceipt, error) {
	payload := chatRequest{
		Model:       "gpt-4o-mini",
		Temperature: 0.2,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "あなたはレシートOCR結果を JSON に整形するアシスタントです。必ず有効な JSON のみを返してください。",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("以下のOCR結果を JSON で返してください。スキーマ: {\\\"store\\\": string, \\\"date\\\": string, \\\"total\\\": number, \\\"items\\\": [{\\\"name\\\": string, \\\"price\\\": number}]}. 数値は半角で。\\nOCR結果:\n%s", rawText),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return receipt.FormattedReceipt{}, fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return receipt.FormattedReceipt{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return receipt.FormattedReceipt{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return receipt.FormattedReceipt{}, fmt.Errorf("openai api error: status %d", resp.StatusCode)
	}

	var completion chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return receipt.FormattedReceipt{}, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return receipt.FormattedReceipt{}, errors.New("no choices returned from OpenAI")
	}

	var formatted receipt.FormattedReceipt
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &formatted); err != nil {
		return receipt.FormattedReceipt{}, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return formatted, nil
}
