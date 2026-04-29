package openrouter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Model struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ContextLength int     `json:"context_length"`
	InputCost     float64
	OutputCost    float64
}

type modelsAPIResponse struct {
	Data []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ContextLength int    `json:"context_length"`
		Pricing       struct {
			Prompt     string `json:"prompt"`
			Completion string `json:"completion"`
		} `json:"pricing"`
	} `json:"data"`
}

func (c *Client) FetchModels(apiKey string) ([]Model, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter returned status %d", resp.StatusCode)
	}

	var body modelsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("unexpected response: %w", err)
	}

	models := make([]Model, 0, len(body.Data))
	for _, d := range body.Data {
		m := Model{
			ID:            d.ID,
			Name:          d.Name,
			ContextLength: d.ContextLength,
		}
		fmt.Sscanf(d.Pricing.Prompt, "%f", &m.InputCost)
		fmt.Sscanf(d.Pricing.Completion, "%f", &m.OutputCost)
		models = append(models, m)
	}
	return models, nil
}

// FilterModels filters by case-insensitive substring match on ID or Name.
func FilterModels(models []Model, query string) []Model {
	q := strings.ToLower(query)
	var out []Model
	for _, m := range models {
		if strings.Contains(strings.ToLower(m.ID), q) || strings.Contains(strings.ToLower(m.Name), q) {
			out = append(out, m)
		}
	}
	return out
}
