package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sipeed/trenchlaw/web/backend/utils"
)

var launcherAccessToken string

func initLauncherAccessToken() error {
	token, err := loadOrCreateLauncherAccessToken()
	if err != nil {
		return err
	}
	launcherAccessToken = token
	return nil
}

func loadOrCreateLauncherAccessToken() (string, error) {
	tokenPath := filepath.Join(utils.GetJameclawHome(), "launcher-access-token")

	if data, err := os.ReadFile(tokenPath); err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			return token, nil
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("read launcher access token: %w", err)
	}

	token, err := generateLauncherAccessToken()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(tokenPath), 0o755); err != nil {
		return "", fmt.Errorf("create launcher access token directory: %w", err)
	}
	if err := os.WriteFile(tokenPath, []byte(token+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write launcher access token: %w", err)
	}

	return token, nil
}

func generateLauncherAccessToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate launcher access token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func launcherOpenURL(baseURL string) string {
	if baseURL == "" || launcherAccessToken == "" {
		return baseURL
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	query := parsed.Query()
	query.Set("access_token", launcherAccessToken)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
