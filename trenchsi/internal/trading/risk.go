package trading

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// RiskManager evaluates whether new trades or exits are allowed.
type RiskManager struct {
	limits RiskLimits
}

// NewRiskManager creates a new risk manager with normalized limits.
func NewRiskManager(limits RiskLimits) *RiskManager {
	return &RiskManager{limits: limits.Validate()}
}

// Limits returns the normalized limit set.
func (r *RiskManager) Limits() RiskLimits {
	if r == nil {
		return RiskLimits{}
	}
	return r.limits
}

// CanOpenPosition applies portfolio-level and wallet-level entry checks.
func (r *RiskManager) CanOpenPosition(now time.Time, portfolio *Portfolio, walletSOL float64, quote *QuoteView, token TokenInfo) error {
	if r == nil {
		return errors.New("risk manager is nil")
	}
	if portfolio == nil {
		return errors.New("portfolio is nil")
	}
	if r.limits.EmergencyHalt || portfolio.EmergencyHalt() {
		return errors.New("trading is halted")
	}
	if r.limits.MaxOpenPositions > 0 && portfolio.OpenCount() >= r.limits.MaxOpenPositions {
		return fmt.Errorf("max open positions reached: %d", r.limits.MaxOpenPositions)
	}
	if !r.limits.AllowDuplicateEntries && portfolio.HasOpenPosition(token.Mint) {
		return fmt.Errorf("duplicate position for %s is disabled", token.Mint)
	}
	if r.limits.MaxDailyLossSOL > 0 && portfolio.DailyRealizedPnLSOL() <= -r.limits.MaxDailyLossSOL {
		return fmt.Errorf("daily loss limit reached: %.4f SOL", portfolio.DailyRealizedPnLSOL())
	}
	if r.limits.TradeCooldownSeconds > 0 {
		last := portfolio.LastTradeAt()
		if !last.IsZero() && now.Sub(last) < time.Duration(r.limits.TradeCooldownSeconds)*time.Second {
			return fmt.Errorf("cooldown active for %s", time.Duration(r.limits.TradeCooldownSeconds)*time.Second)
		}
	}
	if !r.limits.DryRun && !r.limits.EnablePaperTrading {
		if walletSOL <= r.limits.ReserveSOL {
			return fmt.Errorf("wallet balance %.4f SOL is below reserve %.4f SOL", walletSOL, r.limits.ReserveSOL)
		}
		if r.limits.MaxPositionSOL > 0 {
			available := math.Max(0, walletSOL-r.limits.ReserveSOL)
			if available < r.limits.MaxPositionSOL {
				return fmt.Errorf("insufficient wallet balance for max position: available %.4f SOL, need %.4f SOL", available, r.limits.MaxPositionSOL)
			}
		}
	}
	if quote != nil && r.limits.MaxPriceImpactPct > 0 && quote.PriceImpactPct > r.limits.MaxPriceImpactPct {
		return fmt.Errorf("price impact %.4f exceeds limit %.4f", quote.PriceImpactPct, r.limits.MaxPriceImpactPct)
	}
	return nil
}

// CanSellPosition checks whether an exit is permitted.
func (r *RiskManager) CanSellPosition(now time.Time, portfolio *Portfolio) error {
	if r == nil {
		return errors.New("risk manager is nil")
	}
	if portfolio == nil {
		return errors.New("portfolio is nil")
	}
	if r.limits.EmergencyHalt || portfolio.EmergencyHalt() {
		return errors.New("trading is halted")
	}
	if !r.limits.EnableSells && !r.limits.DryRun && !r.limits.EnablePaperTrading {
		return errors.New("sell execution is disabled")
	}
	return nil
}

// ShouldExitPosition determines whether a position should be reduced or closed.
func (r *RiskManager) ShouldExitPosition(now time.Time, pos Position, token TokenInfo) (bool, string) {
	if r == nil {
		return false, ""
	}
	currentPrice := token.USDPrice
	if currentPrice <= 0 {
		currentPrice = pos.CurrentPriceUSD
	}

	if r.limits.StopLossPct > 0 && pos.EntryPriceUSD > 0 && currentPrice > 0 {
		if currentPrice <= pos.EntryPriceUSD*(1-r.limits.StopLossPct/100) {
			return true, "stop loss hit"
		}
	}
	if r.limits.TakeProfitPct > 0 && pos.EntryPriceUSD > 0 && currentPrice > 0 {
		if currentPrice >= pos.EntryPriceUSD*(1+r.limits.TakeProfitPct/100) {
			return true, "take profit hit"
		}
	}
	if r.limits.TrailingStopPct > 0 && pos.HighWatermarkUSD > 0 && currentPrice > 0 {
		if currentPrice <= pos.HighWatermarkUSD*(1-r.limits.TrailingStopPct/100) {
			return true, "trailing stop hit"
		}
	}
	if r.limits.MaxHoldMinutes > 0 && !pos.OpenedAt.IsZero() {
		if now.Sub(pos.OpenedAt) >= time.Duration(r.limits.MaxHoldMinutes)*time.Minute {
			return true, "max hold duration exceeded"
		}
	}
	if r.limits.EmergencyHalt {
		return true, "emergency halt"
	}
	return false, ""
}

// ValidateQuoteAge verifies that a quote is not stale relative to the current slot.
func (r *RiskManager) ValidateQuoteAge(quote *QuoteView, currentSlot uint64) error {
	if quote == nil {
		return errors.New("quote is nil")
	}
	if r.limits.MaxQuoteAgeSlots <= 0 {
		return nil
	}
	if quote.ContextSlot == 0 {
		return errors.New("quote is missing context slot")
	}
	if currentSlot < quote.ContextSlot {
		return nil
	}
	if diff := currentSlot - quote.ContextSlot; diff > uint64(r.limits.MaxQuoteAgeSlots) {
		return fmt.Errorf("quote is stale by %d slots", diff)
	}
	return nil
}
