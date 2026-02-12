// Example: Fetching public market data from Polymarket CLOB API.
//
// This example demonstrates L0 (unauthenticated) operations:
//   - Querying server time
//   - Iterating over markets with auto-pagination
//   - Fetching an order book snapshot
//   - Fetching midpoint, spread, and last trade price
//
// No private key or API credentials are needed.
//
// Usage:
//
//	go run ./examples/markets
package main

import (
	"context"
	"fmt"
	"log"

	polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
	ctx := context.Background()
	c := polymarket.NewClobClient()

	// -----------------------------------------------------------------------
	// 1. Server time
	// -----------------------------------------------------------------------
	t, err := c.ServerTime(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server time: %d\n\n", t)

	// -----------------------------------------------------------------------
	// 2. Iterate over the first 5 markets
	// -----------------------------------------------------------------------
	fmt.Println("=== First 5 Markets ===")

	var firstTokenID string
	count := 0

	for market, err := range c.GetMarkets(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Market: %s\n", market.Question)
		fmt.Printf("  Condition ID:  %s\n", market.ConditionID)
		fmt.Printf("  Active:        %v\n", market.Active)
		fmt.Printf("  Accepting:     %v\n", market.AcceptingOrders)
		fmt.Printf("  Neg Risk:      %v\n", market.NegRisk)
		fmt.Printf("  Tokens:        %d\n", len(market.Tokens))
		for i, tok := range market.Tokens {
			fmt.Printf("    [%d] %s  price=%s  id=%s\n", i, tok.Outcome, tok.Price.String(), tok.TokenID)
		}
		fmt.Println()

		// Save the first token ID we find for the order book demo below
		if firstTokenID == "" && len(market.Tokens) > 0 && market.Active {
			firstTokenID = market.Tokens[0].TokenID
		}

		count++
		if count >= 5 {
			break
		}
	}

	// -----------------------------------------------------------------------
	// 3. Fetch order book for the first active token
	// -----------------------------------------------------------------------
	if firstTokenID == "" {
		fmt.Println("No active token found; skipping order book demo.")
		return
	}

	fmt.Printf("=== Order Book (token %s) ===\n", firstTokenID)

	book, err := c.GetOrderBook(ctx, firstTokenID)
	if err != nil {
		log.Fatalf("GetOrderBook: %v", err)
	}
	fmt.Printf("  Asset ID:   %s\n", book.AssetID)
	fmt.Printf("  Tick Size:  %s\n", book.TickSize)
	fmt.Printf("  Neg Risk:   %v\n", book.NegRisk)
	fmt.Printf("  Bids:       %d levels\n", len(book.Bids))
	fmt.Printf("  Asks:       %d levels\n", len(book.Asks))

	maxLevels := 5
	if len(book.Bids) > 0 {
		fmt.Println("  Top bids:")
		for i, lvl := range book.Bids {
			if i >= maxLevels {
				break
			}
			fmt.Printf("    %s @ %s\n", lvl.Size, lvl.Price)
		}
	}
	if len(book.Asks) > 0 {
		fmt.Println("  Top asks:")
		for i, lvl := range book.Asks {
			if i >= maxLevels {
				break
			}
			fmt.Printf("    %s @ %s\n", lvl.Size, lvl.Price)
		}
	}
	fmt.Println()

	// -----------------------------------------------------------------------
	// 4. Midpoint, spread, and last trade price
	// -----------------------------------------------------------------------
	mid, err := c.GetMidpoint(ctx, firstTokenID)
	if err != nil {
		log.Fatalf("GetMidpoint: %v", err)
	}
	fmt.Printf("  Midpoint:          %s\n", mid.String())

	spread, err := c.GetSpread(ctx, firstTokenID)
	if err != nil {
		log.Fatalf("GetSpread: %v", err)
	}
	fmt.Printf("  Spread:            %s (bid=%s, ask=%s)\n", spread.Spread, spread.Bid, spread.Ask)

	lastPrice, err := c.GetLastTradePrice(ctx, firstTokenID)
	if err != nil {
		log.Fatalf("GetLastTradePrice: %v", err)
	}
	fmt.Printf("  Last Trade Price:  %s\n", lastPrice.String())

	bestBuy, err := c.GetPrice(ctx, firstTokenID, polymarket.Buy)
	if err != nil {
		log.Fatalf("GetPrice(BUY): %v", err)
	}
	bestSell, err := c.GetPrice(ctx, firstTokenID, polymarket.Sell)
	if err != nil {
		log.Fatalf("GetPrice(SELL): %v", err)
	}
	fmt.Printf("  Best Buy Price:    %s\n", bestBuy.String())
	fmt.Printf("  Best Sell Price:   %s\n", bestSell.String())
}
