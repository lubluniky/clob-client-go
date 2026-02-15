package signing

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// --- HMAC Tests ---

func TestBuildHMACSignature_RustVector(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	timestamp := "1000000"
	method := "test-sign"
	requestPath := "/orders"
	body := `{"hash":"0x123"}`

	expected := "4gJVbox-R6XlDK4nlaicig0_ANVL1qdcahiL8CXfXLM="

	sig, err := BuildHMACSignature(secret, timestamp, method, requestPath, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig != expected {
		t.Errorf("signature mismatch\n  got:  %s\n  want: %s", sig, expected)
	}
}

func TestBuildHMACSignature_PythonVector(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	timestamp := "1000000"
	method := "test-sign"
	requestPath := "/orders"
	body := `{"hash": "0x123"}` // note the space after the colon

	expected := "ZwAdJKvoYRlEKDkNMwd5BuwNNtg93kNaR_oU2HrfVvc="

	sig, err := BuildHMACSignature(secret, timestamp, method, requestPath, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig != expected {
		t.Errorf("signature mismatch\n  got:  %s\n  want: %s", sig, expected)
	}
}

func TestBuildHMACSignature_EmptyBody(t *testing.T) {
	secret := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	timestamp := "1000000"
	method := "GET"
	requestPath := "/markets"
	body := ""

	_, err := BuildHMACSignature(secret, timestamp, method, requestPath, body)
	if err != nil {
		t.Fatalf("expected no error for empty body GET request, got: %v", err)
	}
}

func TestBuildHMACSignature_InvalidBase64Secret(t *testing.T) {
	secret := "not-valid-base64!!!"
	timestamp := "1000000"
	method := "GET"
	requestPath := "/markets"
	body := ""

	_, err := BuildHMACSignature(secret, timestamp, method, requestPath, body)
	if err == nil {
		t.Fatal("expected error for invalid base64 secret, got nil")
	}
}

// --- EIP-712 Tests ---

func TestSignClobAuth_Vector(t *testing.T) {
	// Hardhat account 0 private key.
	privKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	key, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	expectedAddress := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	if address != expectedAddress {
		t.Errorf("address mismatch\n  got:  %s\n  want: %s", address, expectedAddress)
	}

	chainID := 137
	timestamp := "1000000"
	nonce := 0

	sig, err := SignClobAuth(key, chainID, address, timestamp, nonce)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify signature format.
	if sig == "" {
		t.Fatal("signature is empty")
	}
	if !strings.HasPrefix(sig, "0x") {
		t.Errorf("signature should start with 0x, got: %s", sig)
	}
	if len(sig) != 132 {
		t.Errorf("signature length should be 132 (0x + 65 bytes hex = 130 hex chars + 2), got %d: %s", len(sig), sig)
	}
}

func TestSignClobAuth_RustCrossLanguageVector(t *testing.T) {
	// Cross-language test vector from Polymarket rs-clob-client/src/auth.rs.
	// Uses Amoy (chain 80002), timestamp "10000000", nonce 23.
	privKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	key, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	chainID := 80002
	timestamp := "10000000"
	nonce := 23

	sig, err := SignClobAuth(key, chainID, address, timestamp, nonce)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "0xf62319a987514da40e57e2f4d7529f7bac38f0355bd88bb5adbb3768d80de6c1682518e0af677d5260366425f4361e7b70c25ae232aff0ab2331e2b164a1aedc1b"
	if sig != expected {
		t.Errorf("signature mismatch with Rust test vector\n  got:  %s\n  want: %s", sig, expected)
	}
}

func TestSignClobAuth_DifferentChainIDs(t *testing.T) {
	privKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	key, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	timestamp := "1000000"
	nonce := 0

	sig137, err := SignClobAuth(key, 137, address, timestamp, nonce)
	if err != nil {
		t.Fatalf("unexpected error signing with chainID 137: %v", err)
	}

	sig80002, err := SignClobAuth(key, 80002, address, timestamp, nonce)
	if err != nil {
		t.Fatalf("unexpected error signing with chainID 80002: %v", err)
	}

	if sig137 == sig80002 {
		t.Error("signatures for different chain IDs should differ, but they are the same")
	}
}

// --- Header Tests ---

func TestBuildL0Headers(t *testing.T) {
	h := BuildL0Headers()
	if len(h) != 0 {
		t.Errorf("expected empty headers, got %d entries", len(h))
	}
}

func TestBuildL1Headers(t *testing.T) {
	privKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	key, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	expectedAddress := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	nonce := 42

	h, err := BuildL1Headers(key, 137, nonce)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all 4 required headers are present.
	requiredHeaders := []string{HeaderAddress, HeaderSignature, HeaderTimestamp, HeaderNonce}
	for _, name := range requiredHeaders {
		if h.Get(name) == "" {
			t.Errorf("missing required header: %s", name)
		}
	}

	// Verify address value.
	if got := h.Get(HeaderAddress); got != expectedAddress {
		t.Errorf("address header mismatch\n  got:  %s\n  want: %s", got, expectedAddress)
	}

	// Verify nonce value.
	if got := h.Get(HeaderNonce); got != "42" {
		t.Errorf("nonce header mismatch\n  got:  %s\n  want: %s", got, "42")
	}
}

func TestBuildL2Headers(t *testing.T) {
	creds := L2Credentials{
		ApiKey:        "test-api-key",
		ApiSecret:     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		ApiPassphrase: "test-passphrase",
		Address:       "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
	}

	h, err := BuildL2Headers(creds, "GET", "/orders", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all 5 required headers are present.
	requiredHeaders := []string{HeaderAddress, HeaderSignature, HeaderTimestamp, HeaderApiKey, HeaderPassphrase}
	for _, name := range requiredHeaders {
		if h.Get(name) == "" {
			t.Errorf("missing required header: %s", name)
		}
	}

	// Verify specific header values.
	if got := h.Get(HeaderAddress); got != creds.Address {
		t.Errorf("address header mismatch\n  got:  %s\n  want: %s", got, creds.Address)
	}
	if got := h.Get(HeaderApiKey); got != creds.ApiKey {
		t.Errorf("api key header mismatch\n  got:  %s\n  want: %s", got, creds.ApiKey)
	}
	if got := h.Get(HeaderPassphrase); got != creds.ApiPassphrase {
		t.Errorf("passphrase header mismatch\n  got:  %s\n  want: %s", got, creds.ApiPassphrase)
	}
}
