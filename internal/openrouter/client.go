package openrouter

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://openrouter.ai/api",
		http:    &http.Client{},
	}
}

func (c *Client) httpClient() *http.Client {
	if c.http != nil {
		return c.http
	}
	return http.DefaultClient
}

func (c *Client) ValidateKey(apiKey string) error {
	req, err := http.NewRequest("GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("could not reach OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenRouter returned status %d", resp.StatusCode)
	}

	var body struct {
		Data []interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("unexpected response from OpenRouter: %w", err)
	}
	if len(body.Data) == 0 {
		return fmt.Errorf("API key valid but no models returned")
	}
	return nil
}
