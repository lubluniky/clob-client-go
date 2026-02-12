package orderbuilder

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// RoundConfig defines the number of decimal places used for price, size, and
// amount rounding at a given tick size.
type RoundConfig struct {
	Price  int32 // decimal places for price
	Size   int32 // decimal places for size
	Amount int32 // decimal places for amount
}

// RoundingConfigs maps each supported tick-size string to its rounding
// configuration. The mapping follows the Polymarket CLOB specification:
//
//	| Tick    | Price decimals | Size decimals | Amount decimals |
//	|---------|---------------|---------------|-----------------|
//	| 0.1     | 1             | 2             | 3               |
//	| 0.01    | 2             | 2             | 4               |
//	| 0.001   | 3             | 2             | 5               |
//	| 0.0001  | 4             | 2             | 6               |
var RoundingConfigs = map[string]RoundConfig{
	"0.1":    {Price: 1, Size: 2, Amount: 3},
	"0.01":   {Price: 2, Size: 2, Amount: 4},
	"0.001":  {Price: 3, Size: 2, Amount: 5},
	"0.0001": {Price: 4, Size: 2, Amount: 6},
}

// RoundDown truncates toward zero at the given number of decimal places.
func RoundDown(d decimal.Decimal, places int32) decimal.Decimal {
	return d.Truncate(places)
}

// RoundNormal rounds to nearest at the given number of decimal places (standard rounding).
func RoundNormal(d decimal.Decimal, places int32) decimal.Decimal {
	return d.Round(places)
}

// RoundUp rounds away from zero (ceiling for positive numbers).
func RoundUp(d decimal.Decimal, places int32) decimal.Decimal {
	factor := decimal.New(1, places) // 10^places
	return d.Mul(factor).Ceil().Div(factor)
}

// DecimalPlaces returns the number of decimal digits in d.
func DecimalPlaces(d decimal.Decimal) int32 {
	// Use the exponent from the decimal representation
	exp := d.Exponent()
	if exp >= 0 {
		return 0
	}
	return -exp
}

// ToTokenDecimals converts a decimal to USDC base units (6 decimals) as a string.
// Result is truncated to integer (no fractional wei).
func ToTokenDecimals(d decimal.Decimal) string {
	return d.Mul(decimal.New(1, 6)).Truncate(0).String()
}

// GetRoundConfig returns the RoundConfig for the given tick size string.
// Returns an error if the tick size is not recognized.
func GetRoundConfig(tickSize string) (RoundConfig, error) {
	rc, ok := RoundingConfigs[tickSize]
	if !ok {
		return RoundConfig{}, fmt.Errorf("orderbuilder: unsupported tick size: %s", tickSize)
	}
	return rc, nil
}
