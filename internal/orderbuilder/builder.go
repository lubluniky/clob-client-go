package orderbuilder

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"
)

const (
	// USDCDecimals is the number of decimals used by USDC on Polygon.
	USDCDecimals = 6
	// MaxSafeInt is the max safe integer for JavaScript compatibility (salt masking).
	MaxSafeInt = (1 << 53) - 1
)

// Exchange contract addresses for EIP-712 order signing.
var (
	PolygonExchange        = common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E")
	PolygonNegRiskExchange = common.HexToAddress("0xC5d563A36AE78145C45a50134d48A1215220f80a")
	AmoyExchange           = common.HexToAddress("0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40")
	AmoyNegRiskExchange    = common.HexToAddress("0xC5d563A36AE78145C45a50134d48A1215220f80a")
)

// OrderData holds all the fields of a CTF Exchange order before signing.
type OrderData struct {
	Maker         common.Address
	Taker         common.Address
	TokenID       string
	MakerAmount   string // in base units (6 decimals)
	TakerAmount   string // in base units (6 decimals)
	Side          int    // 0=Buy, 1=Sell
	FeeRateBps    string
	Nonce         string
	Signer        common.Address
	Expiration    string
	SignatureType int
	Salt          string
}

// GenerateSalt returns a random integer string suitable for use as an order
// salt. The value is masked to fit within JavaScript's safe integer range.
func GenerateSalt() string {
	// Generate a random salt masked to 2^53-1 for JavaScript compatibility.
	// Uses rand.Int64N directly to avoid float64 precision loss that occurs
	// when multiplying large nanosecond timestamps by random floats.
	salt := rand.Int64N(MaxSafeInt)
	return fmt.Sprintf("%d", salt)
}

// CalculateLimitOrderAmounts computes maker and taker amounts for a limit order.
//
// For BUY: maker provides collateral (price * size), taker provides shares (size)
// For SELL: maker provides shares (size), taker provides collateral (price * size)
//
// Uses the Rust approach: truncate amounts at (tickDecimals + 2) scale, then convert to base units.
func CalculateLimitOrderAmounts(side string, price, size decimal.Decimal, tickSize string) (makerAmount, takerAmount string, err error) {
	rc, err := GetRoundConfig(tickSize)
	if err != nil {
		return "", "", err
	}

	rawPrice := RoundNormal(price, rc.Price)
	rawSize := RoundDown(size, rc.Size)

	// truncation scale = price decimals + size decimals (matching Rust: tick_decimals + LOT_SIZE_SCALE)
	truncScale := rc.Price + rc.Size

	switch side {
	case "BUY":
		// maker pays collateral, taker provides shares
		rawMakerAmt := rawSize.Mul(rawPrice).Truncate(truncScale)
		makerAmount = ToTokenDecimals(rawMakerAmt)
		takerAmount = ToTokenDecimals(rawSize)
	case "SELL":
		// maker provides shares, taker pays collateral
		rawTakerAmt := rawSize.Mul(rawPrice).Truncate(truncScale)
		makerAmount = ToTokenDecimals(rawSize)
		takerAmount = ToTokenDecimals(rawTakerAmt)
	default:
		return "", "", fmt.Errorf("orderbuilder: invalid side: %s", side)
	}

	return makerAmount, takerAmount, nil
}

// CalculateMarketOrderAmounts computes maker and taker amounts for a market order.
//
// For BUY: amount is USDC to spend, need to determine shares to receive based on price
// For SELL: amount is shares to sell, need to determine USDC to receive based on price
func CalculateMarketOrderAmounts(side string, amount, price decimal.Decimal, tickSize string) (makerAmount, takerAmount string, err error) {
	rc, err := GetRoundConfig(tickSize)
	if err != nil {
		return "", "", err
	}

	rawPrice := RoundNormal(price, rc.Price)
	truncScale := rc.Price + rc.Size

	switch side {
	case "BUY":
		// amount is USDC to spend, calculate shares to receive
		rawMakerAmt := RoundDown(amount, rc.Size)
		rawTakerAmt := rawMakerAmt.Div(rawPrice).Truncate(truncScale)
		makerAmount = ToTokenDecimals(rawMakerAmt)
		takerAmount = ToTokenDecimals(rawTakerAmt)
	case "SELL":
		// amount is shares to sell, calculate USDC to receive
		rawMakerAmt := RoundDown(amount, rc.Size)
		rawTakerAmt := rawMakerAmt.Mul(rawPrice).Truncate(truncScale)
		makerAmount = ToTokenDecimals(rawMakerAmt)
		takerAmount = ToTokenDecimals(rawTakerAmt)
	default:
		return "", "", fmt.Errorf("orderbuilder: invalid side: %s", side)
	}

	return makerAmount, takerAmount, nil
}

// SignOrder signs an OrderData using EIP-712 typed data signing for the CTF exchange.
func SignOrder(key *ecdsa.PrivateKey, chainID int, order OrderData, negRisk bool) (string, error) {
	// Select the correct exchange address
	var exchangeAddr common.Address
	switch {
	case chainID == 137 && !negRisk:
		exchangeAddr = PolygonExchange
	case chainID == 137 && negRisk:
		exchangeAddr = PolygonNegRiskExchange
	case chainID == 80002 && !negRisk:
		exchangeAddr = AmoyExchange
	case chainID == 80002 && negRisk:
		exchangeAddr = AmoyNegRiskExchange
	default:
		return "", fmt.Errorf("orderbuilder: unsupported chain ID: %d", chainID)
	}

	// Convert string fields to big.Int strings where the EIP-712 type is uint256.
	// The go-ethereum apitypes library handles string -> big.Int conversion for
	// uint256 fields.
	tokenID := new(big.Int)
	if _, ok := tokenID.SetString(order.TokenID, 10); !ok {
		return "", fmt.Errorf("orderbuilder: invalid tokenID: %s", order.TokenID)
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "expiration", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "feeRateBps", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
			},
		},
		PrimaryType: "Order",
		Domain: apitypes.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(int64(chainID)),
			VerifyingContract: exchangeAddr.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"salt":          order.Salt,
			"maker":         order.Maker.Hex(),
			"signer":        order.Signer.Hex(),
			"taker":         order.Taker.Hex(),
			"tokenId":       order.TokenID,
			"makerAmount":   order.MakerAmount,
			"takerAmount":   order.TakerAmount,
			"expiration":    order.Expiration,
			"nonce":         order.Nonce,
			"feeRateBps":    order.FeeRateBps,
			"side":          fmt.Sprintf("%d", order.Side),
			"signatureType": fmt.Sprintf("%d", order.SignatureType),
		},
	}

	// Hash domain separator
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("orderbuilder: domain hash failed: %w", err)
	}

	// Hash the Order message
	messageHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", fmt.Errorf("orderbuilder: message hash failed: %w", err)
	}

	// EIP-712 hash
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, messageHash...)
	hash := crypto.Keccak256Hash(rawData)

	sig, err := crypto.Sign(hash.Bytes(), key)
	if err != nil {
		return "", fmt.Errorf("orderbuilder: signing failed: %w", err)
	}

	// Adjust V: 0/1 -> 27/28
	sig[64] += 27

	return fmt.Sprintf("0x%x", sig), nil
}

// ValidatePrice checks that a price is valid for the given tick size.
// Price must be >= tickSize and <= 1 - tickSize.
func ValidatePrice(price decimal.Decimal, tickSize string) error {
	tick, err := decimal.NewFromString(tickSize)
	if err != nil {
		return fmt.Errorf("orderbuilder: invalid tick size: %s", tickSize)
	}

	one := decimal.NewFromInt(1)
	if price.LessThan(tick) || price.GreaterThan(one.Sub(tick)) {
		return fmt.Errorf("orderbuilder: price %s outside valid range [%s, %s]", price, tick, one.Sub(tick))
	}

	return nil
}
