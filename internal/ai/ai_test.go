package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewValidation(t *testing.T) {
	t.Run("Empty API Key", func(t *testing.T) {
		client, err := New("", "gpt-4", "http://localhost")
		if err != nil {
			t.Fatalf("expected nil error for empty API key, got: %v", err)
		}
		if client != nil {
			t.Errorf("expected nil client for empty API key, got: %v", client)
		}
	})

	t.Run("Empty Model", func(t *testing.T) {
		_, err := New("api-key", "", "http://localhost")
		if err == nil {
			t.Error("expected error for empty model, got nil")
		}
	})

	t.Run("Empty Base URL", func(t *testing.T) {
		_, err := New("api-key", "gpt-4", "")
		if err == nil {
			t.Error("expected error for empty base URL, got nil")
		}
	})

	t.Run("Valid Setup", func(t *testing.T) {
		client, err := New("api-key", "gpt-4", "http://localhost")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.model != "gpt-4" {
			t.Errorf("expected model gpt-4, got %s", client.model)
		}
	})
}

func TestNewEmptyAPIKey(t *testing.T) {
	client, err := New("", "model", "http://localhost")
	if err != nil {
		t.Fatalf("expected nil error for empty API key, got: %v", err)
	}
	if client != nil {
		t.Errorf("expected nil client for empty API key, got: %v", client)
	}
}

func TestSendMessage(t *testing.T) {
	t.Run("Empty Message Error", func(t *testing.T) {
		client, _ := New("key", "model", "http://localhost")
		_, err := client.SendMessage(context.Background(), "", "prompt")
		if err == nil {
			t.Error("expected error for empty message, got nil")
		}
	})

	t.Run("Successful Response", func(t *testing.T) {
		// Mock local OpenAI API server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleAssistant,
							Content: "Mocked AI Response",
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := New("key", "model", server.URL)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		response, err := client.SendMessage(context.Background(), "hello", "system")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if response != "Mocked AI Response" {
			t.Errorf("expected 'Mocked AI Response', got '%s'", response)
		}
	})

	t.Run("Empty AI Response Choice Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := New("key", "model", server.URL)
		_, err := client.SendMessage(context.Background(), "hello", "system")
		if err == nil {
			t.Error("expected error for empty choices response, got nil")
		}
	})
}
