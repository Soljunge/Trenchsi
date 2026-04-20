package trading

import (
	"errors"
	"fmt"
	"math"
)

// Strategy evaluates a token and decides whether it is worth trenching.
type Strategy interface {
	EvaluateToken(token TokenInfo) (StrategySignal, error)
}

// BasicStrategyConfig controls the default momentum/liquidity screen.
type BasicStrategyConfig struct {
	MinLiquidityUSD           float64
	MinVolumeUSD              float64
	MaxVolatilityPct          float64
	MinMarketCapUSD           float64
	MaxMarketCapUSD           float64
	MinOrganicScore           float64
	MaxHolderConcentrationPct float64
	MomentumWeight            float64
	VolumeWeight              float64
	LiquidityWeight           float64
}

// BasicStrategy is a configurable, conservative trenching strategy.
type BasicStrategy struct {
	cfg BasicStrategyConfig
}

// NewBasicStrategy creates the default screen used by the runner.
func NewBasicStrategy(cfg BasicStrategyConfig) *BasicStrategy {
	if cfg.MomentumWeight == 0 && cfg.VolumeWeight == 0 && cfg.LiquidityWeight == 0 {
		cfg.MomentumWeight = 0.45
		cfg.VolumeWeight = 0.30
		cfg.LiquidityWeight = 0.25
	}
	return &BasicStrategy{cfg: cfg}
}

// EvaluateToken returns a buy signal only when the token passes conservative filters.
func (s *BasicStrategy) EvaluateToken(token TokenInfo) (StrategySignal, error) {
	if token.Mint == "" {
		return StrategySignal{Action: SignalSkip, Reason: "missing mint"}, errors.New("token mint is required")
	}
	if !token.HasMetadata() {
		return StrategySignal{Action: SignalSkip, Reason: "missing metadata"}, nil
	}

	if s.cfg.MinLiquidityUSD > 0 && token.LiquidityUSD < s.cfg.MinLiquidityUSD {
		return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("liquidity below threshold: %.2f < %.2f", token.LiquidityUSD, s.cfg.MinLiquidityUSD)}, nil
	}
	if s.cfg.MinVolumeUSD > 0 && token.VolumeScore() < s.cfg.MinVolumeUSD {
		return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("volume below threshold: %.2f < %.2f", token.VolumeScore(), s.cfg.MinVolumeUSD)}, nil
	}
	if s.cfg.MaxVolatilityPct > 0 && math.Abs(token.MomentumScore()) > s.cfg.MaxVolatilityPct {
		return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("volatility above threshold: %.2f > %.2f", math.Abs(token.MomentumScore()), s.cfg.MaxVolatilityPct)}, nil
	}
	if s.cfg.MinMarketCapUSD > 0 && token.MarketCapUSD > 0 && token.MarketCapUSD < s.cfg.MinMarketCapUSD {
		return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("market cap below threshold: %.2f < %.2f", token.MarketCapUSD, s.cfg.MinMarketCapUSD)}, nil
	}
	if s.cfg.MaxMarketCapUSD > 0 && token.MarketCapUSD > s.cfg.MaxMarketCapUSD {
		return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("market cap above threshold: %.2f > %.2f", token.MarketCapUSD, s.cfg.MaxMarketCapUSD)}, nil
	}
	if s.cfg.MinOrganicScore > 0 && token.OrganicScore > 0 && token.OrganicScore < s.cfg.MinOrganicScore {
		return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("organic score below threshold: %.2f < %.2f", token.OrganicScore, s.cfg.MinOrganicScore)}, nil
	}
	if s.cfg.MaxHolderConcentrationPct > 0 {
		if hc := token.HolderConcentration(); hc > 0 && hc > s.cfg.MaxHolderConcentrationPct {
			return StrategySignal{Action: SignalSkip, Reason: fmt.Sprintf("holder concentration above threshold: %.2f > %.2f", hc, s.cfg.MaxHolderConcentrationPct)}, nil
		}
	}

	score := s.scoreToken(token)
	confidence := math.Max(0, math.Min(1, score/100))
	return StrategySignal{
		Action:               SignalBuy,
		Score:                score,
		Reason:               "token passed the conservative screen",
		Confidence:           confidence,
		SuggestedPositionSOL: 0,
	}, nil
}

func (s *BasicStrategy) scoreToken(token TokenInfo) float64 {
	momentum := normalizeMomentum(token.MomentumScore())
	volume := normalizePositive(token.VolumeScore())
	liquidity := normalizePositive(token.LiquidityScore())

	return momentum*s.cfg.MomentumWeight*100 +
		volume*s.cfg.VolumeWeight*100 +
		liquidity*s.cfg.LiquidityWeight*100
}

func normalizeMomentum(value float64) float64 {
	if value == 0 {
		return 0
	}
	abs := math.Abs(value)
	score := abs / 100
	if score > 1 {
		score = 1
	}
	if value < 0 {
		score *= 0.35
	}
	return score
}

func normalizePositive(value float64) float64 {
	if value <= 0 {
		return 0
	}
	score := math.Log10(value + 1)
	maxScore := math.Log10(1_000_000 + 1)
	if maxScore <= 0 {
		return 0
	}
	score /= maxScore
	if score > 1 {
		score = 1
	}
	return score
}
