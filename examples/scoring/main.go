package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
	privateKeyHex := os.Getenv("POLY_PRIVATE_KEY")
	apiKey := os.Getenv("POLY_API_KEY")
	apiSecret := os.Getenv("POLY_API_SECRET")
	apiPassphrase := os.Getenv("POLY_API_PASSPHRASE")
	orderIDs := strings.Split(os.Getenv("POLY_ORDER_IDS"), ",")

	if privateKeyHex == "" || apiKey == "" || apiSecret == "" || apiPassphrase == "" || len(orderIDs) == 0 || orderIDs[0] == "" {
		log.Fatal("Set POLY_PRIVATE_KEY, POLY_API_KEY, POLY_API_SECRET, POLY_API_PASSPHRASE, POLY_ORDER_IDS=id1,id2")
	}

	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("invalid private key: %v", err)
	}

	ctx := context.Background()
	c := polymarket.NewClobClient(
		polymarket.WithSigner(key),
		polymarket.WithCreds(polymarket.ApiCreds{
			ApiKey:        apiKey,
			ApiSecret:     apiSecret,
			ApiPassphrase: apiPassphrase,
		}),
	)

	firstID := strings.TrimSpace(orderIDs[0])
	one, err := c.IsOrderScoring(ctx, firstID)
	if err != nil {
		log.Fatalf("IsOrderScoring: %v", err)
	}
	log.Printf("order %s scoring=%v", firstID, one.Scoring)

	trimmed := make([]string, 0, len(orderIDs))
	for _, id := range orderIDs {
		trimmed = append(trimmed, strings.TrimSpace(id))
	}
	many, err := c.AreOrdersScoring(ctx, trimmed)
	if err != nil {
		log.Fatalf("AreOrdersScoring: %v", err)
	}
	log.Printf("orders scoring: %#v", many)
}
