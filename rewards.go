package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// GetEarningsForDay returns the authenticated user's earnings for the current
// day. The response is returned as raw JSON since its structure may vary.
// Requires L2 authentication.
func (c *ClobClient) GetEarningsForDay(ctx context.Context) (json.RawMessage, error) {
	headers, err := c.l2Headers("GET", EndpointRewardsUser, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRewardsUser, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(raw), nil
}

// GetEarningsForUserForDay is an alias for GetEarningsForDay.
func (c *ClobClient) GetEarningsForUserForDay(ctx context.Context) (json.RawMessage, error) {
	return c.GetEarningsForDay(ctx)
}

// GetTotalEarnings returns the authenticated user's total accumulated earnings.
// The response is returned as raw JSON since its structure may vary.
// Requires L2 authentication.
func (c *ClobClient) GetTotalEarnings(ctx context.Context) (json.RawMessage, error) {
	headers, err := c.l2Headers("GET", EndpointRewardsUserTotal, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRewardsUserTotal, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(raw), nil
}

// GetTotalEarningsForUserForDay is an alias for GetTotalEarnings.
func (c *ClobClient) GetTotalEarningsForUserForDay(ctx context.Context) (json.RawMessage, error) {
	return c.GetTotalEarnings(ctx)
}

// GetRewardPercentages returns reward rate percentages for market assets that
// the authenticated user is participating in.
// Requires L2 authentication.
func (c *ClobClient) GetRewardPercentages(ctx context.Context) ([]RewardPercentage, error) {
	headers, err := c.l2Headers("GET", EndpointRewardsUserPercentages, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRewardsUserPercentages, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var percentages []RewardPercentage
	if err := json.Unmarshal(raw, &percentages); err != nil {
		return nil, fmt.Errorf("polymarket: parsing reward percentages: %w", err)
	}
	return percentages, nil
}

// GetCurrentRewardsMarkets returns markets currently eligible for rewards.
// The response is returned as raw JSON since its structure may vary.
// This is a public (L0) endpoint.
func (c *ClobClient) GetCurrentRewardsMarkets(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.http.Get(ctx, EndpointRewardsMarketsCurrent, c.l0Headers(), nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(raw), nil
}

// GetCurrentRewards is an alias for GetCurrentRewardsMarkets.
func (c *ClobClient) GetCurrentRewards(ctx context.Context) (json.RawMessage, error) {
	return c.GetCurrentRewardsMarkets(ctx)
}

// GetRewardsForMarket returns reward details for a specific market identified
// by its condition ID. The response is returned as raw JSON since its structure
// may vary. This is a public (L0) endpoint.
func (c *ClobClient) GetRewardsForMarket(ctx context.Context, conditionID string) (json.RawMessage, error) {
	path := EndpointRewardsMarket + conditionID

	resp, err := c.http.Get(ctx, path, c.l0Headers(), nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(raw), nil
}

// GetRawRewardsForMarket is an alias for GetRewardsForMarket.
func (c *ClobClient) GetRawRewardsForMarket(ctx context.Context, conditionID string) (json.RawMessage, error) {
	return c.GetRewardsForMarket(ctx, conditionID)
}

// GetUserMarketRewards returns the authenticated user's rewards across all
// markets. The response is returned as raw JSON since its structure may vary.
// Requires L2 authentication.
func (c *ClobClient) GetUserMarketRewards(ctx context.Context) (json.RawMessage, error) {
	headers, err := c.l2Headers("GET", EndpointRewardsUserMarkets, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRewardsUserMarkets, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(raw), nil
}

// GetUserEarningsAndMarketsConfig is an alias for GetUserMarketRewards.
func (c *ClobClient) GetUserEarningsAndMarketsConfig(ctx context.Context) (json.RawMessage, error) {
	return c.GetUserMarketRewards(ctx)
}
