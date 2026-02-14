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
	heartbeatID := os.Getenv("POLY_HEARTBEAT_ID")

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

	// Start a heartbeat chain by sending null heartbeat_id.
	if err := c.PostHeartbeat(ctx, ""); err != nil {
		log.Fatalf("PostHeartbeat(start): %v", err)
	}
	log.Println("heartbeat chain started")

	// Continue an existing chain if heartbeat ID is known.
	if heartbeatID != "" {
		if err := c.PostHeartbeat(ctx, heartbeatID); err != nil {
			log.Fatalf("PostHeartbeat(continue): %v", err)
		}
		log.Println("heartbeat chain continued")
	}
}
