package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/ocr"
)

type client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewFormatterClient は OpenAI API を利用する整形クライアントを生成します。
func NewFormatterClient() (ocr.Formatter, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, ErrMissingAPIKey
	}

	model := os.Getenv("OPENAI_CHAT_MODEL")
	if model == "" {
		model = "gpt-4o-mini" // lowest-cost Chat Completions model as of 2024-08
	}

	return &client{
		apiKey: key,
		model:  model,
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
		Model:       c.model,
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

	const maxRetries = 3

	var completion chatResponse
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := c.newChatRequest(ctx, body)
		if err != nil {
			return receipt.FormattedReceipt{}, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return receipt.FormattedReceipt{}, err
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return receipt.FormattedReceipt{}, fmt.Errorf("failed to read OpenAI response: %w", err)
		}

		var apiErr apiError
		_ = json.Unmarshal(respBody, &apiErr)

		if resp.StatusCode == http.StatusTooManyRequests {
			if apiErr.Error.Code == "insufficient_quota" {
				return receipt.FormattedReceipt{}, fmt.Errorf("%w: %s (請求/クレジットを確認してください)", ocr.ErrInsufficientQuota, apiErr.Error.Message)
			}

			if attempt < maxRetries {
				time.Sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
				continue
			}
		}

		if apiErr.Error.Message != "" {
			return receipt.FormattedReceipt{}, fmt.Errorf("openai api error: status %d: %s", resp.StatusCode, apiErr.Error.Message)
		}

		if resp.StatusCode >= 400 {
			time.Sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
			return receipt.FormattedReceipt{}, fmt.Errorf("openai api error: status %d: %s", resp.StatusCode, string(respBody))
		}

		if err := json.Unmarshal(respBody, &completion); err != nil {
			return receipt.FormattedReceipt{}, fmt.Errorf("failed to decode OpenAI response: %w", err)
		}

		break
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

func (c *client) newChatRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if retryAfter != "" {
		if sec, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(sec) * time.Second
		}

		if t, err := http.ParseTime(retryAfter); err == nil {
			if d := time.Until(t); d > 0 {
				return d
			}
		}
	}

	return time.Duration(1<<attempt) * time.Second
}
