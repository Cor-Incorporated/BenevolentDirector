package llmclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPLLMClientCompleteParsesNDJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(`{"type":"content","content":"hello "}` + "\n"))
		_, _ = w.Write([]byte(`{"type":"content","content":"world"}` + "\n"))
		_, _ = w.Write([]byte(`{"type":"done","done":true,"event_type":"conversation.turn.completed"}` + "\n"))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	resp, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{Stream: true})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "hello world" {
		t.Fatalf("Content = %q, want %q", resp.Content, "hello world")
	}
	if len(resp.Chunks) != 3 {
		t.Fatalf("len(Chunks) = %d, want 3", len(resp.Chunks))
	}
}

func TestHTTPLLMClientCompleteParsesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"stub","choices":[{"message":{"content":"buffered"}}]}`))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	resp, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "buffered" {
		t.Fatalf("Content = %q, want %q", resp.Content, "buffered")
	}
}

func TestNewHTTPLLMClient_DefaultBaseURL(t *testing.T) {
	client := NewHTTPLLMClient("", nil)
	if client.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, defaultBaseURL)
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewHTTPLLMClient_WhitespaceOnlyBaseURL(t *testing.T) {
	client := NewHTTPLLMClient("   ", nil)
	if client.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, defaultBaseURL)
	}
}

func TestNewHTTPLLMClient_TrailingSlashTrimmed(t *testing.T) {
	client := NewHTTPLLMClient("http://example.com/", nil)
	if client.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", client.baseURL)
	}
}

func TestNewHTTPLLMClient_CustomHTTPClient(t *testing.T) {
	custom := &http.Client{}
	client := NewHTTPLLMClient("http://example.com", custom)
	if client.httpClient != custom {
		t.Error("expected custom httpClient to be used")
	}
}

func TestNewFromEnv(t *testing.T) {
	t.Setenv("LLM_GATEWAY_URL", "http://test-gateway:9090")
	client := NewFromEnv()
	if client.baseURL != "http://test-gateway:9090" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "http://test-gateway:9090")
	}
}

func TestNewFromEnv_EmptyEnv(t *testing.T) {
	t.Setenv("LLM_GATEWAY_URL", "")
	client := NewFromEnv()
	if client.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want default %q", client.baseURL, defaultBaseURL)
	}
}

func TestComplete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestComplete_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestComplete_ConnectionRefused(t *testing.T) {
	client := NewHTTPLLMClient("http://127.0.0.1:1", nil)
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestComplete_InvalidJSON_Buffered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestComplete_InvalidNDJSON_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not valid json\n"))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{Stream: true})
	if err == nil {
		t.Fatal("expected error for invalid NDJSON chunk")
	}
}

func TestComplete_StreamWithOnChunkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"type":"content","content":"hello"}` + "\n"))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{
		Stream: true,
		OnChunk: func(c Chunk) error {
			return fmt.Errorf("chunk handler error")
		},
	})
	if err == nil {
		t.Fatal("expected error from OnChunk callback")
	}
}

func TestComplete_StreamWithOnChunkSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"type":"content","content":"hi","data_classification":"internal"}` + "\n"))
	}))
	defer server.Close()

	var chunks []Chunk
	client := NewHTTPLLMClient(server.URL, server.Client())
	resp, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{
		Stream: true,
		OnChunk: func(c Chunk) error {
			chunks = append(chunks, c)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("OnChunk called %d times, want 1", len(chunks))
	}
	if resp.DataClassification != "internal" {
		t.Errorf("DataClassification = %q, want %q", resp.DataClassification, "internal")
	}
}

func TestComplete_StreamEmptyLines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("\n\n"))
		_, _ = w.Write([]byte(`{"type":"content","content":"data"}` + "\n"))
		_, _ = w.Write([]byte("   \n"))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	resp, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{Stream: true})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "data" {
		t.Errorf("Content = %q, want %q", resp.Content, "data")
	}
}

func TestComplete_BufferedEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"stub","choices":[]}`))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	resp, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "" {
		t.Errorf("Content = %q, want empty", resp.Content)
	}
}

func TestComplete_BufferedWithFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"fallback-model","choices":[{"message":{"content":"ok"}}],"fallback":{"used":true},"data_classification":"confidential"}`))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	resp, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if !resp.FallbackUsed {
		t.Error("FallbackUsed = false, want true")
	}
	if resp.Model != "fallback-model" {
		t.Errorf("Model = %q, want %q", resp.Model, "fallback-model")
	}
	if resp.DataClassification != "confidential" {
		t.Errorf("DataClassification = %q, want %q", resp.DataClassification, "confidential")
	}
}

func TestComplete_WithDataClassificationHeader(t *testing.T) {
	var gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Data-Classification")
		_, _ = w.Write([]byte(`{"model":"stub","choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{
		DataClassification: "internal",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if gotHeader != "internal" {
		t.Errorf("X-Data-Classification header = %q, want %q", gotHeader, "internal")
	}
}

func TestComplete_WithCustomModelAndTemperature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"custom","choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	client := NewHTTPLLMClient(server.URL, server.Client())
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}}, CompletionOptions{
		Model:       "custom-model",
		Temperature: 0.9,
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{name: "first non-empty", values: []string{"", "hello", "world"}, want: "hello"},
		{name: "all empty", values: []string{"", "", ""}, want: ""},
		{name: "whitespace only skipped", values: []string{"  ", "real"}, want: "real"},
		{name: "no values", values: nil, want: ""},
		{name: "first value wins", values: []string{"a", "b"}, want: "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonEmpty(tt.values...)
			if got != tt.want {
				t.Errorf("firstNonEmpty() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNonZero(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		fallback float64
		want     float64
	}{
		{name: "non-zero returns value", value: 0.5, fallback: 0.7, want: 0.5},
		{name: "zero returns fallback", value: 0, fallback: 0.7, want: 0.7},
		{name: "negative is non-zero", value: -1.0, fallback: 0.7, want: -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nonZero(tt.value, tt.fallback)
			if got != tt.want {
				t.Errorf("nonZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
