package telegram

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockTransport struct {
	targetURL *url.URL
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = m.targetURL.Scheme
	req.URL.Host = m.targetURL.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestNewValidation(t *testing.T) {
	t.Run("Empty Token", func(t *testing.T) {
		_, err := New("", "1234")
		if err == nil {
			t.Error("expected error for empty bot token, got nil")
		}
	})

	t.Run("Empty Chat ID", func(t *testing.T) {
		_, err := New("token", "")
		if err == nil {
			t.Error("expected error for empty chat ID, got nil")
		}
	})

	t.Run("Valid Setup", func(t *testing.T) {
		b, err := New("token", "1234")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.botToken != "token" || b.chatID != "1234" {
			t.Errorf("mismatched configuration: %+v", b)
		}
	})
}

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = r.ParseForm()
		if r.FormValue("chat_id") != "1234" {
			t.Errorf("expected chat_id 1234, got %s", r.FormValue("chat_id"))
		}
		if r.FormValue("text") != "hello world" {
			t.Errorf("expected text 'hello world', got %s", r.FormValue("text"))
		}
		if r.FormValue("parse_mode") != "HTML" {
			t.Errorf("expected parse_mode 'HTML', got %s", r.FormValue("parse_mode"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	oldTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &mockTransport{targetURL: u}
	defer func() { http.DefaultClient.Transport = oldTransport }()

	b, _ := New("token", "1234")
	err := b.SendMessage("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendChatAction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendChatAction") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = r.ParseForm()
		if r.FormValue("chat_id") != "1234" {
			t.Errorf("expected chat_id 1234, got %s", r.FormValue("chat_id"))
		}
		if r.FormValue("action") != "typing" {
			t.Errorf("expected action 'typing', got %s", r.FormValue("action"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	oldTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &mockTransport{targetURL: u}
	defer func() { http.DefaultClient.Transport = oldTransport }()

	b, _ := New("token", "1234")
	err := b.SendChatAction(Typing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendPhoto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendPhoto") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		if r.FormValue("chat_id") != "1234" {
			t.Errorf("expected chat_id 1234, got %s", r.FormValue("chat_id"))
		}
		if r.FormValue("caption") != "caption text" {
			t.Errorf("expected caption 'caption text', got %s", r.FormValue("caption"))
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			t.Fatalf("expected photo file, got error: %v", err)
		}
		defer file.Close()

		if header.Filename != "graph.png" {
			t.Errorf("expected filename graph.png, got %s", header.Filename)
		}

		content, _ := io.ReadAll(file)
		if !bytes.Equal(content, []byte("fake-photo")) {
			t.Errorf("unexpected file content: %s", string(content))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	oldTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &mockTransport{targetURL: u}
	defer func() { http.DefaultClient.Transport = oldTransport }()

	b, _ := New("token", "1234")
	err := b.SendPhoto([]byte("fake-photo"), "caption text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
