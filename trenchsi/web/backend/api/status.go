package api

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/sipeed/trenchlaw/pkg/config"
)

type appStatusResponse struct {
	Status          string `json:"status"`
	Version         string `json:"version"`
	Uptime          string `json:"uptime"`
	GitCommit       string `json:"git_commit,omitempty"`
	RepoHeadCommit  string `json:"repo_head_commit,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	UpdateMessage   string `json:"update_message,omitempty"`
}

func (h *Handler) registerStatusRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/status", h.handleStatus)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	currentCommit := currentBuildCommit()
	repoHeadCommit := detectLocalRepoHeadCommit()

	resp := appStatusResponse{
		Status:         "ok",
		Version:        config.GetVersion(),
		Uptime:         timeSince(h.startTime),
		GitCommit:      currentCommit,
		RepoHeadCommit: repoHeadCommit,
	}

	if currentCommit != "" && repoHeadCommit != "" && currentCommit != repoHeadCommit {
		resp.UpdateAvailable = true
		resp.UpdateMessage = "Local git HEAD is newer than the running build."
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func timeSince(startTime time.Time) string {
	if startTime.IsZero() {
		return "unknown"
	}
	return time.Since(startTime).Round(time.Second).String()
}

func currentBuildCommit() string {
	if commit := normalizeCommit(config.GitCommit); commit != "" {
		return commit
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return normalizeCommit(setting.Value)
			}
		}
	}

	return ""
}

func detectLocalRepoHeadCommit() string {
	repoRoot := detectRepoRoot()
	if repoRoot == "" {
		return ""
	}

	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--short=8", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return normalizeCommit(string(output))
}

func detectRepoRoot() string {
	for _, candidate := range repoRootCandidates() {
		root := findRepoRoot(candidate)
		if root != "" {
			return root
		}
	}
	return ""
}

func repoRootCandidates() []string {
	candidates := make([]string, 0, 4)

	if envRoot := strings.TrimSpace(os.Getenv("TRENCHSI_REPO_ROOT")); envRoot != "" {
		candidates = append(candidates, envRoot)
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}

	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(exePath))
	}

	return candidates
}

func findRepoRoot(start string) string {
	current := filepath.Clean(start)
	for {
		if isTrenchsiRepoRoot(current) {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func isTrenchsiRepoRoot(path string) bool {
	if path == "" {
		return false
	}

	if stat, err := os.Stat(filepath.Join(path, ".git")); err != nil || !stat.IsDir() {
		return false
	}

	data, err := os.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		return false
	}

	return strings.Contains(string(data), "module github.com/sipeed/trenchlaw")
}

func normalizeCommit(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) > 8 {
		return trimmed[:8]
	}
	return trimmed
}
