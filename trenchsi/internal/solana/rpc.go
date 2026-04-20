package solana

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	bin "github.com/gagliardetto/binary"
	solanagogo "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

const (
	defaultRPCTimeout    = 12 * time.Second
	defaultRetryAttempts = 3
)

// RPCClient wraps a Solana RPC client with timeouts and retry/backoff logic.
type RPCClient struct {
	client        *rpc.Client
	timeout       time.Duration
	retryAttempts int
}

// NewRPCClient creates a wrapper around a Solana RPC endpoint.
func NewRPCClient(endpoint string) *RPCClient {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = rpc.MainNetBeta_RPC
	}
	return &RPCClient{
		client:        rpc.New(endpoint),
		timeout:       defaultRPCTimeout,
		retryAttempts: defaultRetryAttempts,
	}
}

// SetTimeout adjusts the per-call timeout used by the wrapper.
func (c *RPCClient) SetTimeout(timeout time.Duration) {
	if c == nil || timeout <= 0 {
		return
	}
	c.timeout = timeout
}

// GetLatestBlockhash fetches a fresh blockhash and its last valid block height.
func (c *RPCClient) GetLatestBlockhash(ctx context.Context) (string, uint64, error) {
	var out *rpc.GetLatestBlockhashResult
	if err := c.do(ctx, func(callCtx context.Context) error {
		res, err := c.client.GetLatestBlockhash(callCtx, rpc.CommitmentConfirmed)
		if err != nil {
			return err
		}
		out = res
		return nil
	}); err != nil {
		return "", 0, err
	}
	if out == nil || out.Value == nil {
		return "", 0, errors.New("missing latest blockhash response")
	}
	return out.Value.Blockhash.String(), out.Value.LastValidBlockHeight, nil
}

// SendTransaction sends a signed transaction and returns the transaction signature.
func (c *RPCClient) SendTransaction(ctx context.Context, tx *solanagogo.Transaction) (string, error) {
	var signature solanagogo.Signature
	if err := c.do(ctx, func(callCtx context.Context) error {
		sig, err := c.client.SendTransactionWithOpts(callCtx, tx, rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentConfirmed,
		})
		if err != nil {
			return err
		}
		signature = sig
		return nil
	}); err != nil {
		return "", err
	}
	return signature.String(), nil
}

// ConfirmSignature waits until the transaction is confirmed or finalized.
func (c *RPCClient) ConfirmSignature(ctx context.Context, signature string, timeout time.Duration) error {
	sig, err := solanagogo.SignatureFromBase58(signature)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	confirmCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		confirmCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		var status *rpc.GetSignatureStatusesResult
		if err := c.do(confirmCtx, func(callCtx context.Context) error {
			res, err := c.client.GetSignatureStatuses(callCtx, true, sig)
			if err != nil {
				return err
			}
			status = res
			return nil
		}); err != nil {
			return err
		}

		if status != nil && len(status.Value) > 0 && status.Value[0] != nil {
			entry := status.Value[0]
			if entry.Err != nil {
				return fmt.Errorf("transaction failed: %v", entry.Err)
			}
			switch entry.ConfirmationStatus {
			case rpc.ConfirmationStatusConfirmed, rpc.ConfirmationStatusFinalized:
				return nil
			}
			if entry.Confirmations == nil && entry.Slot > 0 {
				return nil
			}
		}

		select {
		case <-confirmCtx.Done():
			return fmt.Errorf("confirm signature timed out: %w", confirmCtx.Err())
		case <-ticker.C:
		}
	}
}

// GetSOLBalance returns the current lamport balance for a wallet.
func (c *RPCClient) GetSOLBalance(ctx context.Context, owner solanagogo.PublicKey) (uint64, error) {
	var out *rpc.GetBalanceResult
	if err := c.do(ctx, func(callCtx context.Context) error {
		res, err := c.client.GetBalance(callCtx, owner, rpc.CommitmentConfirmed)
		if err != nil {
			return err
		}
		out = res
		return nil
	}); err != nil {
		return 0, err
	}
	if out == nil {
		return 0, errors.New("missing balance response")
	}
	return out.Value, nil
}

// GetTokenBalance returns the raw token balance for a specific mint.
func (c *RPCClient) GetTokenBalance(ctx context.Context, owner, mint solanagogo.PublicKey) (uint64, error) {
	var out *rpc.GetTokenAccountsResult
	if err := c.do(ctx, func(callCtx context.Context) error {
		res, err := c.client.GetTokenAccountsByOwner(
			callCtx,
			owner,
			&rpc.GetTokenAccountsConfig{Mint: mint.ToPointer()},
			&rpc.GetTokenAccountsOpts{Encoding: solanagogo.EncodingBase64},
		)
		if err != nil {
			return err
		}
		out = res
		return nil
	}); err != nil {
		return 0, err
	}

	var total uint64
	for _, account := range out.Value {
		if account == nil || account.Account.Data == nil {
			continue
		}
		var tok token.Account
		if err := bin.NewBinDecoder(account.Account.Data.GetBinary()).Decode(&tok); err != nil {
			return 0, fmt.Errorf("decode token account: %w", err)
		}
		total += tok.Amount
	}
	return total, nil
}

// SimulateTransaction runs a transaction simulation against the RPC node.
func (c *RPCClient) SimulateTransaction(ctx context.Context, tx *solanagogo.Transaction) (*rpc.SimulateTransactionResponse, error) {
	var out *rpc.SimulateTransactionResponse
	if err := c.do(ctx, func(callCtx context.Context) error {
		res, err := c.client.SimulateTransaction(callCtx, tx)
		if err != nil {
			return err
		}
		out = res
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RPCClient) do(ctx context.Context, fn func(context.Context) error) error {
	if c == nil || c.client == nil {
		return errors.New("rpc client is nil")
	}

	attempts := c.retryAttempts
	if attempts <= 0 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		err := fn(callCtx)
		cancel()
		if err == nil {
			return nil
		}

		lastErr = err
		if ctx.Err() != nil || attempt == attempts-1 || !isRetryableRPCError(err) {
			return err
		}

		backoff := time.Duration(math.Pow(2, float64(attempt))) * 250 * time.Millisecond
		backoff += time.Duration(rand.Intn(150)) * time.Millisecond

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}

func isRetryableRPCError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"),
		strings.Contains(msg, "temporarily unavailable"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "eof"),
		strings.Contains(msg, "429"),
		strings.Contains(msg, "502"),
		strings.Contains(msg, "503"),
		strings.Contains(msg, "504"):
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	return false
}
