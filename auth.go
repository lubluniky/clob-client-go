package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// CreateApiKey creates a new API key using L1 (EIP-712 wallet) authentication.
// The nonce should be a unique value for each request (0 is commonly used for
// the first call).
func (c *ClobClient) CreateApiKey(ctx context.Context, nonce int) (*ApiCreds, error) {
	headers, err := c.l1Headers(nonce)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointCreateApiKey, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var creds ApiCreds
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("polymarket: parsing api key response: %w", err)
	}
	return &creds, nil
}

// DeriveApiKey derives an existing API key using L1 (EIP-712 wallet)
// authentication. If a key was previously created for this wallet, this will
// return the same credentials.
func (c *ClobClient) DeriveApiKey(ctx context.Context, nonce int) (*ApiCreds, error) {
	headers, err := c.l1Headers(nonce)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointDeriveApiKey, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var creds ApiCreds
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("polymarket: parsing derived api key response: %w", err)
	}
	return &creds, nil
}

// CreateOrDeriveApiKey attempts to create a new API key. If creation fails
// (e.g., a key already exists), it falls back to deriving the existing key.
// Uses nonce 0 for both calls.
func (c *ClobClient) CreateOrDeriveApiKey(ctx context.Context) (*ApiCreds, error) {
	creds, err := c.CreateApiKey(ctx, 0)
	if err != nil {
		creds, err = c.DeriveApiKey(ctx, 0)
		if err != nil {
			return nil, err
		}
	}
	return creds, nil
}

// GetApiKeys returns all API keys for the authenticated user.
// Requires L2 authentication.
func (c *ClobClient) GetApiKeys(ctx context.Context) ([]ApiKeyResponse, error) {
	headers, err := c.l2Headers("GET", EndpointGetApiKeys, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointGetApiKeys, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	// Try array first, then fall back to single object wrapped in array.
	var keys []ApiKeyResponse
	if err := json.Unmarshal(raw, &keys); err != nil {
		var single ApiKeyResponse
		if err2 := json.Unmarshal(raw, &single); err2 != nil {
			return nil, fmt.Errorf("polymarket: parsing api keys response: %w", err)
		}
		return []ApiKeyResponse{single}, nil
	}
	return keys, nil
}

// DeleteApiKey deletes the current API key. Requires L2 authentication.
func (c *ClobClient) DeleteApiKey(ctx context.Context) error {
	headers, err := c.l2Headers("DELETE", EndpointDeleteApiKey, "")
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointDeleteApiKey, headers, nil)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}
