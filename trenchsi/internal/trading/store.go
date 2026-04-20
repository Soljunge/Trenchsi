package trading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sipeed/trenchlaw/pkg/fileutil"
)

// PortfolioStore persists portfolio snapshots on disk.
type PortfolioStore struct {
	mu   sync.Mutex
	path string
}

// NewPortfolioStore creates a file-backed portfolio store.
func NewPortfolioStore(workspace string) *PortfolioStore {
	dir := filepath.Join(workspace, "state")
	return &PortfolioStore{
		path: filepath.Join(dir, "solana-trading.json"),
	}
}

// Load reads the portfolio snapshot from disk.
func (s *PortfolioStore) Load(ctx context.Context) (PortfolioSnapshot, error) {
	var snapshot PortfolioSnapshot
	if s == nil {
		return snapshot, errors.New("portfolio store is nil")
	}
	select {
	case <-ctx.Done():
		return snapshot, ctx.Err()
	default:
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			snapshot.Positions = map[string]*Position{}
			return snapshot, nil
		}
		return snapshot, fmt.Errorf("read portfolio snapshot: %w", err)
	}

	if err := json.Unmarshal(data, &snapshot); err != nil {
		return snapshot, fmt.Errorf("decode portfolio snapshot: %w", err)
	}
	if snapshot.Positions == nil {
		snapshot.Positions = map[string]*Position{}
	}
	return snapshot, nil
}

// Save writes the portfolio snapshot atomically.
func (s *PortfolioStore) Save(ctx context.Context, snapshot PortfolioSnapshot) error {
	if s == nil {
		return errors.New("portfolio store is nil")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create portfolio state directory: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal portfolio snapshot: %w", err)
	}

	if err := fileutil.WriteFileAtomic(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write portfolio snapshot: %w", err)
	}
	return nil
}
