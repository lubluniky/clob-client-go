// Example: Creating and (optionally) posting orders on Polymarket.
//
// This example demonstrates L1/L2 (authenticated) operations:
//   - Configuring the client with a signer key and API credentials
//   - Building and signing a limit order
//   - Building and signing a market order (FOK)
//   - Posting an order to the exchange (commented out for safety)
//   - Cancelling an order (commented out for safety)
//
// Required environment variables:
//
//	POLY_PRIVATE_KEY   - hex-encoded ECDSA private key (without 0x prefix)
//	POLY_API_KEY       - API key from Polymarket
//	POLY_API_SECRET    - API secret
//	POLY_API_PASSPHRASE - API passphrase
//	POLY_TOKEN_ID      - token ID of the outcome to trade
//
// Usage:
//
//	export POLY_PRIVATE_KEY=deadbeef...
//	export POLY_API_KEY=...
//	export POLY_API_SECRET=...
//	export POLY_API_PASSPHRASE=...
//	export POLY_TOKEN_ID=12345...
//	go run ./examples/orders
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
	// -----------------------------------------------------------------------
	// 1. Read configuration from environment
	// -----------------------------------------------------------------------
	privateKeyHex := os.Getenv("POLY_PRIVATE_KEY")
	apiKey := os.Getenv("POLY_API_KEY")
	apiSecret := os.Getenv("POLY_API_SECRET")
	apiPassphrase := os.Getenv("POLY_API_PASSPHRASE")
	tokenID := os.Getenv("POLY_TOKEN_ID")

	if privateKeyHex == "" || apiKey == "" || apiSecret == "" || apiPassphrase == "" || tokenID == "" {
		log.Fatal("Set POLY_PRIVATE_KEY, POLY_API_KEY, POLY_API_SECRET, POLY_API_PASSPHRASE, and POLY_TOKEN_ID")
	}

	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
	}

	// -----------------------------------------------------------------------
	// 2. Create authenticated client
	// -----------------------------------------------------------------------
	ctx := context.Background()
	c := polymarket.NewClobClient(
		polymarket.WithSigner(key),
		polymarket.WithCreds(polymarket.ApiCreds{
			ApiKey:        apiKey,
			ApiSecret:     apiSecret,
			ApiPassphrase: apiPassphrase,
		}),
	)

	fmt.Printf("Client address: %s\n\n", c.Address())

	// -----------------------------------------------------------------------
	// 3. Fetch token metadata (tick size, neg risk, fee rate)
	// -----------------------------------------------------------------------
	tickSize, err := c.GetTickSize(ctx, tokenID)
	if err != nil {
		log.Fatalf("GetTickSize: %v", err)
	}
	fmt.Printf("Token %s\n", tokenID)
	fmt.Printf("  Tick size: %s\n", tickSize)

	negRisk, err := c.GetNegRisk(ctx, tokenID)
	if err != nil {
		log.Fatalf("GetNegRisk: %v", err)
	}
	fmt.Printf("  Neg risk:  %v\n", negRisk)

	feeRate, err := c.GetFeeRateBps(ctx, tokenID)
	if err != nil {
		log.Fatalf("GetFeeRateBps: %v", err)
	}
	fmt.Printf("  Fee rate:  %s bps\n\n", feeRate)

	// -----------------------------------------------------------------------
	// 4. Build and sign a limit order (BUY 10 shares @ $0.50)
	// -----------------------------------------------------------------------
	limitArgs := polymarket.OrderArgs{
		TokenID: tokenID,
		Price:   decimal.NewFromFloat(0.50),
		Size:    decimal.NewFromFloat(10),
		Side:    polymarket.Buy,
	}

	signed, err := c.CreateOrder(ctx, limitArgs)
	if err != nil {
		log.Fatalf("CreateOrder: %v", err)
	}

	fmt.Println("=== Signed Limit Order ===")
	fmt.Printf("  Maker:       %s\n", signed.Maker)
	fmt.Printf("  Signer:      %s\n", signed.Signer)
	fmt.Printf("  Token ID:    %s\n", signed.TokenID)
	fmt.Printf("  Side:        %s\n", signed.Side)
	fmt.Printf("  MakerAmount: %s\n", signed.MakerAmount)
	fmt.Printf("  TakerAmount: %s\n", signed.TakerAmount)
	fmt.Printf("  FeeRateBps:  %s\n", signed.FeeRateBps)
	fmt.Printf("  Expiration:  %s\n", signed.Expiration)
	fmt.Printf("  Nonce:       %s\n", signed.Nonce)
	fmt.Printf("  Salt:        %s\n", signed.Salt)
	fmt.Printf("  Signature:   %s...\n\n", signed.Signature[:20])

	// -----------------------------------------------------------------------
	// 5. Build and sign a market order (FOK, BUY $25 worth @ worst price 0.60)
	// -----------------------------------------------------------------------
	marketArgs := polymarket.MarketOrderArgs{
		TokenID:   tokenID,
		Amount:    decimal.NewFromFloat(25),
		Side:      polymarket.Buy,
		Price:     decimal.NewFromFloat(0.60),
		OrderType: polymarket.FOK,
	}

	marketSigned, err := c.CreateMarketOrder(ctx, marketArgs)
	if err != nil {
		log.Fatalf("CreateMarketOrder: %v", err)
	}

	fmt.Println("=== Signed Market Order (FOK) ===")
	fmt.Printf("  Maker:       %s\n", marketSigned.Maker)
	fmt.Printf("  Side:        %s\n", marketSigned.Side)
	fmt.Printf("  MakerAmount: %s\n", marketSigned.MakerAmount)
	fmt.Printf("  TakerAmount: %s\n", marketSigned.TakerAmount)
	fmt.Printf("  Expiration:  %s\n\n", marketSigned.Expiration)

	// -----------------------------------------------------------------------
	// 6. (Optional) Post the limit order -- uncomment to actually submit
	// -----------------------------------------------------------------------
	// resp, err := c.PostOrder(ctx, *signed, polymarket.GTC, false)
	// if err != nil {
	// 	log.Fatalf("PostOrder: %v", err)
	// }
	// fmt.Printf("Order posted: ID=%s  Status=%s\n", resp.ID, resp.Status)

	// -----------------------------------------------------------------------
	// 7. (Optional) List open orders
	// -----------------------------------------------------------------------
	// fmt.Println("=== Open Orders ===")
	// for order, err := range c.GetOpenOrders(ctx, polymarket.OpenOrderParams{}) {
	// 	if err != nil {
	// 		log.Fatalf("GetOpenOrders: %v", err)
	// 	}
	// 	fmt.Printf("  %s  %s  %s @ %s  status=%s\n",
	// 		order.ID, order.Side, order.OriginalSize, order.Price, order.Status)
	// }

	// -----------------------------------------------------------------------
	// 8. (Optional) Cancel an order
	// -----------------------------------------------------------------------
	// err = c.CancelOrder(ctx, "order-id-here")
	// if err != nil {
	// 	log.Fatalf("CancelOrder: %v", err)
	// }
	// fmt.Println("Order cancelled successfully")

	// -----------------------------------------------------------------------
	// 9. (Optional) Cancel all open orders
	// -----------------------------------------------------------------------
	// err = c.CancelAll(ctx)
	// if err != nil {
	// 	log.Fatalf("CancelAll: %v", err)
	// }
	// fmt.Println("All orders cancelled")

	fmt.Println("Done. Uncomment the PostOrder/Cancel sections above to submit real orders.")
}
