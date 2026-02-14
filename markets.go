package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strconv"

	"github.com/lubluniky/clob-client-go/internal/transport"
	"github.com/shopspring/decimal"
)

// ---------------------------------------------------------------------------
// Single-resource market data methods (all L0 / public)
// ---------------------------------------------------------------------------

// GetOk returns a simple health-check payload from the root endpoint.
func (c *ClobClient) GetOk(ctx context.Context) (string, error) {
	resp, err := c.http.Get(ctx, "/", c.l0Headers(), nil)
	if err != nil {
		return "", err
	}
	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return "", err
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	return string(raw), nil
}

// ServerTime returns the server's current unix timestamp.
func (c *ClobClient) ServerTime(ctx context.Context) (int64, error) {
	raw, err := c.getJSON(ctx, EndpointTime, nil)
	if err != nil {
		return 0, err
	}
	// The API returns a bare number (not wrapped in an object).
	var ts int64
	if err := json.Unmarshal(raw, &ts); err != nil {
		return 0, fmt.Errorf("polymarket: parsing server time: %w", err)
	}
	return ts, nil
}

// GetMarkets returns an iterator over all markets with auto-pagination.
func (c *ClobClient) GetMarkets(ctx context.Context) iter.Seq2[Market, error] {
	return paginate[Market](ctx, func(cursor string) (PaginatedResponse[Market], error) {
		query := map[string]string{}
		if cursor != "" {
			query["next_cursor"] = cursor
		}
		raw, err := c.getJSON(ctx, EndpointMarkets, query)
		if err != nil {
			return PaginatedResponse[Market]{}, err
		}
		var page PaginatedResponse[Market]
		if err := json.Unmarshal(raw, &page); err != nil {
			return PaginatedResponse[Market]{}, fmt.Errorf("polymarket: parsing markets: %w", err)
		}
		return page, nil
	})
}

// GetSamplingMarkets returns an iterator over sampling markets.
func (c *ClobClient) GetSamplingMarkets(ctx context.Context) iter.Seq2[Market, error] {
	return paginate[Market](ctx, func(cursor string) (PaginatedResponse[Market], error) {
		query := map[string]string{}
		if cursor != "" {
			query["next_cursor"] = cursor
		}
		raw, err := c.getJSON(ctx, EndpointSamplingMarkets, query)
		if err != nil {
			return PaginatedResponse[Market]{}, err
		}
		var page PaginatedResponse[Market]
		if err := json.Unmarshal(raw, &page); err != nil {
			return PaginatedResponse[Market]{}, fmt.Errorf("polymarket: parsing sampling markets: %w", err)
		}
		return page, nil
	})
}

// GetMarket returns a single market by condition ID.
func (c *ClobClient) GetMarket(ctx context.Context, conditionID string) (*Market, error) {
	raw, err := c.getJSON(ctx, EndpointMarket+conditionID, nil)
	if err != nil {
		return nil, err
	}
	var m Market
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("polymarket: parsing market: %w", err)
	}
	return &m, nil
}

// GetSimplifiedMarkets returns an iterator over simplified markets.
func (c *ClobClient) GetSimplifiedMarkets(ctx context.Context) iter.Seq2[SimplifiedMarket, error] {
	return paginate[SimplifiedMarket](ctx, func(cursor string) (PaginatedResponse[SimplifiedMarket], error) {
		query := map[string]string{}
		if cursor != "" {
			query["next_cursor"] = cursor
		}
		raw, err := c.getJSON(ctx, EndpointSimplifiedMarkets, query)
		if err != nil {
			return PaginatedResponse[SimplifiedMarket]{}, err
		}
		var page PaginatedResponse[SimplifiedMarket]
		if err := json.Unmarshal(raw, &page); err != nil {
			return PaginatedResponse[SimplifiedMarket]{}, fmt.Errorf("polymarket: parsing simplified markets: %w", err)
		}
		return page, nil
	})
}

// GetSamplingSimplifiedMarkets returns an iterator over sampling simplified markets.
func (c *ClobClient) GetSamplingSimplifiedMarkets(ctx context.Context) iter.Seq2[SimplifiedMarket, error] {
	return paginate[SimplifiedMarket](ctx, func(cursor string) (PaginatedResponse[SimplifiedMarket], error) {
		query := map[string]string{}
		if cursor != "" {
			query["next_cursor"] = cursor
		}
		raw, err := c.getJSON(ctx, EndpointSamplingSimplifiedMarkets, query)
		if err != nil {
			return PaginatedResponse[SimplifiedMarket]{}, err
		}
		var page PaginatedResponse[SimplifiedMarket]
		if err := json.Unmarshal(raw, &page); err != nil {
			return PaginatedResponse[SimplifiedMarket]{}, fmt.Errorf("polymarket: parsing sampling simplified markets: %w", err)
		}
		return page, nil
	})
}

// GetOrderBook returns the order book for a token.
func (c *ClobClient) GetOrderBook(ctx context.Context, tokenID string) (*OrderBookSummary, error) {
	raw, err := c.getJSON(ctx, EndpointOrderBook, map[string]string{"token_id": tokenID})
	if err != nil {
		return nil, err
	}
	var ob OrderBookSummary
	if err := json.Unmarshal(raw, &ob); err != nil {
		return nil, fmt.Errorf("polymarket: parsing order book: %w", err)
	}
	return &ob, nil
}

// GetOrderBooks returns order books for multiple tokens.
func (c *ClobClient) GetOrderBooks(ctx context.Context, params []BookParams) ([]OrderBookSummary, error) {
	resp, err := c.http.Post(ctx, EndpointOrderBooks, c.l0Headers(), params)
	if err != nil {
		return nil, err
	}
	data, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	var result []OrderBookSummary
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing order books: %w", err)
	}
	return result, nil
}

// GetMidpoint returns the midpoint price for a token.
func (c *ClobClient) GetMidpoint(ctx context.Context, tokenID string) (decimal.Decimal, error) {
	raw, err := c.getJSON(ctx, EndpointMidpoint, map[string]string{"token_id": tokenID})
	if err != nil {
		return decimal.Zero, err
	}
	var resp MidpointResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return decimal.Zero, fmt.Errorf("polymarket: parsing midpoint: %w", err)
	}
	return decimal.NewFromString(resp.Mid)
}

// GetPrice returns the best price for a given side.
func (c *ClobClient) GetPrice(ctx context.Context, tokenID string, side Side) (decimal.Decimal, error) {
	raw, err := c.getJSON(ctx, EndpointPrice, map[string]string{
		"token_id": tokenID,
		"side":     string(side),
	})
	if err != nil {
		return decimal.Zero, err
	}
	var resp PriceResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return decimal.Zero, fmt.Errorf("polymarket: parsing price: %w", err)
	}
	return decimal.NewFromString(resp.Price)
}

// GetSpread returns the bid-ask spread for a token.
func (c *ClobClient) GetSpread(ctx context.Context, tokenID string) (*SpreadResponse, error) {
	raw, err := c.getJSON(ctx, EndpointSpread, map[string]string{"token_id": tokenID})
	if err != nil {
		return nil, err
	}
	var resp SpreadResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("polymarket: parsing spread: %w", err)
	}
	return &resp, nil
}

// GetLastTradePrice returns the last traded price for a token.
func (c *ClobClient) GetLastTradePrice(ctx context.Context, tokenID string) (decimal.Decimal, error) {
	raw, err := c.getJSON(ctx, EndpointLastTradePrice, map[string]string{"token_id": tokenID})
	if err != nil {
		return decimal.Zero, err
	}
	var resp LastTradePriceResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return decimal.Zero, fmt.Errorf("polymarket: parsing last trade price: %w", err)
	}
	return decimal.NewFromString(resp.Price)
}

// ---------------------------------------------------------------------------
// Batch market data methods (POST with JSON arrays of token IDs)
// ---------------------------------------------------------------------------

// GetMidpoints returns midpoint prices for multiple tokens.
func (c *ClobClient) GetMidpoints(ctx context.Context, tokenIDs []string) (map[string]string, error) {
	body := make([]BookParams, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		body = append(body, BookParams{TokenID: tokenID})
	}
	resp, err := c.http.Post(ctx, EndpointMidpoints, c.l0Headers(), body)
	if err != nil {
		return nil, err
	}
	data, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing midpoints: %w", err)
	}
	return result, nil
}

// GetPrices returns best prices for multiple tokens.
func (c *ClobClient) GetPrices(ctx context.Context, tokenIDs []string, side Side) (map[string]string, error) {
	body := make([]BookParams, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		body = append(body, BookParams{TokenID: tokenID, Side: side})
	}
	resp, err := c.http.Post(ctx, EndpointPrices, c.l0Headers(), body)
	if err != nil {
		return nil, err
	}
	data, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing prices: %w", err)
	}
	return result, nil
}

// GetSpreads returns spreads for multiple tokens.
func (c *ClobClient) GetSpreads(ctx context.Context, tokenIDs []string) (map[string]SpreadResponse, error) {
	body := make([]BookParams, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		body = append(body, BookParams{TokenID: tokenID})
	}
	resp, err := c.http.Post(ctx, EndpointSpreads, c.l0Headers(), body)
	if err != nil {
		return nil, err
	}
	data, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	var result map[string]SpreadResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing spreads: %w", err)
	}
	return result, nil
}

// GetLastTradesPrices returns last trade prices for multiple tokens.
func (c *ClobClient) GetLastTradesPrices(ctx context.Context, tokenIDs []string) (map[string]string, error) {
	body := make([]BookParams, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		body = append(body, BookParams{TokenID: tokenID})
	}
	resp, err := c.http.Post(ctx, EndpointLastTradesPrices, c.l0Headers(), body)
	if err != nil {
		return nil, err
	}
	data, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing last trade prices: %w", err)
	}
	return result, nil
}

// GetPricesHistory returns historical prices for market filters.
func (c *ClobClient) GetPricesHistory(ctx context.Context, params PriceHistoryFilterParams) ([]MarketPrice, error) {
	query := map[string]string{}
	if params.Market != "" {
		query["market"] = params.Market
	}
	if params.StartTS > 0 {
		query["startTs"] = strconv.FormatInt(params.StartTS, 10)
	}
	if params.EndTS > 0 {
		query["endTs"] = strconv.FormatInt(params.EndTS, 10)
	}
	if params.Fidelity > 0 {
		query["fidelity"] = strconv.Itoa(params.Fidelity)
	}
	if params.Interval != "" {
		query["interval"] = string(params.Interval)
	}

	raw, err := c.getJSON(ctx, EndpointPriceHistory, query)
	if err != nil {
		return nil, err
	}
	var prices []MarketPrice
	if err := json.Unmarshal(raw, &prices); err != nil {
		return nil, fmt.Errorf("polymarket: parsing prices history: %w", err)
	}
	return prices, nil
}

// GetServerTime is an alias for ServerTime.
func (c *ClobClient) GetServerTime(ctx context.Context) (int64, error) {
	return c.ServerTime(ctx)
}

// ---------------------------------------------------------------------------
// Cached metadata lookups
// ---------------------------------------------------------------------------

// GetTickSize returns the tick size for a token (cached after first lookup).
func (c *ClobClient) GetTickSize(ctx context.Context, tokenID string) (string, error) {
	if v, ok := c.tickSizes.Load(tokenID); ok {
		return v.(string), nil
	}
	raw, err := c.getJSON(ctx, EndpointTickSize, map[string]string{"token_id": tokenID})
	if err != nil {
		return "", err
	}
	// API returns {"minimum_tick_size": 0.01}
	var resp struct {
		MinimumTickSize json.Number `json:"minimum_tick_size"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("polymarket: parsing tick size: %w", err)
	}
	ts := resp.MinimumTickSize.String()
	c.tickSizes.Store(tokenID, ts)
	return ts, nil
}

// GetNegRisk returns whether a token uses neg-risk (cached after first lookup).
func (c *ClobClient) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	if v, ok := c.negRisk.Load(tokenID); ok {
		return v.(bool), nil
	}
	raw, err := c.getJSON(ctx, EndpointNegRisk, map[string]string{"token_id": tokenID})
	if err != nil {
		return false, err
	}
	// API returns {"neg_risk": true/false}
	var resp struct {
		NegRisk bool `json:"neg_risk"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return false, fmt.Errorf("polymarket: parsing neg risk: %w", err)
	}
	c.negRisk.Store(tokenID, resp.NegRisk)
	return resp.NegRisk, nil
}

// GetFeeRateBps returns the fee rate in basis points for a token (cached after first lookup).
func (c *ClobClient) GetFeeRateBps(ctx context.Context, tokenID string) (string, error) {
	if v, ok := c.feeRates.Load(tokenID); ok {
		return v.(string), nil
	}
	raw, err := c.getJSON(ctx, EndpointFeeRate, map[string]string{"token_id": tokenID})
	if err != nil {
		return "", err
	}
	// API returns {"base_fee": 0}
	var resp struct {
		BaseFee json.Number `json:"base_fee"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("polymarket: parsing fee rate: %w", err)
	}
	fr := resp.BaseFee.String()
	c.feeRates.Store(tokenID, fr)
	return fr, nil
}
