package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestJupiterQuoteAndSwap(t *testing.T) {
	var quoteHeader string
	var swapBody SwapRequest

	client := NewJupiterClient("https://example.invalid", "secret")
	client.httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/swap/v1/quote":
				quoteHeader = req.Header.Get("x-api-key")
				if req.Method != http.MethodGet {
					t.Fatalf("unexpected quote method: %s", req.Method)
				}
				values := req.URL.Query()
				if values.Get("inputMint") != WrappedSOLMint {
					t.Fatalf("unexpected input mint: %s", values.Get("inputMint"))
				}
				if values.Get("outputMint") != USDCMint {
					t.Fatalf("unexpected output mint: %s", values.Get("outputMint"))
				}
				body := QuoteResponse{
					InputMint:            WrappedSOLMint,
					OutputMint:           USDCMint,
					InAmount:             "100000000",
					OutAmount:            "17000000",
					OtherAmountThreshold: "16830000",
					SwapMode:             "ExactIn",
					SlippageBps:          100,
					PriceImpactPct:       "0.01",
					ContextSlot:          123,
				}
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(body); err != nil {
					t.Fatalf("encode quote: %v", err)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
					Header:     make(http.Header),
				}, nil
			case "/swap/v1/swap":
				if req.Method != http.MethodPost {
					t.Fatalf("unexpected swap method: %s", req.Method)
				}
				if err := json.NewDecoder(req.Body).Decode(&swapBody); err != nil {
					t.Fatalf("decode swap request: %v", err)
				}
				body := SwapResponse{
					SwapTransaction:      "AQ==",
					LastValidBlockHeight: 999,
				}
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(body); err != nil {
					t.Fatalf("encode swap: %v", err)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
					Header:     make(http.Header),
				}, nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			return nil, nil
		}),
	}

	quote, err := client.Quote(context.Background(), WrappedSOLMint, USDCMint, 100000000, 100)
	if err != nil {
		t.Fatalf("Quote() error: %v", err)
	}
	if quoteHeader != "secret" {
		t.Fatalf("unexpected api key header: %s", quoteHeader)
	}
	if quote.OutAmount != "17000000" {
		t.Fatalf("unexpected quote out amount: %s", quote.OutAmount)
	}

	resp, err := client.BuildSwapTransaction(context.Background(), SwapRequest{
		UserPublicKey: "wallet",
		QuoteResponse: QuoteResponse{
			InputMint:      WrappedSOLMint,
			OutputMint:     USDCMint,
			InAmount:       "100000000",
			OutAmount:      "17000000",
			SwapMode:       "ExactIn",
			SlippageBps:    100,
			PriceImpactPct: "0.01",
		},
		WrapAndUnwrapSOL:        true,
		DynamicComputeUnitLimit: true,
	})
	if err != nil {
		t.Fatalf("BuildSwapTransaction() error: %v", err)
	}
	if resp.SwapTransaction != "AQ==" {
		t.Fatalf("unexpected swap tx: %s", resp.SwapTransaction)
	}
	if swapBody.UserPublicKey != "wallet" {
		t.Fatalf("unexpected request body user: %s", swapBody.UserPublicKey)
	}
	if !swapBody.WrapAndUnwrapSOL || !swapBody.DynamicComputeUnitLimit {
		t.Fatal("expected swap flags to be forwarded")
	}
	if !strings.EqualFold(swapBody.QuoteResponse.InputMint, WrappedSOLMint) {
		t.Fatalf("unexpected swap quote input mint: %s", swapBody.QuoteResponse.InputMint)
	}
}
