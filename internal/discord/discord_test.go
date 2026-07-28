package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vernak2539/netmon-go/internal/notifier"
)

func TestNewValidation(t *testing.T) {
	t.Run("Empty Webhook URL", func(t *testing.T) {
		_, err := New("")
		if err == nil {
			t.Error("expected error for empty webhook URL, got nil")
		}
	})

	t.Run("Whitespace Webhook URL", func(t *testing.T) {
		_, err := New("   ")
		if err == nil {
			t.Error("expected error for whitespace webhook URL, got nil")
		}
	})

	t.Run("Valid Webhook URL", func(t *testing.T) {
		client, err := New("https://discord.com/api/webhooks/123/abc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})
}

func TestSendMessage(t *testing.T) {
	t.Run("Successful Message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				t.Errorf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading body: %v", err)
			}

			var payload map[string]string
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("unmarshaling json body: %v", err)
			}

			if payload["content"] != "Hello Discord" {
				t.Errorf("expected content 'Hello Discord', got '%s'", payload["content"])
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := New(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := client.SendMessage(context.Background(), "Hello Discord"); err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}
	})

	t.Run("HTTP Error Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message": "Invalid Webhook"}`))
		}))
		defer server.Close()

		client, err := New(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.SendMessage(context.Background(), "Hello Discord")
		if err == nil {
			t.Fatal("expected error on 400 response, got nil")
		}
		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected status code 400 in error, got: %v", err)
		}
	})
}

func TestSendPhoto(t *testing.T) {
	t.Run("Successful Photo Upload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				t.Fatalf("parsing multipart form: %v", err)
			}

			payloadJSON := r.FormValue("payload_json")
			var payload map[string]string
			if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
				t.Fatalf("unmarshaling payload_json: %v", err)
			}

			if payload["content"] != "Graph Caption" {
				t.Errorf("expected payload content 'Graph Caption', got '%s'", payload["content"])
			}

			file, header, err := r.FormFile("file")
			if err != nil {
				t.Fatalf("retrieving form file 'file': %v", err)
			}
			defer file.Close()

			if header.Filename != "graph.png" {
				t.Errorf("expected filename 'graph.png', got '%s'", header.Filename)
			}

			content, _ := io.ReadAll(file)
			if !bytes.Equal(content, []byte("fake-png-bytes")) {
				t.Errorf("unexpected file content: %s", string(content))
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := New(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.SendPhoto(context.Background(), []byte("fake-png-bytes"), "Graph Caption")
		if err != nil {
			t.Fatalf("SendPhoto failed: %v", err)
		}
	})

	t.Run("HTTP Error Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal server error"))
		}))
		defer server.Close()

		client, err := New(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.SendPhoto(context.Background(), []byte("fake-png-bytes"), "Graph Caption")
		if err == nil {
			t.Fatal("expected error on 500 response, got nil")
		}
	})
}

func TestSendChatAction(t *testing.T) {
	client, err := New("https://discord.com/api/webhooks/123/abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := client.SendChatAction(context.Background(), notifier.ChatActionTyping); err != nil {
		t.Errorf("expected nil error for SendChatAction, got %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = client.SendMessage(ctx, "test message")
	if err == nil {
		t.Fatal("expected error on context timeout, got nil")
	}
}
