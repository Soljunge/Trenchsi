package solana

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	solanagogo "github.com/gagliardetto/solana-go"
)

// Wallet wraps a Solana private key without ever exposing it in logs.
type Wallet struct {
	key solanagogo.PrivateKey
}

// LoadWalletFromEnv loads a wallet from SOLANA_PRIVATE_KEY or a keypair file path.
func LoadWalletFromEnv() (*Wallet, error) {
	raw := strings.TrimSpace(os.Getenv("SOLANA_PRIVATE_KEY"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("SOLANA_PRIVATE_KEY_PATH"))
	}
	if raw == "" {
		return nil, errors.New("SOLANA_PRIVATE_KEY is not set")
	}
	return LoadWalletFromValue(raw)
}

// LoadWalletFromValue loads a wallet from a base58 private key, a Solana keypair
// file path, or a JSON array of integers.
func LoadWalletFromValue(value string) (*Wallet, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("empty Solana private key")
	}

	if strings.HasPrefix(value, "file://") {
		value = strings.TrimPrefix(value, "file://")
	}

	if _, err := os.Stat(value); err == nil {
		key, err := solanagogo.PrivateKeyFromSolanaKeygenFile(value)
		if err != nil {
			return nil, fmt.Errorf("load solana keypair file: %w", err)
		}
		return &Wallet{key: key}, nil
	}

	if strings.HasPrefix(value, "[") {
		var keyBytes []byte
		if err := json.Unmarshal([]byte(value), &keyBytes); err == nil && len(keyBytes) > 0 {
			return &Wallet{key: solanagogo.PrivateKey(keyBytes)}, nil
		}

		var ints []int
		if err := json.Unmarshal([]byte(value), &ints); err == nil && len(ints) > 0 {
			keyBytes = make([]byte, len(ints))
			for i, n := range ints {
				keyBytes[i] = byte(n)
			}
			return &Wallet{key: solanagogo.PrivateKey(keyBytes)}, nil
		}
	}

	key, err := solanagogo.PrivateKeyFromBase58(value)
	if err != nil {
		return nil, fmt.Errorf("parse solana private key: %w", err)
	}
	return &Wallet{key: key}, nil
}

// PublicKey returns the wallet public key.
func (w *Wallet) PublicKey() solanagogo.PublicKey {
	if w == nil {
		return solanagogo.PublicKey{}
	}
	return w.key.PublicKey()
}

// SignTransaction signs a Solana transaction in place.
func (w *Wallet) SignTransaction(tx *solanagogo.Transaction) error {
	if w == nil || tx == nil {
		return errors.New("wallet or transaction is nil")
	}

	_, err := tx.Sign(func(pub solanagogo.PublicKey) *solanagogo.PrivateKey {
		if pub.Equals(w.PublicKey()) {
			return &w.key
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sign transaction: %w", err)
	}

	return nil
}

// String intentionally avoids revealing the private key.
func (w *Wallet) String() string {
	if w == nil {
		return "Wallet<nil>"
	}
	return fmt.Sprintf("Wallet{%s}", w.PublicKey().String())
}
