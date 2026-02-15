package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// CreateBuilderApiKey creates builder API credentials.
// Requires L2 authentication.
func (c *ClobClient) CreateBuilderApiKey(ctx context.Context) (*BuilderApiKey, error) {
	headers, err := c.l2Headers("POST", EndpointCreateBuilderApiKey, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointCreateBuilderApiKey, headers, nil)
	if err != nil {
		return nil, err
	}
	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result BuilderApiKey
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing builder api key: %w", err)
	}
	return &result, nil
}

// GetBuilderApiKeys lists builder API keys.
// Requires L2 authentication.
func (c *ClobClient) GetBuilderApiKeys(ctx context.Context) ([]BuilderApiKey, error) {
	headers, err := c.l2Headers("GET", EndpointGetBuilderApiKeys, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointGetBuilderApiKeys, headers, nil)
	if err != nil {
		return nil, err
	}
	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result []BuilderApiKey
	if err := json.Unmarshal(raw, &result); err == nil {
		return result, nil
	}
	var single BuilderApiKey
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, fmt.Errorf("polymarket: parsing builder api keys: %w", err)
	}
	return []BuilderApiKey{single}, nil
}

// RevokeBuilderApiKey revokes active builder API credentials.
// Requires L2 authentication.
func (c *ClobClient) RevokeBuilderApiKey(ctx context.Context) error {
	headers, err := c.l2Headers("DELETE", EndpointRevokeBuilderApiKey, "")
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointRevokeBuilderApiKey, headers, nil)
	if err != nil {
		return err
	}
	_, err = transport.ParseResponse(resp)
	return err
}

// GetBuilderTrades returns an iterator over builder-originated trades.
// Requires L2 authentication.
func (c *ClobClient) GetBuilderTrades(ctx context.Context, params TradeParams) iter.Seq2[Trade, error] {
	return paginate[Trade](ctx, func(cursor string) (PaginatedResponse[Trade], error) {
		headers, err := c.l2Headers("GET", EndpointBuilderTrades, "")
		if err != nil {
			return PaginatedResponse[Trade]{}, err
		}
		query := make(map[string]string)
		if params.Market != "" {
			query["market"] = params.Market
		}
		if params.AssetID != "" {
			query["asset_id"] = params.AssetID
		}
		if params.Maker != "" {
			query["maker_address"] = params.Maker
		}
		if params.ID != "" {
			query["id"] = params.ID
		}
		if params.Before != "" {
			query["before"] = params.Before
		}
		if params.After != "" {
			query["after"] = params.After
		}
		if cursor != "" {
			query["next_cursor"] = cursor
		}

		resp, err := c.http.Get(ctx, EndpointBuilderTrades, headers, query)
		if err != nil {
			return PaginatedResponse[Trade]{}, err
		}
		body, err := transport.ParseResponse(resp)
		if err != nil {
			return PaginatedResponse[Trade]{}, err
		}
		var page PaginatedResponse[Trade]
		if err := json.Unmarshal(body, &page); err != nil {
			return PaginatedResponse[Trade]{}, fmt.Errorf("polymarket: parsing builder trades: %w", err)
		}
		return page, nil
	})
}
