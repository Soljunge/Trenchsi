package trading

import (
	"testing"
	"time"
)

func TestPositionCalculations(t *testing.T) {
	pos := Position{
		Decimals:      6,
		QuantityRaw:   2_000_000,
		EntryValueSOL: 1.0,
		EntryPriceUSD: 0.5,
	}

	if got := pos.CurrentQuantityUI(); got != 2 {
		t.Fatalf("unexpected quantity: %v", got)
	}
	if got := pos.AverageEntryPriceSOL(); got != 0.5 {
		t.Fatalf("unexpected average entry: %v", got)
	}
	if got := pos.ExposureSOL(0.75); got != 1.5 {
		t.Fatalf("unexpected exposure: %v", got)
	}
	if got := pos.UnrealizedPnLSOL(0.75); got != 0.5 {
		t.Fatalf("unexpected pnl: %v", got)
	}
}

func TestPortfolioDuplicateEntriesBlocked(t *testing.T) {
	portfolio := NewPortfolio(false)
	err := portfolio.UpsertPosition(&Position{
		Mint:      "mint",
		Status:    PositionOpen,
		OpenedAt:  time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("UpsertPosition() error: %v", err)
	}
	err = portfolio.UpsertPosition(&Position{
		Mint:      "mint",
		Status:    PositionOpen,
		OpenedAt:  time.Now(),
		UpdatedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("expected duplicate entry to be blocked")
	}
}
