package trading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenScanner discovers candidates and refreshes token metadata.
type TokenScanner interface {
	DiscoverCandidates(ctx context.Context) ([]TokenInfo, error)
	RefreshToken(ctx context.Context, mint string) (TokenInfo, error)
}

// JupiterScanner uses Jupiter Tokens V2 as the discovery source.
type JupiterScanner struct {
	baseURL    string
	apiKey     string
	category   string
	interval   string
	limit      int
	httpClient *http.Client
}

// NewJupiterScanner creates a new discovery source using Jupiter Tokens V2.
func NewJupiterScanner(baseURL, apiKey, category, interval string, limit int) *JupiterScanner {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.jup.ag"
	}
	if category == "" {
		category = "toptrending"
	}
	if interval == "" {
		interval = "5m"
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return &JupiterScanner{
		baseURL:  strings.TrimRight(baseURL, "/"),
		apiKey:   strings.TrimSpace(apiKey),
		category: category,
		interval: interval,
		limit:    limit,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// DiscoverCandidates returns the currently trending candidate list.
func (s *JupiterScanner) DiscoverCandidates(ctx context.Context) ([]TokenInfo, error) {
	if s == nil {
		return nil, errors.New("scanner is nil")
	}
	path := fmt.Sprintf("%s/tokens/v2/%s/%s?limit=%d", s.baseURL, url.PathEscape(s.category), url.PathEscape(s.interval), s.limit)
	return s.fetchTokens(ctx, path)
}

// RefreshToken reloads the latest token metadata by mint.
func (s *JupiterScanner) RefreshToken(ctx context.Context, mint string) (TokenInfo, error) {
	if s == nil {
		return TokenInfo{}, errors.New("scanner is nil")
	}
	mint = strings.TrimSpace(mint)
	if mint == "" {
		return TokenInfo{}, errors.New("mint is required")
	}

	path := fmt.Sprintf("%s/tokens/v2/search?query=%s", s.baseURL, url.QueryEscape(mint))
	tokens, err := s.fetchTokens(ctx, path)
	if err != nil {
		return TokenInfo{}, err
	}
	for _, token := range tokens {
		if token.Mint == mint {
			return token, nil
		}
	}
	if len(tokens) > 0 {
		return tokens[0], nil
	}
	return TokenInfo{}, fmt.Errorf("token %s not found", mint)
}

func (s *JupiterScanner) fetchTokens(ctx context.Context, path string) ([]TokenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("build Jupiter request: %w", err)
	}
	if s.apiKey != "" {
		req.Header.Set("x-api-key", s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Jupiter tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jupiter token request failed: %s", resp.Status)
	}

	var out []TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode Jupiter tokens response: %w", err)
	}
	return out, nil
}
