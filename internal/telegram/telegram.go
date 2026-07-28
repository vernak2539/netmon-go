package telegram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vernak2539/netmon-go/internal/notifier"
)

type ChatAction = notifier.ChatAction

const (
	Typing      ChatAction = notifier.ChatActionTyping
	UploadPhoto ChatAction = notifier.ChatActionUploadPhoto
)

type Bot struct {
	botToken   string
	chatID     string
	httpClient *http.Client
}

type Client = Bot

var _ notifier.Notifier = (*Client)(nil)

// New creates a new Telegram Bot client.
func New(botToken, chatID string, timeouts ...time.Duration) (*Bot, error) {
	if strings.TrimSpace(botToken) == "" {
		return nil, fmt.Errorf("bot token cannot be empty")
	}
	if strings.TrimSpace(chatID) == "" {
		return nil, fmt.Errorf("chat ID cannot be empty")
	}
	client := http.DefaultClient
	if len(timeouts) > 0 && timeouts[0] > 0 {
		client = &http.Client{Timeout: timeouts[0]}
	}
	return &Bot{
		botToken:   botToken,
		chatID:     chatID,
		httpClient: client,
	}, nil
}

func (b *Bot) getHTTPClient() *http.Client {
	if b.httpClient != nil {
		return b.httpClient
	}
	return http.DefaultClient
}

// SendMessage sends a text message with HTML parse mode to the configured chat.
func (b *Bot) SendMessage(ctx context.Context, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.botToken)
	data := url.Values{}
	data.Set("chat_id", b.chatID)
	data.Set("text", message)
	data.Set("parse_mode", "HTML")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.getHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendPhoto uploads a photo with a caption using multipart/form-data.
func (b *Bot) SendPhoto(ctx context.Context, photo []byte, caption string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", b.botToken)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add photo file
	part, err := writer.CreateFormFile("photo", "graph.png")
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(photo)); err != nil {
		return fmt.Errorf("copying photo data: %w", err)
	}

	// Add other fields
	if err := writer.WriteField("chat_id", b.chatID); err != nil {
		return fmt.Errorf("writing chat_id: %w", err)
	}
	if err := writer.WriteField("caption", caption); err != nil {
		return fmt.Errorf("writing caption: %w", err)
	}
	if err := writer.WriteField("parse_mode", "HTML"); err != nil {
		return fmt.Errorf("writing parse_mode: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.getHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendChatAction sends a status indicator (typing, uploading photo) to the configured chat.
func (b *Bot) SendChatAction(ctx context.Context, action notifier.ChatAction) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendChatAction", b.botToken)
	data := url.Values{}
	data.Set("chat_id", b.chatID)
	data.Set("action", string(action))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.getHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
