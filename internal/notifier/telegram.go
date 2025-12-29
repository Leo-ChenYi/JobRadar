package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"jobradar/internal/config"
	"jobradar/internal/model"
)

const telegramAPIURL = "https://api.telegram.org/bot%s/sendMessage"

// TelegramNotifier sends notifications via Telegram
type TelegramNotifier struct {
	config config.TelegramConfig
	client *http.Client
}

// NewTelegram creates a new Telegram notifier
func NewTelegram(cfg config.TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the notifier name
func (t *TelegramNotifier) Name() string {
	return "telegram"
}

// Send sends a notification for a matched job
func (t *TelegramNotifier) Send(matched *model.MatchedJob) error {
	message := FormatTelegramMessage(matched)
	return t.sendMessage(message)
}

// SendTest sends a test notification
func (t *TelegramNotifier) SendTest() error {
	message := FormatTestMessage()
	return t.sendMessage(message)
}

// sendMessage sends a message to Telegram
func (t *TelegramNotifier) sendMessage(message string) error {
	payload := map[string]interface{}{
		"chat_id":                  t.config.ChatID,
		"text":                     message,
		"parse_mode":               "MarkdownV2",
		"disable_web_page_preview": false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf(telegramAPIURL, t.config.BotToken)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			OK          bool   `json:"ok"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Description != "" {
			return fmt.Errorf("telegram API error: %s", errResp.Description)
		}
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("telegram API returned ok=false")
	}

	return nil
}
