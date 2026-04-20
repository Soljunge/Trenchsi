package trading

import (
	"context"
	"errors"
	"time"

	"github.com/sipeed/trenchlaw/pkg/logger"
)

// Runner polls for opportunities and keeps the portfolio up to date.
type Runner struct {
	engine          *Engine
	pollInterval    time.Duration
	monitorInterval time.Duration
	once            bool
}

// NewRunner creates a trading runner.
func NewRunner(engine *Engine, pollInterval, monitorInterval time.Duration, once bool) (*Runner, error) {
	if engine == nil {
		return nil, errors.New("engine is nil")
	}
	if pollInterval <= 0 {
		pollInterval = 30 * time.Second
	}
	if monitorInterval <= 0 {
		monitorInterval = 15 * time.Second
	}
	return &Runner{
		engine:          engine,
		pollInterval:    pollInterval,
		monitorInterval: monitorInterval,
		once:            once,
	}, nil
}

// Run starts the polling loop and stops on context cancellation.
func (r *Runner) Run(ctx context.Context) error {
	if r == nil || r.engine == nil {
		return errors.New("runner is not initialized")
	}

	if err := r.engine.LoadPortfolio(ctx); err != nil {
		logger.WarnCF("trading", "failed to load portfolio state", map[string]any{"error": err})
	}

	if r.once {
		return r.engine.RunCycle(ctx)
	}

	pollTicker := time.NewTicker(r.pollInterval)
	monitorTicker := time.NewTicker(r.monitorInterval)
	defer pollTicker.Stop()
	defer monitorTicker.Stop()

	// Run one cycle immediately so the operator does not wait for the first tick.
	if err := r.engine.RunCycle(ctx); err != nil {
		logger.WarnCF("trading", "initial trading cycle failed", map[string]any{"error": err})
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pollTicker.C:
			if err := r.engine.RunCycle(ctx); err != nil {
				logger.WarnCF("trading", "poll cycle failed", map[string]any{"error": err})
			}
		case <-monitorTicker.C:
			if err := r.engine.monitorOpenPositions(ctx); err != nil {
				logger.WarnCF("trading", "monitor cycle failed", map[string]any{"error": err})
			}
		}
	}
}
