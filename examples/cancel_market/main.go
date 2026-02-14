package main

import (
	"context"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/crypto"

	polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
	privateKeyHex := os.Getenv("POLY_PRIVATE_KEY")
	apiKey := os.Getenv("POLY_API_KEY")
	apiSecret := os.Getenv("POLY_API_SECRET")
	apiPassphrase := os.Getenv("POLY_API_PASSPHRASE")

	market := os.Getenv("POLY_MARKET")
	assetID := os.Getenv("POLY_ASSET_ID")

	if privateKeyHex == "" || apiKey == "" || apiSecret == "" || apiPassphrase == "" {
		log.Fatal("Set POLY_PRIVATE_KEY, POLY_API_KEY, POLY_API_SECRET, POLY_API_PASSPHRASE")
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

	if err := c.CancelMarketOrders(ctx, market, assetID); err != nil {
		log.Fatalf("CancelMarketOrders: %v", err)
	}
	log.Println("cancel market orders request sent")
}
