package signing

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// EIP-712 domain constants for ClobAuth signing.
const (
	ClobDomainName    = "ClobAuthDomain"
	ClobDomainVersion = "1"
	ClobAuthMessage   = "This message attests that I control the given wallet"
)

// SignClobAuth creates an EIP-712 signature for CLOB authentication (L1).
//
// Parameters:
//   - key: ECDSA private key for signing
//   - chainID: blockchain chain ID (137 for Polygon, 80002 for Amoy)
//   - address: signer's Ethereum address (checksummed)
//   - timestamp: unix timestamp as string
//   - nonce: request nonce
//
// Returns the 0x-prefixed hex-encoded signature with V value adjusted to 27/28.
func SignClobAuth(key *ecdsa.PrivateKey, chainID int, address string, timestamp string, nonce int) (string, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": {
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    ClobDomainName,
			Version: ClobDomainVersion,
			ChainId: math.NewHexOrDecimal256(int64(chainID)),
		},
		Message: apitypes.TypedDataMessage{
			"address":   address,
			"timestamp": timestamp,
			"nonce":     fmt.Sprintf("%d", nonce),
			"message":   ClobAuthMessage,
		},
	}

	// Hash the domain separator.
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("signing: domain hash failed: %w", err)
	}

	// Hash the primary type message.
	messageHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", fmt.Errorf("signing: message hash failed: %w", err)
	}

	// EIP-712 final hash: keccak256("\x19\x01" || domainSeparator || messageHash)
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, messageHash...)
	hash := crypto.Keccak256Hash(rawData)

	// Sign with the private key.
	sig, err := crypto.Sign(hash.Bytes(), key)
	if err != nil {
		return "", fmt.Errorf("signing: ecdsa sign failed: %w", err)
	}

	// Adjust V value: go-ethereum returns V as 0/1, EIP-712 expects 27/28.
	sig[64] += 27

	return fmt.Sprintf("0x%x", sig), nil
}
