package trading

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Portfolio tracks open positions and realized results.
type Portfolio struct {
	mu                    sync.RWMutex
	positions             map[string]*Position
	realizedPnLSOL        float64
	realizedPnLUSD        float64
	dailyRealizedPnLSOL   float64
	dailyRealizedPnLUSD   float64
	lastTradeAt           time.Time
	lastResetDay          string
	allowDuplicateEntries bool
	emergencyHalt         bool
}

// NewPortfolio creates an empty portfolio.
func NewPortfolio(allowDuplicateEntries bool) *Portfolio {
	return &Portfolio{
		positions:             make(map[string]*Position),
		allowDuplicateEntries: allowDuplicateEntries,
		lastResetDay:          utcDayKey(time.Now()),
	}
}

// Restore loads a portfolio snapshot into the live portfolio.
func (p *Portfolio) Restore(snapshot PortfolioSnapshot) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.positions = make(map[string]*Position, len(snapshot.Positions))
	for mint, pos := range snapshot.Positions {
		if pos == nil {
			continue
		}
		copyPos := *pos
		p.positions[mint] = &copyPos
	}
	p.realizedPnLSOL = snapshot.RealizedPnLSOL
	p.realizedPnLUSD = snapshot.RealizedPnLUSD
	p.dailyRealizedPnLSOL = snapshot.DailyRealizedPnLSOL
	p.dailyRealizedPnLUSD = snapshot.DailyRealizedPnLUSD
	p.lastTradeAt = snapshot.LastTradeAt
	p.lastResetDay = snapshot.LastResetDay
	p.emergencyHalt = snapshot.EmergencyHalt
}

// Snapshot returns a serializable copy of the portfolio state.
func (p *Portfolio) Snapshot() PortfolioSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	positions := make(map[string]*Position, len(p.positions))
	for mint, pos := range p.positions {
		if pos == nil {
			continue
		}
		copyPos := *pos
		positions[mint] = &copyPos
	}

	return PortfolioSnapshot{
		Positions:           positions,
		RealizedPnLSOL:      p.realizedPnLSOL,
		RealizedPnLUSD:      p.realizedPnLUSD,
		DailyRealizedPnLSOL: p.dailyRealizedPnLSOL,
		DailyRealizedPnLUSD: p.dailyRealizedPnLUSD,
		LastTradeAt:         p.lastTradeAt,
		LastResetDay:        p.lastResetDay,
		EmergencyHalt:       p.emergencyHalt,
	}
}

// OpenCount returns the number of open positions.
func (p *Portfolio) OpenCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, pos := range p.positions {
		if pos != nil && pos.Status == PositionOpen {
			count++
		}
	}
	return count
}

// HasOpenPosition returns true when the mint is already held.
func (p *Portfolio) HasOpenPosition(mint string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pos, ok := p.positions[mint]
	return ok && pos != nil && pos.Status == PositionOpen
}

// GetPosition returns a copy of an open or closed position.
func (p *Portfolio) GetPosition(mint string) (*Position, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pos, ok := p.positions[mint]
	if !ok || pos == nil {
		return nil, false
	}
	copyPos := *pos
	return &copyPos, true
}

// Positions returns the tracked positions as copies.
func (p *Portfolio) Positions() []Position {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]Position, 0, len(p.positions))
	for _, pos := range p.positions {
		if pos == nil {
			continue
		}
		out = append(out, *pos)
	}
	return out
}

// UpsertPosition stores or updates a position.
func (p *Portfolio) UpsertPosition(pos *Position) error {
	if pos == nil {
		return errors.New("position is nil")
	}
	if pos.Mint == "" {
		return errors.New("position mint is empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.allowDuplicateEntries {
		if existing, ok := p.positions[pos.Mint]; ok && existing != nil && existing.Status == PositionOpen && pos.Status == PositionOpen {
			return fmt.Errorf("duplicate open position for mint %s", pos.Mint)
		}
	}

	copyPos := *pos
	copyPos.UpdatedAt = time.Now().UTC()
	if copyPos.Status == "" {
		copyPos.Status = PositionOpen
	}
	p.positions[pos.Mint] = &copyPos
	return nil
}

// MarkPosition updates the current price and trailing high-water mark.
func (p *Portfolio) MarkPosition(mint string, priceSOL float64, priceUSD float64) (*Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pos, ok := p.positions[mint]
	if !ok || pos == nil {
		return nil, fmt.Errorf("position %s not found", mint)
	}
	pos.LastObservedPriceSOL = priceSOL
	if priceSOL > 0 {
		pos.CurrentPriceSOL = priceSOL
		pos.UpdateHighWatermark(priceSOL)
		pos.UnrealizedPnLSOLValue = pos.UnrealizedPnLSOL(priceSOL)
	}
	if priceUSD > 0 {
		pos.CurrentPriceUSD = priceUSD
		if priceUSD > pos.HighWatermarkUSD {
			pos.HighWatermarkUSD = priceUSD
		}
		pos.UnrealizedPnLUSDValue = pos.CurrentQuantityUI()*priceUSD - pos.EntryValueUSD
	}
	pos.UpdatedAt = time.Now().UTC()

	copyPos := *pos
	return &copyPos, nil
}

// ClosePosition closes an open position and records realized PnL.
func (p *Portfolio) ClosePosition(mint string, exitValueSOL float64, exitValueUSD float64, exitTx string, reason string) (*Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pos, ok := p.positions[mint]
	if !ok || pos == nil {
		return nil, fmt.Errorf("position %s not found", mint)
	}
	if pos.Status == PositionClosed {
		copyPos := *pos
		return &copyPos, nil
	}

	pos.ExitTxSignature = exitTx
	pos.Status = PositionClosed
	pos.ClosedAt = time.Now().UTC()
	pos.UpdatedAt = pos.ClosedAt

	realizedSOL := exitValueSOL - pos.EntryValueSOL
	realizedUSD := exitValueUSD - pos.EntryValueUSD
	pos.RealizedPnLSOL = realizedSOL
	pos.RealizedPnLUSD = realizedUSD
	if qty := pos.CurrentQuantityUI(); qty > 0 {
		pos.CurrentPriceSOL = exitValueSOL / qty
		pos.CurrentPriceUSD = exitValueUSD / qty
	}
	p.realizedPnLSOL += realizedSOL
	p.realizedPnLUSD += realizedUSD
	p.recordDailyPnLLocked(realizedSOL, realizedUSD)

	copyPos := *pos
	return &copyPos, nil
}

// RecordTrade updates the last trade timestamp and daily loss counters.
func (p *Portfolio) RecordTrade(now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lastTradeAt = now.UTC()
	p.ensureDayLocked(now.UTC())
}

// LastTradeAt returns the time of the most recent buy or sell.
func (p *Portfolio) LastTradeAt() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastTradeAt
}

// DailyRealizedPnLSOL returns the current UTC-day realized PnL.
func (p *Portfolio) DailyRealizedPnLSOL() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dailyRealizedPnLSOL
}

// DailyRealizedPnLUSD returns the current UTC-day realized PnL in USD.
func (p *Portfolio) DailyRealizedPnLUSD() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dailyRealizedPnLUSD
}

// EmergencyHalt returns the current emergency halt state.
func (p *Portfolio) EmergencyHalt() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.emergencyHalt
}

// SetEmergencyHalt updates the emergency halt state.
func (p *Portfolio) SetEmergencyHalt(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.emergencyHalt = enabled
}

// MarshalJSON exposes the portfolio as a snapshot.
func (p *Portfolio) MarshalJSON() ([]byte, error) {
	snapshot := p.Snapshot()
	return json.Marshal(snapshot)
}

// UnmarshalJSON restores a portfolio from a snapshot.
func (p *Portfolio) UnmarshalJSON(data []byte) error {
	var snapshot PortfolioSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	p.Restore(snapshot)
	return nil
}

func (p *Portfolio) ensureDayLocked(now time.Time) {
	dayKey := utcDayKey(now)
	if p.lastResetDay == dayKey {
		return
	}
	p.lastResetDay = dayKey
	p.dailyRealizedPnLSOL = 0
	p.dailyRealizedPnLUSD = 0
}

func (p *Portfolio) recordDailyPnLLocked(sol, usd float64) {
	p.ensureDayLocked(time.Now().UTC())
	p.dailyRealizedPnLSOL += sol
	p.dailyRealizedPnLUSD += usd
}

// PortfolioSnapshot is the persisted portfolio state.
type PortfolioSnapshot struct {
	Positions           map[string]*Position `json:"positions"`
	RealizedPnLSOL      float64              `json:"realized_pnl_sol"`
	RealizedPnLUSD      float64              `json:"realized_pnl_usd"`
	DailyRealizedPnLSOL float64              `json:"daily_realized_pnl_sol"`
	DailyRealizedPnLUSD float64              `json:"daily_realized_pnl_usd"`
	LastTradeAt         time.Time            `json:"last_trade_at,omitempty"`
	LastResetDay        string               `json:"last_reset_day,omitempty"`
	EmergencyHalt       bool                 `json:"emergency_halt,omitempty"`
}

func utcDayKey(now time.Time) string {
	return now.UTC().Format("2006-01-02")
}
