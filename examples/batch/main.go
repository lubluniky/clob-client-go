package main

import (
	"context"
	"log"
	"os"
	"strings"

	polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
	tokenIDs := strings.Split(os.Getenv("POLY_TOKEN_IDS"), ",")
	if len(tokenIDs) == 0 || tokenIDs[0] == "" {
		log.Fatal("Set POLY_TOKEN_IDS=token1,token2,...")
	}

	ctx := context.Background()
	c := polymarket.NewClobClient()

	params := make([]polymarket.BookParams, 0, len(tokenIDs))
	for _, id := range tokenIDs {
		params = append(params, polymarket.BookParams{TokenID: strings.TrimSpace(id)})
	}

	if _, err := c.GetOrderBooks(ctx, params); err != nil {
		log.Fatalf("GetOrderBooks: %v", err)
	}
	if _, err := c.GetMidpoints(ctx, tokenIDs); err != nil {
		log.Fatalf("GetMidpoints: %v", err)
	}
	if _, err := c.GetPrices(ctx, tokenIDs, polymarket.Buy); err != nil {
		log.Fatalf("GetPrices: %v", err)
	}
	if _, err := c.GetSpreads(ctx, tokenIDs); err != nil {
		log.Fatalf("GetSpreads: %v", err)
	}
	if _, err := c.GetLastTradesPrices(ctx, tokenIDs); err != nil {
		log.Fatalf("GetLastTradesPrices: %v", err)
	}
}
