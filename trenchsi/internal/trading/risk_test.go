package trading

import (
	"testing"
	"time"
)

func TestRiskManagerBlocksWhenBalanceBelowReserve(t *testing.T) {
	risk := NewRiskManager(RiskLimits{
		MaxPositionSOL:   0.25,
		ReserveSOL:       0.5,
		MaxOpenPositions: 1,
	})

	portfolio := NewPortfolio(false)
	err := risk.CanOpenPosition(time.Now(), portfolio, 0.4, &QuoteView{PriceImpactPct: 1}, TokenInfo{Mint: "mint", Name: "Token", Symbol: "TOK", Decimals: 6})
	if err == nil {
		t.Fatal("expected reserve check to fail")
	}
}

func TestRiskManagerBlocksCooldown(t *testing.T) {
	risk := NewRiskManager(RiskLimits{
		TradeCooldownSeconds: 60,
		ReserveSOL:           0.1,
		MaxPositionSOL:       0.1,
	})
	portfolio := NewPortfolio(false)
	portfolio.RecordTrade(time.Now().Add(-30 * time.Second))
	err := risk.CanOpenPosition(time.Now(), portfolio, 1.0, &QuoteView{}, TokenInfo{Mint: "mint", Name: "Token", Symbol: "TOK", Decimals: 6})
	if err == nil {
		t.Fatal("expected cooldown to block entry")
	}
}

func TestRiskManagerShouldExitOnStopLoss(t *testing.T) {
	risk := NewRiskManager(RiskLimits{StopLossPct: 10})
	exit, reason := risk.ShouldExitPosition(time.Now(), Position{
		Mint:            "mint",
		Status:          PositionOpen,
		EntryPriceUSD:   1.0,
		CurrentPriceUSD: 0.89,
	}, TokenInfo{USDPrice: 0.89})
	if !exit || reason == "" {
		t.Fatal("expected stop loss exit")
	}
}

func TestRiskManagerValidatesQuoteAge(t *testing.T) {
	risk := NewRiskManager(RiskLimits{MaxQuoteAgeSlots: 5})
	err := risk.ValidateQuoteAge(&QuoteView{ContextSlot: 10}, 20)
	if err == nil {
		t.Fatal("expected stale quote to fail")
	}
}
