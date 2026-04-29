package openrouter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateKeySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[{"id":"openai/gpt-4"}]}`))
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL}
	if err := c.ValidateKey("good-key"); err != nil {
		t.Errorf("expected valid key to pass, got: %v", err)
	}
}

func TestValidateKeyFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid key"}`))
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL}
	if err := c.ValidateKey("bad-key"); err == nil {
		t.Error("expected invalid key to fail")
	}
}

func TestNewClientDefaultBaseURL(t *testing.T) {
	c := NewClient()
	if c.baseURL != "https://openrouter.ai/api" {
		t.Errorf("unexpected baseURL: %s", c.baseURL)
	}
}
