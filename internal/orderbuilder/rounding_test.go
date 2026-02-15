package orderbuilder

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateLimitOrderAmounts(t *testing.T) {
	tests := []struct {
		name        string
		side        string
		price       string
		size        string
		tickSize    string
		wantMaker   string
		wantTaker   string
	}{
		{
			name:      "Tick 0.1, BUY, price=0.5, size=100",
			side:      "BUY",
			price:     "0.5",
			size:      "100",
			tickSize:  "0.1",
			wantMaker: "50000000",
			wantTaker: "100000000",
		},
		{
			name:      "Tick 0.01, BUY, price=0.05, size=100",
			side:      "BUY",
			price:     "0.05",
			size:      "100",
			tickSize:  "0.01",
			wantMaker: "5000000",
			wantTaker: "100000000",
		},
		{
			name:      "Tick 0.001, BUY, price=0.005, size=100",
			side:      "BUY",
			price:     "0.005",
			size:      "100",
			tickSize:  "0.001",
			wantMaker: "500000",
			wantTaker: "100000000",
		},
		{
			name:      "Tick 0.0001, BUY, price=0.0005, size=100",
			side:      "BUY",
			price:     "0.0005",
			size:      "100",
			tickSize:  "0.0001",
			wantMaker: "50000",
			wantTaker: "100000000",
		},
		{
			name:      "Tick 0.01, BUY, price=0.34, size=100",
			side:      "BUY",
			price:     "0.34",
			size:      "100",
			tickSize:  "0.01",
			wantMaker: "34000000",
			wantTaker: "100000000",
		},
		{
			name:      "Tick 0.001, BUY, price=0.512, size=100",
			side:      "BUY",
			price:     "0.512",
			size:      "100",
			tickSize:  "0.001",
			wantMaker: "51200000",
			wantTaker: "100000000",
		},
		{
			name:      "Tick 0.1, SELL, price=0.5, size=100",
			side:      "SELL",
			price:     "0.5",
			size:      "100",
			tickSize:  "0.1",
			wantMaker: "100000000",
			wantTaker: "50000000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			price, _ := decimal.NewFromString(tc.price)
			size, _ := decimal.NewFromString(tc.size)

			maker, taker, err := CalculateLimitOrderAmounts(tc.side, price, size, tc.tickSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if maker != tc.wantMaker {
				t.Errorf("makerAmount: got %s, want %s", maker, tc.wantMaker)
			}
			if taker != tc.wantTaker {
				t.Errorf("takerAmount: got %s, want %s", taker, tc.wantTaker)
			}
		})
	}
}

func TestCalculateLimitOrderAmounts_TSCrossLanguageVectors(t *testing.T) {
	// Cross-language test vectors from Polymarket clob-client (TypeScript).
	tests := []struct {
		name      string
		side      string
		price     string
		size      string
		tickSize  string
		wantMaker string
		wantTaker string
	}{
		// 0.1 tick size BUY
		{"0.1 BUY 0.5 21.04", "BUY", "0.5", "21.04", "0.1", "10520000", "21040000"},
		{"0.1 BUY 0.7 170", "BUY", "0.7", "170", "0.1", "119000000", "170000000"},
		{"0.1 BUY 0.8 101", "BUY", "0.8", "101", "0.1", "80800000", "101000000"},
		// 0.01 tick size BUY
		{"0.01 BUY 0.56 21.04", "BUY", "0.56", "21.04", "0.01", "11782400", "21040000"},
		{"0.01 BUY 0.82 101", "BUY", "0.82", "101", "0.01", "82820000", "101000000"},
		{"0.01 BUY 0.78 12.8205", "BUY", "0.78", "12.8205", "0.01", "9999600", "12820000"},
		// 0.001 tick size BUY
		{"0.001 BUY 0.056 21.04", "BUY", "0.056", "21.04", "0.001", "1178240", "21040000"},
		{"0.001 BUY 0.082 101", "BUY", "0.082", "101", "0.001", "8282000", "101000000"},
		{"0.001 BUY 0.078 12.8205", "BUY", "0.078", "12.8205", "0.001", "999960", "12820000"},
		// 0.0001 tick size BUY
		{"0.0001 BUY 0.0056 21.04", "BUY", "0.0056", "21.04", "0.0001", "117824", "21040000"},
		{"0.0001 BUY 0.0082 101", "BUY", "0.0082", "101", "0.0001", "828200", "101000000"},
		{"0.0001 BUY 0.0078 12.8205", "BUY", "0.0078", "12.8205", "0.0001", "99996", "12820000"},
		// SELL (mirrored maker/taker)
		{"0.1 SELL 0.5 21.04", "SELL", "0.5", "21.04", "0.1", "21040000", "10520000"},
		{"0.1 SELL 0.7 170", "SELL", "0.7", "170", "0.1", "170000000", "119000000"},
		{"0.1 SELL 0.8 101", "SELL", "0.8", "101", "0.1", "101000000", "80800000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			price, _ := decimal.NewFromString(tc.price)
			size, _ := decimal.NewFromString(tc.size)

			maker, taker, err := CalculateLimitOrderAmounts(tc.side, price, size, tc.tickSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if maker != tc.wantMaker {
				t.Errorf("makerAmount: got %s, want %s", maker, tc.wantMaker)
			}
			if taker != tc.wantTaker {
				t.Errorf("takerAmount: got %s, want %s", taker, tc.wantTaker)
			}
		})
	}
}

func TestCalculateLimitOrderAmounts_InvalidSide(t *testing.T) {
	price, _ := decimal.NewFromString("0.5")
	size, _ := decimal.NewFromString("100")

	_, _, err := CalculateLimitOrderAmounts("HOLD", price, size, "0.1")
	if err == nil {
		t.Error("expected error for invalid side, got nil")
	}
}

func TestCalculateLimitOrderAmounts_InvalidTickSize(t *testing.T) {
	price, _ := decimal.NewFromString("0.5")
	size, _ := decimal.NewFromString("100")

	_, _, err := CalculateLimitOrderAmounts("BUY", price, size, "0.5")
	if err == nil {
		t.Error("expected error for invalid tick size, got nil")
	}
}

func TestRoundDown(t *testing.T) {
	// 123.456 truncated to 2 decimals = 123.45
	x, _ := decimal.NewFromString("123.456")
	got := RoundDown(x, 2)
	want, _ := decimal.NewFromString("123.45")
	if !got.Equal(want) {
		t.Errorf("RoundDown(123.456, 2): got %s, want %s", got, want)
	}
}

func TestRoundDown_Negative(t *testing.T) {
	// Truncation toward zero: -1.999 truncated to 2 = -1.99
	x, _ := decimal.NewFromString("-1.999")
	got := RoundDown(x, 2)
	want, _ := decimal.NewFromString("-1.99")
	if !got.Equal(want) {
		t.Errorf("RoundDown(-1.999, 2): got %s, want %s", got, want)
	}
}

func TestRoundNormal(t *testing.T) {
	// 1.235 rounded to 2 decimals = 1.24 (half-up)
	x, _ := decimal.NewFromString("1.235")
	got := RoundNormal(x, 2)
	want, _ := decimal.NewFromString("1.24")
	if !got.Equal(want) {
		t.Errorf("RoundNormal(1.235, 2): got %s, want %s", got, want)
	}
}

func TestRoundUp(t *testing.T) {
	// 1.231 rounded up to 2 decimals = 1.24
	x, _ := decimal.NewFromString("1.231")
	got := RoundUp(x, 2)
	want, _ := decimal.NewFromString("1.24")
	if !got.Equal(want) {
		t.Errorf("RoundUp(1.231, 2): got %s, want %s", got, want)
	}
}

func TestToTokenDecimals(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"100", "100000000"},
		{"34", "34000000"},
		{"0.5", "500000"},
		{"0.05", "50000"},
		{"0.005", "5000"},
		{"0.0005", "500"},
	}

	for _, tc := range tests {
		x, _ := decimal.NewFromString(tc.input)
		got := ToTokenDecimals(x)
		if got != tc.want {
			t.Errorf("ToTokenDecimals(%s): got %s, want %s", tc.input, got, tc.want)
		}
	}
}

func TestDecimalPlaces(t *testing.T) {
	tests := []struct {
		input string
		want  int32
	}{
		{"1.23", 2},
		{"5", 0},
		{"0.001", 3},
		{"100.0", 1},
		{"0.0001", 4},
	}

	for _, tc := range tests {
		x, _ := decimal.NewFromString(tc.input)
		got := DecimalPlaces(x)
		if got != tc.want {
			t.Errorf("DecimalPlaces(%s): got %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestGetRoundConfig(t *testing.T) {
	rc, err := GetRoundConfig("0.01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc.Price != 2 || rc.Size != 2 || rc.Amount != 4 {
		t.Errorf("unexpected config: %+v", rc)
	}

	_, err = GetRoundConfig("0.5")
	if err == nil {
		t.Error("expected error for unsupported tick size, got nil")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt := GenerateSalt()
	if salt == "" {
		t.Error("GenerateSalt returned empty string")
	}
	// Verify it's a valid integer string by parsing it.
	_, err := decimal.NewFromString(salt)
	if err != nil {
		t.Errorf("GenerateSalt returned non-numeric string: %s", salt)
	}
}

func TestValidatePrice(t *testing.T) {
	// Valid price
	price, _ := decimal.NewFromString("0.5")
	if err := ValidatePrice(price, "0.1"); err != nil {
		t.Errorf("expected no error for valid price, got: %v", err)
	}

	// Price too low
	low, _ := decimal.NewFromString("0.05")
	if err := ValidatePrice(low, "0.1"); err == nil {
		t.Error("expected error for price below tick size, got nil")
	}

	// Price too high
	high, _ := decimal.NewFromString("0.95")
	if err := ValidatePrice(high, "0.1"); err == nil {
		t.Error("expected error for price above 1-tick, got nil")
	}

	// Edge: price exactly at tick
	exact, _ := decimal.NewFromString("0.1")
	if err := ValidatePrice(exact, "0.1"); err != nil {
		t.Errorf("expected no error for price at tick boundary, got: %v", err)
	}

	// Edge: price exactly at 1-tick
	upper, _ := decimal.NewFromString("0.9")
	if err := ValidatePrice(upper, "0.1"); err != nil {
		t.Errorf("expected no error for price at upper boundary, got: %v", err)
	}

	// Invalid tick size string
	if err := ValidatePrice(price, "abc"); err == nil {
		t.Error("expected error for invalid tick size string, got nil")
	}
}
