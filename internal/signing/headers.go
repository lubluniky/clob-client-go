package signing

import (
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// Header key constants used by the Polymarket CLOB API.
const (
	HeaderAddress    = "POLY_ADDRESS"
	HeaderSignature  = "POLY_SIGNATURE"
	HeaderTimestamp  = "POLY_TIMESTAMP"
	HeaderNonce      = "POLY_NONCE"
	HeaderApiKey     = "POLY_API_KEY"
	HeaderPassphrase = "POLY_PASSPHRASE"
)

// BuildL0Headers returns empty headers for unauthenticated requests.
func BuildL0Headers() http.Header {
	return http.Header{}
}

// BuildL1Headers returns EIP-712 signed headers for L1 authentication.
// Used for API key creation and derivation.
func BuildL1Headers(key *ecdsa.PrivateKey, chainID int, nonce int) (http.Header, error) {
	address := crypto.PubkeyToAddress(key.PublicKey)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	sig, err := SignClobAuth(key, chainID, address.Hex(), timestamp, nonce)
	if err != nil {
		return nil, err
	}

	h := http.Header{}
	h.Set(HeaderAddress, address.Hex())
	h.Set(HeaderSignature, sig)
	h.Set(HeaderTimestamp, timestamp)
	h.Set(HeaderNonce, fmt.Sprintf("%d", nonce))
	return h, nil
}

// L2Credentials holds the API credentials needed for L2 (HMAC) signing.
type L2Credentials struct {
	ApiKey        string
	ApiSecret     string
	ApiPassphrase string
	Address       string // Ethereum address
}

// BuildL2Headers returns HMAC-signed headers for L2 authentication.
// Used for all authenticated API requests (orders, trades, etc.).
func BuildL2Headers(creds L2Credentials, method, path, body string) (http.Header, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	sig, err := BuildHMACSignature(creds.ApiSecret, timestamp, method, path, body)
	if err != nil {
		return nil, err
	}

	h := http.Header{}
	h.Set(HeaderAddress, creds.Address)
	h.Set(HeaderSignature, sig)
	h.Set(HeaderTimestamp, timestamp)
	h.Set(HeaderApiKey, creds.ApiKey)
	h.Set(HeaderPassphrase, creds.ApiPassphrase)
	return h, nil
}
