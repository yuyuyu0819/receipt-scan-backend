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
	"strings" // ★ 追加
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
		model = "gpt-4o-mini"
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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Temperature float32       `json:"temperature"`
	Messages    []chatMessage `json:"messages"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

func (c *client) Format(ctx context.Context, rawText string) (receipt.FormattedReceipt, error) {
	payload := chatRequest{
		Model:       c.model,
		Temperature: 0.2,
		Messages: []chatMessage{
			{
				Role: "system",
				Content: "あなたはレシートOCR結果を JSON に整形するアシスタントです。" +
					"説明文、Markdown、コードブロック（```）を一切含めず、純粋なJSONのみを返してください。",
			},
			{
				Role: "user",
				Content: fmt.Sprintf(
					"以下のOCR結果を JSON で返してください。"+
						"スキーマ: {\"store\": string, \"date\": string, \"total\": number, \"items\": [{\"name\": string, \"price\": number}]}."+
						"date は必ず YYYY-MM-DD 形式（例: 2025-12-13）。数値は半角。"+
						"\nOCR結果:\n%s",
					rawText,
				),
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

		if apiErr.Error.Code == "context_length_exceeded" {
			return receipt.FormattedReceipt{}, fmt.Errorf(
				"%w: 応答上限を超える長さのテキストが送信されました。OCR結果を短くして再試行してください",
				ocr.ErrContextLengthExceeded,
			)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if apiErr.Error.Code == "insufficient_quota" {
				return receipt.FormattedReceipt{}, fmt.Errorf(
					"%w: %s (請求/クレジットを確認してください)",
					ocr.ErrInsufficientQuota,
					apiErr.Error.Message,
				)
			}

			if attempt < maxRetries {
				time.Sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
				continue
			}
		}

		if apiErr.Error.Message != "" {
			return receipt.FormattedReceipt{}, fmt.Errorf(
				"openai api error: status %d: %s",
				resp.StatusCode,
				apiErr.Error.Message,
			)
		}

		if resp.StatusCode >= 400 {
			time.Sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
			return receipt.FormattedReceipt{}, fmt.Errorf(
				"openai api error: status %d: %s",
				resp.StatusCode,
				string(respBody),
			)
		}

		if err := json.Unmarshal(respBody, &completion); err != nil {
			return receipt.FormattedReceipt{}, fmt.Errorf("failed to decode OpenAI response: %w", err)
		}

		break
	}

	if len(completion.Choices) == 0 {
		return receipt.FormattedReceipt{}, errors.New("no choices returned from OpenAI")
	}

	// ===== ★ ここが今回の本質的修正点 =====
	content := strings.TrimSpace(completion.Choices[0].Message.Content)

	// ```json ... ``` を除去
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}
	// ======================================

	var formatted receipt.FormattedReceipt
	if err := json.Unmarshal([]byte(content), &formatted); err != nil {
		return receipt.FormattedReceipt{}, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return formatted, nil
}

func (c *client) newChatRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(body),
	)
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
