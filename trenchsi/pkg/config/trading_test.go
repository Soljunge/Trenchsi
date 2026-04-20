package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultTradingConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Trading.Strategy != "solana-trench" {
		t.Fatalf("unexpected default strategy: %s", cfg.Trading.Strategy)
	}
	if !cfg.Trading.DryRun {
		t.Fatal("default trading config should be dry-run safe")
	}
	if !cfg.Trading.EnablePaperTrading {
		t.Fatal("default trading config should enable paper trading")
	}
	if cfg.Trading.MaxPositionSOL <= 0 {
		t.Fatal("default max position must be positive")
	}
}

func TestLoadConfigAppliesTradingEnv(t *testing.T) {
	t.Setenv("SOLANA_NETWORK", "devnet")
	t.Setenv("SOLANA_RPC_URL", "https://example.invalid")
	t.Setenv("SOLANA_WS_URL", "wss://example.invalid")
	t.Setenv("MAX_POSITION_SOL", "0.5")
	t.Setenv("MAX_DAILY_LOSS_SOL", "1.25")
	t.Setenv("MAX_OPEN_POSITIONS", "3")
	t.Setenv("DEFAULT_SLIPPAGE_BPS", "75")
	t.Setenv("TAKE_PROFIT_PCT", "18")
	t.Setenv("STOP_LOSS_PCT", "9")
	t.Setenv("TRAILING_STOP_PCT", "6")
	t.Setenv("DRY_RUN", "true")
	t.Setenv("ENABLE_PAPER_TRADING", "false")
	t.Setenv("TRADING_POLL_INTERVAL_SECONDS", "42")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.Trading.Network != "devnet" {
		t.Fatalf("unexpected trading network: %s", cfg.Trading.Network)
	}
	if cfg.Trading.RPCURL != "https://example.invalid" {
		t.Fatalf("unexpected rpc url: %s", cfg.Trading.RPCURL)
	}
	if cfg.Trading.WSURL != "wss://example.invalid" {
		t.Fatalf("unexpected ws url: %s", cfg.Trading.WSURL)
	}
	if cfg.Trading.MaxPositionSOL != 0.5 {
		t.Fatalf("unexpected max position: %v", cfg.Trading.MaxPositionSOL)
	}
	if cfg.Trading.MaxDailyLossSOL != 1.25 {
		t.Fatalf("unexpected max daily loss: %v", cfg.Trading.MaxDailyLossSOL)
	}
	if cfg.Trading.MaxOpenPositions != 3 {
		t.Fatalf("unexpected open positions limit: %d", cfg.Trading.MaxOpenPositions)
	}
	if cfg.Trading.DefaultSlippageBps != 75 {
		t.Fatalf("unexpected slippage: %d", cfg.Trading.DefaultSlippageBps)
	}
	if !cfg.Trading.DryRun {
		t.Fatal("dry-run env should be applied")
	}
	if cfg.Trading.EnablePaperTrading {
		t.Fatal("paper trading env should be applied")
	}
	if cfg.Trading.PollIntervalSeconds != 42 {
		t.Fatalf("unexpected poll interval: %d", cfg.Trading.PollIntervalSeconds)
	}
}
