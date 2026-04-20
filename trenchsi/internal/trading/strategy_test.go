package trading

import "testing"

func TestBasicStrategyRejectsLowLiquidity(t *testing.T) {
	strategy := NewBasicStrategy(BasicStrategyConfig{
		MinLiquidityUSD: 1000,
		MinVolumeUSD:    100,
	})

	signal, err := strategy.EvaluateToken(TokenInfo{
		Mint:         "mint",
		Name:         "Token",
		Symbol:       "TOK",
		Decimals:     6,
		LiquidityUSD: 500,
		Stats24h:     TokenStats{BuyVolumeUSD: 1000},
	})
	if err != nil {
		t.Fatalf("EvaluateToken() error: %v", err)
	}
	if signal.Action != SignalSkip {
		t.Fatalf("expected skip, got %s", signal.Action)
	}
}

func TestBasicStrategyRejectsMissingMetadata(t *testing.T) {
	strategy := NewBasicStrategy(BasicStrategyConfig{})
	signal, err := strategy.EvaluateToken(TokenInfo{Mint: "mint"})
	if err != nil {
		t.Fatalf("EvaluateToken() error: %v", err)
	}
	if signal.Action != SignalSkip {
		t.Fatalf("expected skip, got %s", signal.Action)
	}
}

func TestBasicStrategyAcceptsHealthyToken(t *testing.T) {
	strategy := NewBasicStrategy(BasicStrategyConfig{
		MinLiquidityUSD:           1000,
		MinVolumeUSD:              1000,
		MaxVolatilityPct:          40,
		MinOrganicScore:           20,
		MaxHolderConcentrationPct: 50,
	})

	signal, err := strategy.EvaluateToken(TokenInfo{
		Mint:         "mint",
		Name:         "Token",
		Symbol:       "TOK",
		Decimals:     6,
		LiquidityUSD: 10000,
		OrganicScore: 35,
		Audit:        TokenAudit{TopHoldersPercentage: 20},
		Stats24h:     TokenStats{PriceChangePct: 12, BuyVolumeUSD: 5000, SellVolumeUSD: 3000},
	})
	if err != nil {
		t.Fatalf("EvaluateToken() error: %v", err)
	}
	if signal.Action != SignalBuy {
		t.Fatalf("expected buy, got %s", signal.Action)
	}
	if signal.Score <= 0 {
		t.Fatal("expected positive score")
	}
}
