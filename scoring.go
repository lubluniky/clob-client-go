package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// IsOrderScoring checks if an order currently scores for rewards.
// Requires L2 authentication.
func (c *ClobClient) IsOrderScoring(ctx context.Context, orderID string) (*OrderScoringResponse, error) {
	headers, err := c.l2Headers("GET", EndpointOrderScoring, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointOrderScoring, headers, map[string]string{
		"order_id": orderID,
	})
	if err != nil {
		return nil, err
	}
	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result OrderScoringResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing order scoring: %w", err)
	}
	return &result, nil
}

// AreOrdersScoring checks scoring status for multiple orders.
// Requires L2 authentication.
func (c *ClobClient) AreOrdersScoring(ctx context.Context, orderIDs []string) (OrdersScoringResponse, error) {
	bodyBytes, err := json.Marshal(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("polymarket: marshalling orders scoring request: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointOrdersScoring, string(bodyBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointOrdersScoring, headers, orderIDs)
	if err != nil {
		return nil, err
	}
	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result OrdersScoringResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing orders scoring: %w", err)
	}
	return result, nil
}
