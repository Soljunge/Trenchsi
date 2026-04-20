package run

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sipeed/trenchlaw/cmd/trenchlaw/internal"
	solanabind "github.com/sipeed/trenchlaw/internal/solana"
	"github.com/sipeed/trenchlaw/internal/trading"
	"github.com/sipeed/trenchlaw/pkg/config"
)

func NewRunCommand() *cobra.Command {
	var (
		strategyName   string
		once           bool
		dryRunFlag     bool
		paperTrading   bool
		tradeSOL       float64
		pollSeconds    int
		monitorSeconds int
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run automated trading strategies",
		Long: `Run an automated trading strategy against Solana token candidates.

The initial implementation supports the conservative "solana-trench" strategy.
Use dry-run or paper trading first to validate filters, sizing, and exit logic.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(internal.GetConfigPath())
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			strategy := strategyName
			if strategy == "" {
				strategy = cfg.Trading.Strategy
			}
			if strategy == "" {
				strategy = "solana-trench"
			}
			if strategy != "solana-trench" {
				return fmt.Errorf("unsupported strategy %q", strategy)
			}

			tradeSizeSOL := cfg.Trading.MaxPositionSOL
			if cmd.Flags().Changed("trade-sol") && tradeSOL > 0 {
				tradeSizeSOL = tradeSOL
			}
			if tradeSizeSOL <= 0 {
				tradeSizeSOL = 0.1
			}

			paperTradingEnabled := cfg.Trading.EnablePaperTrading
			if cmd.Flags().Changed("paper-trading") {
				paperTradingEnabled = paperTrading
			}
			dryRunEnabled := cfg.Trading.DryRun
			if cmd.Flags().Changed("dry-run") {
				dryRunEnabled = dryRunFlag
			}

			wallet, err := solanabind.LoadWalletFromEnv()
			if err != nil {
				return err
			}

			rpcURL, wsURL := resolveSolanaEndpoints(cfg.Trading.Network, cfg.Trading.RPCURL, cfg.Trading.WSURL)
			rpcClient := solanabind.NewRPCClient(rpcURL)
			_ = wsURL

			jupiterAPIKey := os.Getenv("JUPITER_API_KEY")
			jupiterClient := solanabind.NewJupiterClient("https://api.jup.ag", jupiterAPIKey)
			scanner := trading.NewJupiterScanner("https://api.jup.ag", jupiterAPIKey, cfg.Trading.ScanCategory, cfg.Trading.ScanInterval, cfg.Trading.CandidateLimit)

			limits := trading.RiskLimits{
				MaxPositionSOL:            cfg.Trading.MaxPositionSOL,
				MaxDailyLossSOL:           cfg.Trading.MaxDailyLossSOL,
				MaxOpenPositions:          cfg.Trading.MaxOpenPositions,
				DefaultSlippageBps:        cfg.Trading.DefaultSlippageBps,
				TakeProfitPct:             cfg.Trading.TakeProfitPct,
				StopLossPct:               cfg.Trading.StopLossPct,
				TrailingStopPct:           cfg.Trading.TrailingStopPct,
				TradeCooldownSeconds:      cfg.Trading.TradeCooldownSeconds,
				MaxHoldMinutes:            cfg.Trading.MaxHoldMinutes,
				ReserveSOL:                cfg.Trading.ReserveSOL,
				EnablePaperTrading:        paperTradingEnabled,
				EnableSells:               cfg.Trading.EnableSells,
				DryRun:                    dryRunEnabled,
				MaxQuoteAgeSlots:          cfg.Trading.MaxQuoteAgeSlots,
				MaxPriceImpactPct:         cfg.Trading.MaxPriceImpactPct,
				MaxVolatilityPct:          cfg.Trading.MaxVolatilityPct,
				MinLiquidityUSD:           cfg.Trading.MinLiquidityUSD,
				MinVolumeUSD:              cfg.Trading.MinVolumeUSD,
				MinMarketCapUSD:           cfg.Trading.MinMarketCapUSD,
				MaxMarketCapUSD:           cfg.Trading.MaxMarketCapUSD,
				MinOrganicScore:           cfg.Trading.MinOrganicScore,
				MaxHolderConcentrationPct: cfg.Trading.MaxHolderConcentrationPct,
				AllowDuplicateEntries:     cfg.Trading.AllowDuplicateEntries,
				EmergencyHalt:             cfg.Trading.EmergencyHalt,
			}
			risk := trading.NewRiskManager(limits)
			strategyImpl := trading.NewBasicStrategy(trading.BasicStrategyConfig{
				MinLiquidityUSD:           cfg.Trading.MinLiquidityUSD,
				MinVolumeUSD:              cfg.Trading.MinVolumeUSD,
				MaxVolatilityPct:          cfg.Trading.MaxVolatilityPct,
				MinMarketCapUSD:           cfg.Trading.MinMarketCapUSD,
				MaxMarketCapUSD:           cfg.Trading.MaxMarketCapUSD,
				MinOrganicScore:           cfg.Trading.MinOrganicScore,
				MaxHolderConcentrationPct: cfg.Trading.MaxHolderConcentrationPct,
			})

			store := trading.NewPortfolioStore(cfg.WorkspacePath())
			portfolio := trading.NewPortfolio(cfg.Trading.AllowDuplicateEntries)

			engine, err := trading.NewEngine(trading.EngineOptions{
				Wallet:        wallet,
				RPCClient:     rpcClient,
				Jupiter:       jupiterClient,
				Scanner:       scanner,
				Strategy:      strategyImpl,
				Risk:          risk,
				Portfolio:     portfolio,
				Store:         store,
				TradeSOL:      tradeSizeSOL,
				QuoteAgeSlots: cfg.Trading.MaxQuoteAgeSlots,
			})
			if err != nil {
				return err
			}

			poll := time.Duration(cfg.Trading.PollIntervalSeconds) * time.Second
			if pollSeconds > 0 {
				poll = time.Duration(pollSeconds) * time.Second
			}
			monitor := time.Duration(cfg.Trading.MonitorIntervalSeconds) * time.Second
			if monitorSeconds > 0 {
				monitor = time.Duration(monitorSeconds) * time.Second
			}
			runner, err := trading.NewRunner(engine, poll, monitor, once)
			if err != nil {
				return err
			}

			if strings.EqualFold(cfg.Trading.Network, "devnet") && !limits.DryRun && !limits.EnablePaperTrading {
				fmt.Fprintln(cmd.OutOrStdout(), "Warning: live Jupiter swaps on devnet are not generally supported; dry-run is recommended.")
			}

			return runner.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&strategyName, "strategy", "", "Strategy to run (solana-trench)")
	cmd.Flags().BoolVar(&once, "once", false, "Run a single cycle and exit")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Do not submit transactions to Solana")
	cmd.Flags().BoolVar(&paperTrading, "paper-trading", true, "Track virtual positions instead of live balances")
	cmd.Flags().Float64Var(&tradeSOL, "trade-sol", 0, "Override the SOL position size for buys")
	cmd.Flags().IntVar(&pollSeconds, "poll-seconds", 0, "Override the poll interval in seconds")
	cmd.Flags().IntVar(&monitorSeconds, "monitor-seconds", 0, "Override the monitor interval in seconds")

	return cmd
}

func resolveSolanaEndpoints(network, rpcURL, wsURL string) (string, string) {
	if strings.TrimSpace(rpcURL) == "" {
		switch strings.ToLower(strings.TrimSpace(network)) {
		case "devnet":
			rpcURL = "https://api.devnet.solana.com"
		default:
			rpcURL = "https://api.mainnet-beta.solana.com"
		}
	}
	if strings.TrimSpace(wsURL) == "" {
		switch strings.ToLower(strings.TrimSpace(network)) {
		case "devnet":
			wsURL = "wss://api.devnet.solana.com"
		default:
			wsURL = "wss://api.mainnet-beta.solana.com"
		}
	}
	return rpcURL, wsURL
}
