package trading

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	solanabind "github.com/sipeed/trenchlaw/internal/solana"
	"github.com/sipeed/trenchlaw/pkg/logger"

	solanagogo "github.com/gagliardetto/solana-go"
)

// Engine wires scanners, strategy, risk and Solana execution together.
type Engine struct {
	wallet      *solanabind.Wallet
	rpcClient   *solanabind.RPCClient
	jupiter     *solanabind.JupiterClient
	scanner     TokenScanner
	strategy    Strategy
	risk        *RiskManager
	portfolio   *Portfolio
	store       *PortfolioStore
	limits      RiskLimits
	tradeSOL    float64
	quoteExpiry uint64
}

// EngineOptions describes the runtime dependencies for the engine.
type EngineOptions struct {
	Wallet        *solanabind.Wallet
	RPCClient     *solanabind.RPCClient
	Jupiter       *solanabind.JupiterClient
	Scanner       TokenScanner
	Strategy      Strategy
	Risk          *RiskManager
	Portfolio     *Portfolio
	Store         *PortfolioStore
	TradeSOL      float64
	QuoteAgeSlots int
}

// NewEngine creates a trading engine from its dependencies.
func NewEngine(opts EngineOptions) (*Engine, error) {
	if opts.Wallet == nil {
		return nil, errors.New("wallet is required")
	}
	if opts.RPCClient == nil {
		return nil, errors.New("rpc client is required")
	}
	if opts.Jupiter == nil {
		return nil, errors.New("jupiter client is required")
	}
	if opts.Scanner == nil {
		return nil, errors.New("scanner is required")
	}
	if opts.Strategy == nil {
		return nil, errors.New("strategy is required")
	}
	if opts.Risk == nil {
		return nil, errors.New("risk manager is required")
	}
	if opts.Portfolio == nil {
		opts.Portfolio = NewPortfolio(false)
	}

	limits := opts.Risk.Limits()
	tradeSOL := opts.TradeSOL
	if tradeSOL <= 0 {
		tradeSOL = limits.MaxPositionSOL
	}
	if tradeSOL <= 0 {
		tradeSOL = 0.1
	}

	quoteExpiry := uint64(opts.QuoteAgeSlots)
	if quoteExpiry == 0 {
		quoteExpiry = uint64(limits.MaxQuoteAgeSlots)
	}

	return &Engine{
		wallet:      opts.Wallet,
		rpcClient:   opts.RPCClient,
		jupiter:     opts.Jupiter,
		scanner:     opts.Scanner,
		strategy:    opts.Strategy,
		risk:        opts.Risk,
		portfolio:   opts.Portfolio,
		store:       opts.Store,
		limits:      limits,
		tradeSOL:    tradeSOL,
		quoteExpiry: quoteExpiry,
	}, nil
}

// LoadPortfolio restores persisted positions when a store is available.
func (e *Engine) LoadPortfolio(ctx context.Context) error {
	if e == nil || e.store == nil {
		return nil
	}
	snapshot, err := e.store.Load(ctx)
	if err != nil {
		return err
	}
	e.portfolio.Restore(snapshot)
	return nil
}

// RunCycle scans for candidates and monitors any open positions.
func (e *Engine) RunCycle(ctx context.Context) error {
	if e == nil {
		return errors.New("engine is nil")
	}

	candidates, err := e.scanner.DiscoverCandidates(ctx)
	if err != nil {
		logger.WarnCF("trading", "candidate scan failed", map[string]any{"error": err})
	} else {
		if err := e.processCandidates(ctx, candidates); err != nil {
			logger.WarnCF("trading", "candidate processing failed", map[string]any{"error": err})
		}
	}

	if err := e.monitorOpenPositions(ctx); err != nil {
		logger.WarnCF("trading", "position monitoring failed", map[string]any{"error": err})
	}

	return e.persistPortfolio(ctx)
}

func (e *Engine) processCandidates(ctx context.Context, candidates []TokenInfo) error {
	type scoredCandidate struct {
		token  TokenInfo
		signal StrategySignal
	}

	scored := make([]scoredCandidate, 0, len(candidates))
	for _, token := range candidates {
		signal, err := e.strategy.EvaluateToken(token)
		if err != nil {
			logger.DebugCF("trading", "candidate rejected", map[string]any{
				"mint":  token.Mint,
				"error": err,
			})
			continue
		}
		if signal.Action != SignalBuy {
			logger.DebugCF("trading", "candidate skipped", map[string]any{
				"mint":   token.Mint,
				"symbol": token.Symbol,
				"reason": signal.Reason,
				"score":  signal.Score,
			})
			continue
		}
		scored = append(scored, scoredCandidate{token: token, signal: signal})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].signal.Score > scored[j].signal.Score
	})

	for _, candidate := range scored {
		if e.limits.MaxOpenPositions > 0 && e.portfolio.OpenCount() >= e.limits.MaxOpenPositions {
			break
		}
		if err := e.tryOpenPosition(ctx, candidate.token, candidate.signal); err != nil {
			logger.WarnCF("trading", "buy attempt failed", map[string]any{
				"mint":   candidate.token.Mint,
				"symbol": candidate.token.Symbol,
				"error":  err,
			})
			continue
		}
	}
	return nil
}

func (e *Engine) tryOpenPosition(ctx context.Context, token TokenInfo, signal StrategySignal) error {
	if e == nil {
		return errors.New("engine is nil")
	}

	balanceLamports, err := e.rpcClient.GetSOLBalance(ctx, e.wallet.PublicKey())
	if err != nil {
		return err
	}
	balanceSOL := float64(balanceLamports) / float64(solanagogo.LAMPORTS_PER_SOL)

	quote := &QuoteView{
		InputMint:   solanabind.WrappedSOLMint,
		OutputMint:  token.Mint,
		SlippageBps: e.effectiveSlippageBps(),
	}

	if err := e.risk.CanOpenPosition(time.Now().UTC(), e.portfolio, balanceSOL, quote, token); err != nil {
		return err
	}

	if e.tradeSOL <= 0 {
		return errors.New("trade size is zero")
	}

	inputAmountRaw := uint64(math.Round(e.tradeSOL * float64(solanagogo.LAMPORTS_PER_SOL)))
	if inputAmountRaw == 0 {
		return errors.New("trade amount rounds to zero")
	}

	quoteResp, err := e.jupiter.Quote(ctx, solanabind.WrappedSOLMint, token.Mint, inputAmountRaw, e.effectiveSlippageBps())
	if err != nil {
		return err
	}

	quoteView, err := convertQuoteView(quoteResp)
	if err != nil {
		return err
	}
	currentSlot, err := e.currentSlot(ctx)
	if err == nil {
		if err := e.risk.ValidateQuoteAge(&quoteView, currentSlot); err != nil {
			return err
		}
	}

	if e.limits.MaxPriceImpactPct > 0 && quoteView.PriceImpactPct > e.limits.MaxPriceImpactPct {
		return fmt.Errorf("quote price impact %.4f above limit %.4f", quoteView.PriceImpactPct, e.limits.MaxPriceImpactPct)
	}

	request := solanabind.SwapRequest{
		UserPublicKey:            e.wallet.PublicKey().String(),
		QuoteResponse:            *quoteResp,
		WrapAndUnwrapSOL:         true,
		DynamicComputeUnitLimit:  true,
		SkipUserAccountsRPCCalls: false,
		DynamicSlippage:          false,
		AsLegacyTransaction:      false,
	}
	request.PrioritizationFeeLamports = &solanabind.PrioritizationFeeParams{
		PriorityLevelWithMaxLamports: &solanabind.PriorityLevelWithMaxLamports{
			PriorityLevel: "veryHigh",
			MaxLamports:   1_000_000,
			Global:        false,
		},
	}

	if e.limits.DryRun {
		logger.InfoCF("trading", "dry-run buy signal", map[string]any{
			"mint":     token.Mint,
			"symbol":   token.Symbol,
			"size_sol": e.tradeSOL,
			"score":    signal.Score,
			"reason":   signal.Reason,
			"dry_run":  true,
			"paper":    e.limits.EnablePaperTrading,
		})
		if e.limits.EnablePaperTrading {
			return e.openPaperPosition(ctx, token, inputAmountRaw, quoteView, signal)
		}
		return nil
	}

	beforeBalance, err := e.rpcClient.GetTokenBalance(ctx, e.wallet.PublicKey(), solanagogo.MustPublicKeyFromBase58(token.Mint))
	if err != nil {
		beforeBalance = 0
	}

	swapResp, err := e.jupiter.BuildSwapTransaction(ctx, request)
	if err != nil {
		return err
	}

	tx, err := solanabind.DecodeSwapTransaction(swapResp.SwapTransaction)
	if err != nil {
		return err
	}
	if err := e.wallet.SignTransaction(tx); err != nil {
		return err
	}
	if e.rpcClient == nil {
		return errors.New("rpc client is nil")
	}

	sig, err := e.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		return err
	}
	if err := e.rpcClient.ConfirmSignature(ctx, sig, 2*time.Minute); err != nil {
		return err
	}

	afterBalance, err := e.rpcClient.GetTokenBalance(ctx, e.wallet.PublicKey(), solanagogo.MustPublicKeyFromBase58(token.Mint))
	if err != nil {
		afterBalance = beforeBalance
	}

	actualOut := uint64(0)
	if afterBalance > beforeBalance {
		actualOut = afterBalance - beforeBalance
	}

	return e.finalizeOpenPosition(ctx, token, inputAmountRaw, actualOut, quoteView, sig, signal, false)
}

func (e *Engine) openPaperPosition(ctx context.Context, token TokenInfo, inputAmountRaw uint64, quoteView QuoteView, signal StrategySignal) error {
	actualOutRaw, err := uint64FromString(quoteView.OutAmount)
	if err != nil {
		return err
	}
	return e.finalizeOpenPosition(ctx, token, inputAmountRaw, actualOutRaw, quoteView, "paper-trade", signal, true)
}

func (e *Engine) finalizeOpenPosition(ctx context.Context, token TokenInfo, inputAmountRaw, outputAmountRaw uint64, quoteView QuoteView, signature string, signal StrategySignal, paper bool) error {
	if outputAmountRaw == 0 {
		outputAmountRaw = quoteView.OutAmountRaw()
	}
	quantityUI := rawToUI(outputAmountRaw, token.Decimals)
	if quantityUI <= 0 {
		return errors.New("quoted quantity is zero")
	}

	entryValueSOL := float64(inputAmountRaw) / float64(solanagogo.LAMPORTS_PER_SOL)
	entryPriceSOL := entryValueSOL / quantityUI
	entryValueUSD := quantityUI * token.USDPrice
	entryPriceUSD := token.USDPrice
	if entryPriceUSD <= 0 {
		entryPriceUSD = 0
	}

	pos := &Position{
		Mint:                token.Mint,
		Name:                token.Name,
		Symbol:              token.Symbol,
		Decimals:            token.Decimals,
		QuantityRaw:         outputAmountRaw,
		QuantityUI:          quantityUI,
		EntryValueSOL:       entryValueSOL,
		EntryValueUSD:       entryValueUSD,
		EntryPriceSOL:       entryPriceSOL,
		EntryPriceUSD:       entryPriceUSD,
		CurrentPriceSOL:     entryPriceSOL,
		CurrentPriceUSD:     token.USDPrice,
		HighWatermarkSOL:    entryPriceSOL,
		HighWatermarkUSD:    token.USDPrice,
		EntryTxSignature:    signature,
		Status:              PositionOpen,
		OpenedAt:            time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
		RealizedPnLSOL:      0,
		RealizedPnLUSD:      0,
		AllowDuplicateEntry: signal.Score > 0 && e.limits.AllowDuplicateEntries,
	}
	if pos.HighWatermarkUSD <= 0 {
		pos.HighWatermarkUSD = pos.EntryPriceUSD
	}
	if pos.HighWatermarkSOL <= 0 {
		pos.HighWatermarkSOL = pos.EntryPriceSOL
	}

	if err := e.portfolio.UpsertPosition(pos); err != nil {
		return err
	}
	e.portfolio.RecordTrade(time.Now().UTC())

	result := TradeResult{
		Success:        true,
		Side:           SignalBuy,
		TokenMint:      token.Mint,
		Signature:      signature,
		ExpectedOutRaw: quoteView.OutAmountRaw(),
		ActualOutRaw:   outputAmountRaw,
		InputAmountRaw: inputAmountRaw,
		InputAmountSOL: entryValueSOL,
		PriceImpactPct: quoteView.PriceImpactPct,
		ExecutedAt:     time.Now().UTC(),
		PaperTrading:   paper,
		DryRun:         e.limits.DryRun,
		EntryPriceSOL:  entryPriceSOL,
		ExitPriceSOL:   0,
	}
	logger.InfoCF("trading", "opened position", map[string]any{
		"mint":        token.Mint,
		"symbol":      token.Symbol,
		"signature":   signature,
		"size_sol":    entryValueSOL,
		"entry_price": entryPriceSOL,
		"score":       signal.Score,
		"paper":       paper,
		"dry_run":     e.limits.DryRun,
	})
	_ = result
	return e.persistPortfolio(ctx)
}

func (e *Engine) monitorOpenPositions(ctx context.Context) error {
	if e == nil || e.portfolio == nil {
		return nil
	}

	positions := e.portfolio.Positions()
	for i := range positions {
		pos := positions[i]
		if pos.Status != PositionOpen {
			continue
		}

		token, err := e.scanner.RefreshToken(ctx, pos.Mint)
		if err != nil {
			logger.DebugCF("trading", "refresh token failed", map[string]any{
				"mint":  pos.Mint,
				"error": err,
			})
			continue
		}

		if _, err := e.portfolio.MarkPosition(pos.Mint, 0, token.USDPrice); err != nil {
			logger.DebugCF("trading", "mark position failed", map[string]any{
				"mint":  pos.Mint,
				"error": err,
			})
		}

		latest, ok := e.portfolio.GetPosition(pos.Mint)
		if !ok {
			continue
		}

		exit, reason := e.risk.ShouldExitPosition(time.Now().UTC(), *latest, token)
		if !exit {
			continue
		}
		if !e.limits.EnableSells && !e.limits.DryRun && !e.limits.EnablePaperTrading {
			logger.InfoCF("trading", "exit signal ignored because sells are disabled", map[string]any{
				"mint":   pos.Mint,
				"reason": reason,
			})
			continue
		}
		if err := e.tryClosePosition(ctx, *latest, token, reason); err != nil {
			logger.WarnCF("trading", "sell attempt failed", map[string]any{
				"mint":   pos.Mint,
				"reason": reason,
				"error":  err,
			})
		}
	}
	return nil
}

func (e *Engine) tryClosePosition(ctx context.Context, pos Position, token TokenInfo, reason string) error {
	if e == nil {
		return errors.New("engine is nil")
	}

	if err := e.risk.CanSellPosition(time.Now().UTC(), e.portfolio); err != nil {
		return err
	}
	if e.limits.DryRun {
		logger.InfoCF("trading", "dry-run sell signal", map[string]any{
			"mint":   pos.Mint,
			"reason": reason,
		})
		if e.limits.EnablePaperTrading {
			exitValueSOL := pos.CurrentQuantityUI() * pos.CurrentPriceSOL
			if exitValueSOL <= 0 {
				exitValueSOL = pos.EntryValueSOL
			}
			exitValueUSD := pos.CurrentQuantityUI() * token.USDPrice
			if exitValueUSD <= 0 {
				exitValueUSD = pos.EntryValueUSD
			}
			_, _ = e.portfolio.ClosePosition(pos.Mint, exitValueSOL, exitValueUSD, "paper-sell", reason)
			return e.persistPortfolio(ctx)
		}
		return nil
	}

	swapAmount := pos.QuantityRaw
	if swapAmount == 0 {
		return errors.New("position quantity is zero")
	}

	beforeSOL, err := e.rpcClient.GetSOLBalance(ctx, e.wallet.PublicKey())
	if err != nil {
		beforeSOL = 0
	}

	quoteResp, err := e.jupiter.Quote(ctx, pos.Mint, solanabind.WrappedSOLMint, swapAmount, e.effectiveSlippageBps())
	if err != nil {
		return err
	}
	quoteView, err := convertQuoteView(quoteResp)
	if err != nil {
		return err
	}
	currentSlot, err := e.currentSlot(ctx)
	if err == nil {
		if err := e.risk.ValidateQuoteAge(&quoteView, currentSlot); err != nil {
			return err
		}
	}

	request := solanabind.SwapRequest{
		UserPublicKey:            e.wallet.PublicKey().String(),
		QuoteResponse:            *quoteResp,
		WrapAndUnwrapSOL:         true,
		DynamicComputeUnitLimit:  true,
		SkipUserAccountsRPCCalls: false,
		DynamicSlippage:          false,
		AsLegacyTransaction:      false,
	}
	request.PrioritizationFeeLamports = &solanabind.PrioritizationFeeParams{
		PriorityLevelWithMaxLamports: &solanabind.PriorityLevelWithMaxLamports{
			PriorityLevel: "veryHigh",
			MaxLamports:   1_000_000,
			Global:        false,
		},
	}

	swapResp, err := e.jupiter.BuildSwapTransaction(ctx, request)
	if err != nil {
		return err
	}
	tx, err := solanabind.DecodeSwapTransaction(swapResp.SwapTransaction)
	if err != nil {
		return err
	}
	if err := e.wallet.SignTransaction(tx); err != nil {
		return err
	}
	sig, err := e.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		return err
	}
	if err := e.rpcClient.ConfirmSignature(ctx, sig, 2*time.Minute); err != nil {
		return err
	}
	afterSOL, err := e.rpcClient.GetSOLBalance(ctx, e.wallet.PublicKey())
	if err != nil {
		afterSOL = beforeSOL
	}

	actualOutSOL := float64(0)
	if afterSOL > beforeSOL {
		actualOutSOL = float64(afterSOL-beforeSOL) / float64(solanagogo.LAMPORTS_PER_SOL)
	}

	exitValueUSD := pos.CurrentQuantityUI() * token.USDPrice
	_, _ = e.portfolio.ClosePosition(pos.Mint, actualOutSOL, exitValueUSD, sig, reason)
	e.portfolio.RecordTrade(time.Now().UTC())
	logger.InfoCF("trading", "closed position", map[string]any{
		"mint":    pos.Mint,
		"symbol":  pos.Symbol,
		"sig":     sig,
		"reason":  reason,
		"sol_out": actualOutSOL,
	})
	return e.persistPortfolio(ctx)
}

func (e *Engine) currentSlot(ctx context.Context) (uint64, error) {
	if e == nil || e.rpcClient == nil {
		return 0, errors.New("rpc client is nil")
	}
	_, slot, err := e.rpcClient.GetLatestBlockhash(ctx)
	return slot, err
}

func (e *Engine) persistPortfolio(ctx context.Context) error {
	if e == nil || e.store == nil {
		return nil
	}
	return e.store.Save(ctx, e.portfolio.Snapshot())
}

func (e *Engine) effectiveSlippageBps() int {
	if e == nil || e.limits.DefaultSlippageBps <= 0 {
		return 100
	}
	return e.limits.DefaultSlippageBps
}

func convertQuoteView(quote *solanabind.QuoteResponse) (QuoteView, error) {
	if quote == nil {
		return QuoteView{}, errors.New("quote is nil")
	}
	priceImpact := 0.0
	if quote.PriceImpactPct != "" {
		if parsed, err := strconv.ParseFloat(quote.PriceImpactPct, 64); err == nil {
			priceImpact = parsed * 100
		}
	}
	return QuoteView{
		InputMint:            quote.InputMint,
		OutputMint:           quote.OutputMint,
		InAmount:             quote.InAmount,
		OutAmount:            quote.OutAmount,
		OtherAmountThreshold: quote.OtherAmountThreshold,
		PriceImpactPct:       priceImpact,
		ContextSlot:          quote.ContextSlot,
		SlippageBps:          quote.SlippageBps,
	}, nil
}

func uint64FromString(v string) (uint64, error) {
	if strings.TrimSpace(v) == "" {
		return 0, errors.New("empty numeric string")
	}
	return strconv.ParseUint(v, 10, 64)
}

func rawToUI(raw uint64, decimals int) float64 {
	if raw == 0 {
		return 0
	}
	if decimals <= 0 {
		return float64(raw)
	}
	return float64(raw) / math.Pow10(decimals)
}
