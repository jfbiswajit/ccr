package openrouter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const modelsResponse = `{
  "data": [
    {"id": "anthropic/claude-sonnet-4-6", "name": "Claude Sonnet 4.6", "context_length": 200000, "pricing": {"prompt": "0.000003", "completion": "0.000015"}},
    {"id": "openai/gpt-4o", "name": "GPT-4o", "context_length": 128000, "pricing": {"prompt": "0.0000025", "completion": "0.00001"}},
    {"id": "google/gemini-pro", "name": "Gemini Pro", "context_length": 32000, "pricing": {"prompt": "0.0000005", "completion": "0.0000015"}}
  ]
}`

func TestFetchModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(modelsResponse))
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL}
	models, err := c.FetchModels("any-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 3 {
		t.Errorf("expected 3 models, got %d", len(models))
	}
	if models[0].ID != "anthropic/claude-sonnet-4-6" {
		t.Errorf("unexpected first model ID: %s", models[0].ID)
	}
}

func TestFilterModels(t *testing.T) {
	models := []Model{
		{ID: "anthropic/claude-sonnet-4-6", Name: "Claude Sonnet"},
		{ID: "openai/gpt-4o", Name: "GPT-4o"},
		{ID: "anthropic/claude-opus-4", Name: "Claude Opus"},
	}

	filtered := FilterModels(models, "claude")
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered models, got %d", len(filtered))
	}
	for _, m := range filtered {
		if !strings.Contains(strings.ToLower(m.ID+m.Name), "claude") {
			t.Errorf("unexpected model in filtered results: %s", m.ID)
		}
	}
}

func TestFilterModelsCaseInsensitive(t *testing.T) {
	models := []Model{
		{ID: "anthropic/claude-sonnet", Name: "Claude Sonnet"},
		{ID: "openai/gpt-4", Name: "GPT-4"},
	}
	filtered := FilterModels(models, "CLAUDE")
	if len(filtered) != 1 {
		t.Errorf("expected 1 result, got %d", len(filtered))
	}
}

func TestFilterModelsNoMatch(t *testing.T) {
	models := []Model{
		{ID: "anthropic/claude-sonnet", Name: "Claude Sonnet"},
	}
	filtered := FilterModels(models, "llama")
	if len(filtered) != 0 {
		t.Errorf("expected 0 results, got %d", len(filtered))
	}
}
