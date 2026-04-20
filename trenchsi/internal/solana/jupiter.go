package solana

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	bin "github.com/gagliardetto/binary"
	solanagogo "github.com/gagliardetto/solana-go"
)

const (
	DefaultJupiterBaseURL = "https://api.jup.ag"
	WrappedSOLMint        = "So11111111111111111111111111111111111111112"
	USDCMint              = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
)

// JupiterClient wraps Jupiter quote and swap API calls.
type JupiterClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewJupiterClient creates a client for Jupiter quote/swap calls.
func NewJupiterClient(baseURL, apiKey string) *JupiterClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultJupiterBaseURL
	}
	return &JupiterClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Quote requests a Jupiter quote for the provided mint pair and amount.
func (c *JupiterClient) Quote(ctx context.Context, inputMint, outputMint string, amount uint64, slippageBps int) (*QuoteResponse, error) {
	if c == nil {
		return nil, errors.New("jupiter client is nil")
	}
	if strings.TrimSpace(inputMint) == "" || strings.TrimSpace(outputMint) == "" {
		return nil, errors.New("input mint and output mint are required")
	}
	if amount == 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	quoteURL, err := url.Parse(c.baseURL + "/swap/v1/quote")
	if err != nil {
		return nil, fmt.Errorf("parse quote url: %w", err)
	}
	query := quoteURL.Query()
	query.Set("inputMint", inputMint)
	query.Set("outputMint", outputMint)
	query.Set("amount", fmt.Sprintf("%d", amount))
	query.Set("slippageBps", fmt.Sprintf("%d", slippageBps))
	query.Set("swapMode", "ExactIn")
	quoteURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quoteURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build quote request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("quote request failed: %s", resp.Status)
	}

	var out QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode quote response: %w", err)
	}
	return &out, nil
}

// BuildSwapTransaction requests a serialized unsigned swap transaction.
func (c *JupiterClient) BuildSwapTransaction(ctx context.Context, req SwapRequest) (*SwapResponse, error) {
	if c == nil {
		return nil, errors.New("jupiter client is nil")
	}
	if req.UserPublicKey == "" {
		return nil, errors.New("user public key is required")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal swap request: %w", err)
	}

	swapURL := c.baseURL + "/swap/v1/swap"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, swapURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build swap request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("x-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request swap transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("swap request failed: %s", resp.Status)
	}

	var out SwapResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode swap response: %w", err)
	}
	return &out, nil
}

// DecodeSwapTransaction converts Jupiter's base64 payload into a Solana transaction.
func DecodeSwapTransaction(encoded string) (*solanagogo.Transaction, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, errors.New("empty swap transaction payload")
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode swap transaction base64: %w", err)
	}

	return solanagogo.TransactionFromDecoder(bin.NewBinDecoder(raw))
}

// QuoteResponse represents the Jupiter quote response.
type QuoteResponse struct {
	InputMint            string       `json:"inputMint"`
	InAmount             string       `json:"inAmount"`
	OutputMint           string       `json:"outputMint"`
	OutAmount            string       `json:"outAmount"`
	OtherAmountThreshold string       `json:"otherAmountThreshold"`
	SwapMode             string       `json:"swapMode"`
	SlippageBps          int          `json:"slippageBps"`
	PriceImpactPct       string       `json:"priceImpactPct"`
	RoutePlan            []QuoteRoute `json:"routePlan"`
	PlatformFee          *PlatformFee `json:"platformFee,omitempty"`
	ContextSlot          uint64       `json:"contextSlot,omitempty"`
	TimeTaken            float64      `json:"timeTaken,omitempty"`
}

// QuoteRoute represents one hop in the Jupiter route plan.
type QuoteRoute struct {
	SwapInfo QuoteSwapInfo `json:"swapInfo"`
	Percent  int           `json:"percent"`
	Bps      int           `json:"bps,omitempty"`
}

// QuoteSwapInfo represents a single swap leg.
type QuoteSwapInfo struct {
	AMMKey     string `json:"ammKey"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	Label      string `json:"label"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

// PlatformFee represents the optional platform fee block in a quote response.
type PlatformFee struct {
	Amount string `json:"amount"`
	FeeBps int    `json:"feeBps"`
}

// SwapRequest is sent to Jupiter to create a serialized swap transaction.
type SwapRequest struct {
	UserPublicKey                 string                   `json:"userPublicKey"`
	QuoteResponse                 QuoteResponse            `json:"quoteResponse"`
	Payer                         string                   `json:"payer,omitempty"`
	WrapAndUnwrapSOL              bool                     `json:"wrapAndUnwrapSol,omitempty"`
	DynamicComputeUnitLimit       bool                     `json:"dynamicComputeUnitLimit,omitempty"`
	SkipUserAccountsRPCCalls      bool                     `json:"skipUserAccountsRpcCalls,omitempty"`
	DynamicSlippage               bool                     `json:"dynamicSlippage,omitempty"`
	AsLegacyTransaction           bool                     `json:"asLegacyTransaction,omitempty"`
	ComputeUnitPriceMicroLamports *uint64                  `json:"computeUnitPriceMicroLamports,omitempty"`
	PrioritizationFeeLamports     *PrioritizationFeeParams `json:"prioritizationFeeLamports,omitempty"`
	BlockhashSlotsToExpiry        *uint8                   `json:"blockhashSlotsToExpiry,omitempty"`
}

// PrioritizationFeeParams configures Jupiter priority fee selection.
type PrioritizationFeeParams struct {
	PriorityLevelWithMaxLamports *PriorityLevelWithMaxLamports `json:"priorityLevelWithMaxLamports,omitempty"`
}

// PriorityLevelWithMaxLamports limits the suggested priority fee.
type PriorityLevelWithMaxLamports struct {
	PriorityLevel string `json:"priorityLevel"`
	MaxLamports   uint64 `json:"maxLamports"`
	Global        bool   `json:"global,omitempty"`
}

// SwapResponse contains Jupiter's serialized transaction payload.
type SwapResponse struct {
	SwapTransaction           string                 `json:"swapTransaction"`
	LastValidBlockHeight      uint64                 `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports uint64                 `json:"prioritizationFeeLamports,omitempty"`
	DynamicSlippageReport     *DynamicSlippageReport `json:"dynamicSlippageReport,omitempty"`
}

// DynamicSlippageReport is included when Jupiter recomputes slippage.
type DynamicSlippageReport struct {
	SlippageBps                  int    `json:"slippageBps"`
	OtherAmount                  string `json:"otherAmount"`
	SimulatedIncurredSlippageBps int    `json:"simulatedIncurredSlippageBps"`
	AmplificationRatio           string `json:"amplificationRatio"`
}
