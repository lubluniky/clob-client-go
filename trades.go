package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// GetTrades returns an iterator over trades with auto-pagination. Requires L2
// authentication. The optional params filter by market, asset, maker, or time
// range.
func (c *ClobClient) GetTrades(ctx context.Context, params TradeParams) iter.Seq2[Trade, error] {
	return paginate[Trade](ctx, func(cursor string) (PaginatedResponse[Trade], error) {
		headers, err := c.l2Headers("GET", EndpointTrades, "")
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
		resp, err := c.http.Get(ctx, EndpointTrades, headers, query)
		if err != nil {
			return PaginatedResponse[Trade]{}, err
		}
		body, err := transport.ParseResponse(resp)
		if err != nil {
			return PaginatedResponse[Trade]{}, err
		}
		var page PaginatedResponse[Trade]
		if err := json.Unmarshal(body, &page); err != nil {
			return PaginatedResponse[Trade]{}, fmt.Errorf("polymarket: parsing trades: %w", err)
		}
		return page, nil
	})
}

// GetTradesPaginated returns one page of trades and server paging metadata.
func (c *ClobClient) GetTradesPaginated(ctx context.Context, params TradeParams, cursor string) (PaginatedResponse[Trade], error) {
	headers, err := c.l2Headers("GET", EndpointTrades, "")
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

	resp, err := c.http.Get(ctx, EndpointTrades, headers, query)
	if err != nil {
		return PaginatedResponse[Trade]{}, err
	}
	body, err := transport.ParseResponse(resp)
	if err != nil {
		return PaginatedResponse[Trade]{}, err
	}
	var page PaginatedResponse[Trade]
	if err := json.Unmarshal(body, &page); err != nil {
		return PaginatedResponse[Trade]{}, fmt.Errorf("polymarket: parsing trades page: %w", err)
	}
	return page, nil
}

// GetMarketTradesEvents returns trade events for a market identified by its
// condition ID. This is a public (L0) endpoint.
func (c *ClobClient) GetMarketTradesEvents(ctx context.Context, conditionID string) ([]TradeEvent, error) {
	path := EndpointMarketTradesEvents + conditionID

	resp, err := c.http.Get(ctx, path, c.l0Headers(), nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var events []TradeEvent
	if err := json.Unmarshal(raw, &events); err != nil {
		return nil, fmt.Errorf("polymarket: parsing trade events: %w", err)
	}
	return events, nil
}
