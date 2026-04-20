package trading

import (
	"math"
	"strconv"
	"time"
)

// SignalAction describes what the strategy wants to do.
type SignalAction string

const (
	SignalHold SignalAction = "hold"
	SignalBuy  SignalAction = "buy"
	SignalSell SignalAction = "sell"
	SignalSkip SignalAction = "skip"
)

// PositionStatus captures the lifecycle of a tracked trade.
type PositionStatus string

const (
	PositionOpen   PositionStatus = "open"
	PositionClosed PositionStatus = "closed"
)

// TokenStats stores a subset of Jupiter token activity metrics.
type TokenStats struct {
	PriceChangePct     float64 `json:"priceChange"`
	BuyVolumeUSD       float64 `json:"buyVolume,omitempty"`
	SellVolumeUSD      float64 `json:"sellVolume,omitempty"`
	VolumeChangePct    float64 `json:"volumeChange,omitempty"`
	LiquidityChangePct float64 `json:"liquidityChange,omitempty"`
	HolderChangePct    float64 `json:"holderChange,omitempty"`
	NumBuys            int64   `json:"numBuys,omitempty"`
	NumSells           int64   `json:"numSells,omitempty"`
	NumTraders         int64   `json:"numTraders,omitempty"`
}

// TokenAudit stores the token audit block returned by Jupiter.
type TokenAudit struct {
	MintAuthorityDisabled   bool    `json:"mintAuthorityDisabled,omitempty"`
	FreezeAuthorityDisabled bool    `json:"freezeAuthorityDisabled,omitempty"`
	TopHoldersPercentage    float64 `json:"topHoldersPercentage,omitempty"`
	DevMints                int64   `json:"devMints,omitempty"`
}

// TokenInfo holds the scanner and strategy view of a Solana token.
type TokenInfo struct {
	Mint              string     `json:"id"`
	Name              string     `json:"name"`
	Symbol            string     `json:"symbol"`
	Icon              string     `json:"icon,omitempty"`
	Decimals          int        `json:"decimals,omitempty"`
	TokenProgram      string     `json:"tokenProgram,omitempty"`
	CreatedAt         time.Time  `json:"createdAt,omitempty"`
	UpdatedAt         time.Time  `json:"updatedAt,omitempty"`
	DevWallet         string     `json:"dev,omitempty"`
	HolderCount       int64      `json:"holderCount,omitempty"`
	FDVUSD            float64    `json:"fdv,omitempty"`
	MarketCapUSD      float64    `json:"mcap,omitempty"`
	USDPrice          float64    `json:"usdPrice,omitempty"`
	LiquidityUSD      float64    `json:"liquidity,omitempty"`
	FeesUSD           float64    `json:"fees,omitempty"`
	OrganicScore      float64    `json:"organicScore,omitempty"`
	OrganicScoreLabel string     `json:"organicScoreLabel,omitempty"`
	IsVerified        bool       `json:"isVerified,omitempty"`
	Tags              []string   `json:"tags,omitempty"`
	Audit             TokenAudit `json:"audit,omitempty"`
	Stats5m           TokenStats `json:"stats5m,omitempty"`
	Stats1h           TokenStats `json:"stats1h,omitempty"`
	Stats6h           TokenStats `json:"stats6h,omitempty"`
	Stats24h          TokenStats `json:"stats24h,omitempty"`
	Stats7d           TokenStats `json:"stats7d,omitempty"`
	Stats30d          TokenStats `json:"stats30d,omitempty"`
}

// HasMetadata returns true when the token has enough metadata for basic screening.
func (t TokenInfo) HasMetadata() bool {
	return t.Mint != "" && t.Name != "" && t.Symbol != "" && t.Decimals >= 0
}

// HolderConcentration returns the strongest holder concentration metric available.
func (t TokenInfo) HolderConcentration() float64 {
	if t.Audit.TopHoldersPercentage > 0 {
		return t.Audit.TopHoldersPercentage
	}
	return 0
}

// MomentumScore returns the strongest momentum metric available.
func (t TokenInfo) MomentumScore() float64 {
	if t.Stats5m.PriceChangePct != 0 {
		return t.Stats5m.PriceChangePct
	}
	if t.Stats1h.PriceChangePct != 0 {
		return t.Stats1h.PriceChangePct
	}
	return t.Stats24h.PriceChangePct
}

// VolumeScore returns a liquid volume proxy for ranking.
func (t TokenInfo) VolumeScore() float64 {
	total := t.Stats24h.BuyVolumeUSD + t.Stats24h.SellVolumeUSD
	if total > 0 {
		return total
	}
	return float64(t.Stats24h.NumBuys + t.Stats24h.NumSells)
}

// LiquidityScore returns the reported liquidity.
func (t TokenInfo) LiquidityScore() float64 {
	return t.LiquidityUSD
}

// StrategySignal is the strategy's output for a token.
type StrategySignal struct {
	Action               SignalAction `json:"action"`
	Score                float64      `json:"score"`
	Reason               string       `json:"reason"`
	Confidence           float64      `json:"confidence"`
	SuggestedPositionSOL float64      `json:"suggested_position_sol,omitempty"`
}

// TradeRequest describes a requested trade.
type TradeRequest struct {
	Side              SignalAction   `json:"side"`
	Token             TokenInfo      `json:"token"`
	InputMint         string         `json:"input_mint"`
	OutputMint        string         `json:"output_mint"`
	InputAmountRaw    uint64         `json:"input_amount_raw"`
	InputAmountSOL    float64        `json:"input_amount_sol"`
	SlippageBps       int            `json:"slippage_bps"`
	MaxPriceImpactPct float64        `json:"max_price_impact_pct"`
	DryRun            bool           `json:"dry_run"`
	PaperTrading      bool           `json:"paper_trading"`
	Reason            string         `json:"reason,omitempty"`
	ExpectedQuote     *QuoteView     `json:"expected_quote,omitempty"`
	StrategySignal    StrategySignal `json:"strategy_signal"`
}

// QuoteView is the subset of a Jupiter quote that the trading engine needs.
type QuoteView struct {
	InputMint            string  `json:"inputMint"`
	OutputMint           string  `json:"outputMint"`
	InAmount             string  `json:"inAmount"`
	OutAmount            string  `json:"outAmount"`
	OtherAmountThreshold string  `json:"otherAmountThreshold"`
	PriceImpactPct       float64 `json:"priceImpactPct"`
	ContextSlot          uint64  `json:"contextSlot"`
	SlippageBps          int     `json:"slippageBps"`
}

// InAmountRaw returns the raw input amount.
func (q QuoteView) InAmountRaw() uint64 {
	value, _ := strconv.ParseUint(q.InAmount, 10, 64)
	return value
}

// OutAmountRaw returns the raw output amount.
func (q QuoteView) OutAmountRaw() uint64 {
	value, _ := strconv.ParseUint(q.OutAmount, 10, 64)
	return value
}

// TradeResult captures the outcome of a trade request.
type TradeResult struct {
	Success        bool          `json:"success"`
	Side           SignalAction  `json:"side"`
	TokenMint      string        `json:"token_mint"`
	Signature      string        `json:"signature,omitempty"`
	ExpectedOutRaw uint64        `json:"expected_out_raw,omitempty"`
	ActualOutRaw   uint64        `json:"actual_out_raw,omitempty"`
	ExpectedOutSOL float64       `json:"expected_out_sol,omitempty"`
	ActualOutSOL   float64       `json:"actual_out_sol,omitempty"`
	InputAmountRaw uint64        `json:"input_amount_raw,omitempty"`
	InputAmountSOL float64       `json:"input_amount_sol,omitempty"`
	PriceImpactPct float64       `json:"price_impact_pct,omitempty"`
	ExecutedAt     time.Time     `json:"executed_at,omitempty"`
	Duration       time.Duration `json:"duration,omitempty"`
	PaperTrading   bool          `json:"paper_trading,omitempty"`
	DryRun         bool          `json:"dry_run,omitempty"`
	Error          string        `json:"error,omitempty"`
	EntryPriceSOL  float64       `json:"entry_price_sol,omitempty"`
	ExitPriceSOL   float64       `json:"exit_price_sol,omitempty"`
}

// RiskLimits configures trade and portfolio risk controls.
type RiskLimits struct {
	MaxPositionSOL            float64
	MaxDailyLossSOL           float64
	MaxOpenPositions          int
	DefaultSlippageBps        int
	TakeProfitPct             float64
	StopLossPct               float64
	TrailingStopPct           float64
	TradeCooldownSeconds      int
	MaxHoldMinutes            int
	ReserveSOL                float64
	EnablePaperTrading        bool
	EnableSells               bool
	DryRun                    bool
	MaxQuoteAgeSlots          int
	MaxPriceImpactPct         float64
	MaxVolatilityPct          float64
	MinLiquidityUSD           float64
	MinVolumeUSD              float64
	MinMarketCapUSD           float64
	MaxMarketCapUSD           float64
	MinOrganicScore           float64
	MaxHolderConcentrationPct float64
	AllowDuplicateEntries     bool
	EmergencyHalt             bool
}

// Validate normalizes and clamps the risk limits into safe defaults.
func (r RiskLimits) Validate() RiskLimits {
	if r.MaxPositionSOL < 0 {
		r.MaxPositionSOL = 0
	}
	if r.MaxDailyLossSOL < 0 {
		r.MaxDailyLossSOL = 0
	}
	if r.MaxOpenPositions < 0 {
		r.MaxOpenPositions = 0
	}
	if r.DefaultSlippageBps < 0 {
		r.DefaultSlippageBps = 0
	}
	if r.TradeCooldownSeconds < 0 {
		r.TradeCooldownSeconds = 0
	}
	if r.MaxHoldMinutes < 0 {
		r.MaxHoldMinutes = 0
	}
	if r.ReserveSOL < 0 {
		r.ReserveSOL = 0
	}
	if r.MaxQuoteAgeSlots < 0 {
		r.MaxQuoteAgeSlots = 0
	}
	if r.MaxPriceImpactPct < 0 {
		r.MaxPriceImpactPct = 0
	}
	if r.MaxVolatilityPct < 0 {
		r.MaxVolatilityPct = 0
	}
	if r.MinLiquidityUSD < 0 {
		r.MinLiquidityUSD = 0
	}
	if r.MinVolumeUSD < 0 {
		r.MinVolumeUSD = 0
	}
	if r.MinMarketCapUSD < 0 {
		r.MinMarketCapUSD = 0
	}
	if r.MaxMarketCapUSD < 0 {
		r.MaxMarketCapUSD = 0
	}
	if r.MinOrganicScore < 0 {
		r.MinOrganicScore = 0
	}
	if r.MaxHolderConcentrationPct < 0 {
		r.MaxHolderConcentrationPct = 0
	}
	return r
}

// Position tracks a live or simulated Solana token position.
type Position struct {
	Mint                  string         `json:"mint"`
	Name                  string         `json:"name,omitempty"`
	Symbol                string         `json:"symbol,omitempty"`
	Decimals              int            `json:"decimals"`
	QuantityRaw           uint64         `json:"quantity_raw"`
	QuantityUI            float64        `json:"quantity_ui"`
	EntryValueSOL         float64        `json:"entry_value_sol"`
	EntryValueUSD         float64        `json:"entry_value_usd,omitempty"`
	EntryPriceSOL         float64        `json:"entry_price_sol"`
	EntryPriceUSD         float64        `json:"entry_price_usd,omitempty"`
	CurrentPriceSOL       float64        `json:"current_price_sol,omitempty"`
	CurrentPriceUSD       float64        `json:"current_price_usd,omitempty"`
	HighWatermarkSOL      float64        `json:"high_watermark_sol,omitempty"`
	HighWatermarkUSD      float64        `json:"high_watermark_usd,omitempty"`
	EntryTxSignature      string         `json:"entry_tx_signature,omitempty"`
	ExitTxSignature       string         `json:"exit_tx_signature,omitempty"`
	Status                PositionStatus `json:"status"`
	OpenedAt              time.Time      `json:"opened_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	ClosedAt              time.Time      `json:"closed_at,omitempty"`
	RealizedPnLSOL        float64        `json:"realized_pnl_sol,omitempty"`
	RealizedPnLUSD        float64        `json:"realized_pnl_usd,omitempty"`
	UnrealizedPnLSOLValue float64        `json:"unrealized_pnl_sol,omitempty"`
	UnrealizedPnLUSDValue float64        `json:"unrealized_pnl_usd,omitempty"`
	LastObservedPriceSOL  float64        `json:"last_observed_price_sol,omitempty"`
	AllowDuplicateEntry   bool           `json:"allow_duplicate_entry,omitempty"`
}

// CurrentQuantityUI returns the current token quantity as a floating point value.
func (p Position) CurrentQuantityUI() float64 {
	if p.QuantityUI > 0 {
		return p.QuantityUI
	}
	if p.Decimals <= 0 {
		return float64(p.QuantityRaw)
	}
	return float64(p.QuantityRaw) / math.Pow10(p.Decimals)
}

// AverageEntryPriceSOL returns the cost basis per token in SOL.
func (p Position) AverageEntryPriceSOL() float64 {
	if q := p.CurrentQuantityUI(); q > 0 {
		return p.EntryValueSOL / q
	}
	return 0
}

// ExposureSOL estimates the current SOL exposure at a given token price.
func (p Position) ExposureSOL(currentPriceSOL float64) float64 {
	if currentPriceSOL <= 0 {
		currentPriceSOL = p.CurrentPriceSOL
	}
	return p.CurrentQuantityUI() * currentPriceSOL
}

// UnrealizedPnLSOL estimates profit or loss in SOL.
func (p Position) UnrealizedPnLSOL(currentPriceSOL float64) float64 {
	return p.ExposureSOL(currentPriceSOL) - p.EntryValueSOL
}

// UnrealizedPnLPct estimates profit or loss percentage relative to cost basis.
func (p Position) UnrealizedPnLPct(currentPriceSOL float64) float64 {
	if p.EntryValueSOL <= 0 {
		return 0
	}
	return (p.UnrealizedPnLSOL(currentPriceSOL) / p.EntryValueSOL) * 100
}

// UpdateHighWatermark updates the high-water mark used by trailing stops.
func (p *Position) UpdateHighWatermark(priceSOL float64) {
	if p == nil || priceSOL <= 0 {
		return
	}
	if priceSOL > p.HighWatermarkSOL {
		p.HighWatermarkSOL = priceSOL
	}
}
