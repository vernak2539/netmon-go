package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/vernak2539/netmon-go/internal/notifier"
)

type Client struct {
	webhookURL string
	httpClient *http.Client
}

var _ notifier.Notifier = (*Client)(nil)

// New creates a new Discord webhook client.
func New(webhookURL string, timeouts ...time.Duration) (*Client, error) {
	if strings.TrimSpace(webhookURL) == "" {
		return nil, fmt.Errorf("webhook URL cannot be empty")
	}
	client := &http.Client{}
	if len(timeouts) > 0 && timeouts[0] > 0 {
		client.Timeout = timeouts[0]
	}
	return &Client{
		webhookURL: webhookURL,
		httpClient: client,
	}, nil
}

// SendMessage sends a text message payload via Discord Webhook.
func (c *Client) SendMessage(ctx context.Context, text string) error {
	payload := map[string]string{"content": text}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling json payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendPhoto uploads a photo with caption payload via Discord Webhook using multipart/form-data.
func (c *Client) SendPhoto(ctx context.Context, photo []byte, caption string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	payloadJSON, err := json.Marshal(map[string]string{"content": caption})
	if err != nil {
		return fmt.Errorf("marshaling payload_json: %w", err)
	}

	if err := writer.WriteField("payload_json", string(payloadJSON)); err != nil {
		return fmt.Errorf("writing payload_json: %w", err)
	}

	part, err := writer.CreateFormFile("file", "graph.png")
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(photo)); err != nil {
		return fmt.Errorf("copying photo data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SendChatAction is a no-op for Discord webhooks.
func (c *Client) SendChatAction(ctx context.Context, action notifier.ChatAction) error {
	return nil
}
