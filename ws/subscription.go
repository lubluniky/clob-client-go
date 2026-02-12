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
	return conn.sendJSON(req)
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
	return conn.sendJSON(req)
}
