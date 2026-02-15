package ws

import "context"

// UnsubscribeMarket sends an unsubscribe message for the given asset IDs on the market channel.
func (c *Client) UnsubscribeMarket(ctx context.Context, assetIDs ...string) error {
	c.mu.Lock()
	conn := c.marketConn
	c.mu.Unlock()
	if conn == nil {
		return nil
	}
	req := SubscriptionRequest{
		Type:      ChannelMarket,
		Operation: OpUnsubscribe,
		AssetsIDs: assetIDs,
		Markets:   []string{},
	}
	err := conn.sendJSON(req)
	if err != nil {
		return err
	}
	// Remove matching subscriptions from tracking so reconnect won't re-subscribe them.
	conn.removeTrackedAssets(assetIDs)
	return nil
}

// UnsubscribeUser sends an unsubscribe message for the given markets on the user channel.
func (c *Client) UnsubscribeUser(ctx context.Context, markets ...string) error {
	c.mu.Lock()
	conn := c.userConn
	c.mu.Unlock()
	if conn == nil {
		return nil
	}
	req := SubscriptionRequest{
		Type:      ChannelUser,
		Operation: OpUnsubscribe,
		AssetsIDs: []string{},
		Markets:   markets,
	}
	err := conn.sendJSON(req)
	if err != nil {
		return err
	}
	// Remove matching subscriptions from tracking so reconnect won't re-subscribe them.
	conn.removeTrackedMarkets(markets)
	return nil
}
