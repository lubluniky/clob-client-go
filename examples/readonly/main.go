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

	readonly, err := c.CreateReadonlyApiKey(ctx)
	if err != nil {
		log.Fatalf("CreateReadonlyApiKey: %v", err)
	}
	log.Printf("created readonly key: %s", readonly.ApiKey)

	keys, err := c.GetReadonlyApiKeys(ctx)
	if err != nil {
		log.Fatalf("GetReadonlyApiKeys: %v", err)
	}
	log.Printf("readonly keys: %v", keys)

	validation, err := c.ValidateReadonlyApiKey(ctx, c.Address(), readonly.ApiKey)
	if err != nil {
		log.Fatalf("ValidateReadonlyApiKey: %v", err)
	}
	log.Printf("validation result: %s", validation)
}
