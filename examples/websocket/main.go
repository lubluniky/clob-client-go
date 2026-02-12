// Example: Subscribing to Polymarket WebSocket feeds.
//
// This example demonstrates how to use the ws package to receive
// real-time updates:
//   - Order book snapshots / deltas
//   - Price change events
//   - Last trade price events
//
// Required environment variables:
//
//	POLY_TOKEN_ID - token ID of the asset to subscribe to
//
// Usage:
//
//	export POLY_TOKEN_ID=12345...
//	go run ./examples/websocket
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lubluniky/clob-client-go/ws"
)

func main() {
	tokenID := os.Getenv("POLY_TOKEN_ID")
	if tokenID == "" {
		log.Fatal("Set POLY_TOKEN_ID environment variable")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wsClient := ws.NewClient()
	defer wsClient.Close()

	// -----------------------------------------------------------------------
	// Subscribe to order book updates for a single token
	// -----------------------------------------------------------------------
	fmt.Printf("Subscribing to orderbook for token: %s\n", tokenID)
	books := wsClient.SubscribeOrderBook(ctx, tokenID)

	// -----------------------------------------------------------------------
	// Subscribe to price change events for the same token
	// -----------------------------------------------------------------------
	fmt.Printf("Subscribing to price changes for token: %s\n", tokenID)
	prices := wsClient.SubscribePrices(ctx, tokenID)

	fmt.Println("Listening for updates... (Ctrl+C to stop)")
	fmt.Println()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nShutting down...")
			return

		case book, ok := <-books:
			if !ok {
				fmt.Println("Book channel closed")
				return
			}
			fmt.Printf("[Book] Asset: %s  Market: %s  Bids: %d  Asks: %d  Hash: %s\n",
				book.AssetID, book.Market, len(book.Bids), len(book.Asks), book.Hash)

			// Print top-of-book
			if len(book.Bids) > 0 {
				fmt.Printf("       Best bid: %s @ %s\n", book.Bids[0].Size, book.Bids[0].Price)
			}
			if len(book.Asks) > 0 {
				fmt.Printf("       Best ask: %s @ %s\n", book.Asks[0].Size, book.Asks[0].Price)
			}
			fmt.Println()

		case price, ok := <-prices:
			if !ok {
				fmt.Println("Price channel closed")
				return
			}
			fmt.Printf("[Price] Market: %s  Changes: %d\n", price.Market, len(price.PriceChanges))
			for _, pc := range price.PriceChanges {
				fmt.Printf("        Asset: %s  Price: %s  Side: %s",
					pc.AssetID, pc.Price, pc.Side)
				if pc.BestBid != "" {
					fmt.Printf("  BestBid: %s", pc.BestBid)
				}
				if pc.BestAsk != "" {
					fmt.Printf("  BestAsk: %s", pc.BestAsk)
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}
}
