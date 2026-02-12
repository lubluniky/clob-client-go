package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DefaultEndpoint   = "wss://ws-subscriptions-clob.polymarket.com"
	PingInterval      = 5 * time.Second
	PongTimeout       = 15 * time.Second
	InitialBackoff    = 1 * time.Second
	MaxBackoff        = 60 * time.Second
	BackoffMultiplier = 2.0
	channelBufferSize = 256
)

// Client is a WebSocket client for the Polymarket real-time data API.
type Client struct {
	endpoint string

	mu         sync.Mutex
	marketConn *connection
	userConn   *connection
}

// Option configures the WebSocket client.
type Option func(*Client)

// WithEndpoint overrides the default WebSocket endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) { c.endpoint = endpoint }
}

// NewClient creates a new WebSocket client.
func NewClient(opts ...Option) *Client {
	c := &Client{endpoint: DefaultEndpoint}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// connection manages a single WebSocket connection (market or user channel).
type connection struct {
	url    string
	ctx    context.Context
	cancel context.CancelFunc
	conn   *websocket.Conn
	connMu sync.Mutex

	// Subscription tracking for reconnection
	subsMu sync.Mutex
	subs   []SubscriptionRequest // tracked for re-subscribe on reconnect

	// Message broadcast
	listeners []listener
	listMu    sync.Mutex

	// Heartbeat tracking
	lastPong time.Time
	pongMu   sync.Mutex
}

type listener struct {
	eventType string // filter by event_type, empty = all
	ch        chan json.RawMessage
}

// newConnection creates and starts a connection to the given WS URL.
func newConnection(parentCtx context.Context, url string) *connection {
	ctx, cancel := context.WithCancel(parentCtx)
	c := &connection{
		url:    url,
		ctx:    ctx,
		cancel: cancel,
	}
	go c.connectLoop()
	return c
}

// connectLoop manages connect -> read -> reconnect cycle.
func (c *connection) connectLoop() {
	var attempt int
	for {
		if c.ctx.Err() != nil {
			return
		}

		conn, _, err := websocket.DefaultDialer.DialContext(c.ctx, c.url, nil)
		if err != nil {
			attempt++
			c.backoff(attempt)
			continue
		}

		c.connMu.Lock()
		c.conn = conn
		c.connMu.Unlock()

		// Reset on successful connect
		attempt = 0

		// Re-subscribe all tracked subscriptions
		c.resubscribe()

		// Start heartbeat
		heartbeatCtx, heartbeatCancel := context.WithCancel(c.ctx)
		go c.heartbeatLoop(heartbeatCtx)

		// Read messages until error
		c.readLoop()

		// Connection lost
		heartbeatCancel()
		c.connMu.Lock()
		c.conn.Close()
		c.conn = nil
		c.connMu.Unlock()

		if c.ctx.Err() != nil {
			return
		}
		attempt++
		c.backoff(attempt)
	}
}

// readLoop reads messages from the WebSocket and dispatches to listeners.
func (c *connection) readLoop() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		text := string(message)

		// Handle PONG
		if text == "PONG" {
			c.pongMu.Lock()
			c.lastPong = time.Now()
			c.pongMu.Unlock()
			continue
		}

		// Try to parse as array (batched messages)
		c.dispatch(message)
	}
}

// dispatch routes raw JSON to appropriate listeners.
func (c *connection) dispatch(data []byte) {
	// Check if it's an array
	data = trimSpace(data)
	if len(data) > 0 && data[0] == '[' {
		var messages []json.RawMessage
		if err := json.Unmarshal(data, &messages); err != nil {
			return
		}
		for _, msg := range messages {
			c.dispatchSingle(msg)
		}
		return
	}
	c.dispatchSingle(data)
}

func (c *connection) dispatchSingle(data []byte) {
	var raw RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	c.listMu.Lock()
	defer c.listMu.Unlock()

	for _, l := range c.listeners {
		if l.eventType == "" || l.eventType == raw.EventType {
			select {
			case l.ch <- json.RawMessage(data):
			default:
				// Drop if channel is full (slow consumer)
			}
		}
	}
}

// heartbeatLoop sends PING messages and checks for PONG responses.
func (c *connection) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.connMu.Lock()
			conn := c.conn
			c.connMu.Unlock()

			if conn == nil {
				return
			}

			// Check if we got a PONG recently
			c.pongMu.Lock()
			lastPong := c.lastPong
			c.pongMu.Unlock()

			if !lastPong.IsZero() && time.Since(lastPong) > PongTimeout {
				// No PONG received within timeout, close connection to trigger reconnect
				conn.Close()
				return
			}

			// Send PING as text message
			if err := conn.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
				return
			}
		}
	}
}

// subscribe adds a listener and sends the subscription request.
func (c *connection) subscribe(req SubscriptionRequest, eventType string) <-chan json.RawMessage {
	ch := make(chan json.RawMessage, channelBufferSize)

	c.listMu.Lock()
	c.listeners = append(c.listeners, listener{eventType: eventType, ch: ch})
	c.listMu.Unlock()

	// Track for reconnect
	c.subsMu.Lock()
	c.subs = append(c.subs, req)
	c.subsMu.Unlock()

	// Send subscription message
	c.sendJSON(req)

	return ch
}

// resubscribe sends all tracked subscription requests (after reconnect).
func (c *connection) resubscribe() {
	c.subsMu.Lock()
	subs := make([]SubscriptionRequest, len(c.subs))
	copy(subs, c.subs)
	c.subsMu.Unlock()

	for _, sub := range subs {
		c.sendJSON(sub)
	}
}

// sendJSON sends a JSON message over the WebSocket.
func (c *connection) sendJSON(v interface{}) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("ws: not connected")
	}
	return c.conn.WriteJSON(v)
}

// backoff sleeps for an exponentially increasing duration with jitter.
func (c *connection) backoff(attempt int) {
	delay := float64(InitialBackoff) * math.Pow(BackoffMultiplier, float64(attempt-1))
	if delay > float64(MaxBackoff) {
		delay = float64(MaxBackoff)
	}
	// Jitter: [0.5, 1.5]
	jitter := 0.5 + rand.Float64()
	actual := time.Duration(delay * jitter)

	timer := time.NewTimer(actual)
	defer timer.Stop()
	select {
	case <-c.ctx.Done():
	case <-timer.C:
	}
}

// close shuts down the connection.
func (c *connection) close() {
	c.cancel()
	c.connMu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.connMu.Unlock()

	// Close all listener channels
	c.listMu.Lock()
	for _, l := range c.listeners {
		close(l.ch)
	}
	c.listeners = nil
	c.listMu.Unlock()
}

func trimSpace(data []byte) []byte {
	for len(data) > 0 && (data[0] == ' ' || data[0] == '\t' || data[0] == '\n' || data[0] == '\r') {
		data = data[1:]
	}
	return data
}

// --- Public API ---

// getMarketConn lazily initializes the market channel connection.
func (c *Client) getMarketConn(ctx context.Context) *connection {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.marketConn == nil {
		c.marketConn = newConnection(ctx, c.endpoint+"/ws/market")
	}
	return c.marketConn
}

// getUserConn lazily initializes the user channel connection.
func (c *Client) getUserConn(ctx context.Context) *connection {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.userConn == nil {
		c.userConn = newConnection(ctx, c.endpoint+"/ws/user")
	}
	return c.userConn
}

// SubscribeOrderBook subscribes to orderbook updates for the given asset IDs.
func (c *Client) SubscribeOrderBook(ctx context.Context, assetIDs ...string) <-chan BookUpdate {
	initialDump := true
	req := SubscriptionRequest{
		Type:        ChannelMarket,
		Operation:   OpSubscribe,
		AssetsIDs:   assetIDs,
		Markets:     []string{},
		InitialDump: &initialDump,
	}

	raw := c.getMarketConn(ctx).subscribe(req, EventBook)
	out := make(chan BookUpdate, channelBufferSize)
	go func() {
		defer close(out)
		for msg := range raw {
			var update BookUpdate
			if json.Unmarshal(msg, &update) == nil {
				select {
				case out <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// SubscribePrices subscribes to price change events for the given asset IDs.
func (c *Client) SubscribePrices(ctx context.Context, assetIDs ...string) <-chan PriceChange {
	initialDump := true
	req := SubscriptionRequest{
		Type:        ChannelMarket,
		Operation:   OpSubscribe,
		AssetsIDs:   assetIDs,
		Markets:     []string{},
		InitialDump: &initialDump,
	}

	raw := c.getMarketConn(ctx).subscribe(req, EventPriceChange)
	out := make(chan PriceChange, channelBufferSize)
	go func() {
		defer close(out)
		for msg := range raw {
			var update PriceChange
			if json.Unmarshal(msg, &update) == nil {
				select {
				case out <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// SubscribeLastTradePrice subscribes to last trade price events for the given asset IDs.
func (c *Client) SubscribeLastTradePrice(ctx context.Context, assetIDs ...string) <-chan LastTradePrice {
	initialDump := true
	req := SubscriptionRequest{
		Type:        ChannelMarket,
		Operation:   OpSubscribe,
		AssetsIDs:   assetIDs,
		Markets:     []string{},
		InitialDump: &initialDump,
	}

	raw := c.getMarketConn(ctx).subscribe(req, EventLastTradePrice)
	out := make(chan LastTradePrice, channelBufferSize)
	go func() {
		defer close(out)
		for msg := range raw {
			var update LastTradePrice
			if json.Unmarshal(msg, &update) == nil {
				select {
				case out <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// SubscribeOrders subscribes to order updates on the user channel.
// Requires API credentials for authentication.
func (c *Client) SubscribeOrders(ctx context.Context, apiKey, secret, passphrase string, markets ...string) <-chan OrderUpdate {
	initialDump := true
	req := SubscriptionRequest{
		Type:        ChannelUser,
		Operation:   OpSubscribe,
		AssetsIDs:   []string{},
		Markets:     markets,
		InitialDump: &initialDump,
		Auth: &AuthPayload{
			ApiKey:     apiKey,
			Secret:     secret,
			Passphrase: passphrase,
		},
	}

	raw := c.getUserConn(ctx).subscribe(req, EventOrder)
	out := make(chan OrderUpdate, channelBufferSize)
	go func() {
		defer close(out)
		for msg := range raw {
			var update OrderUpdate
			if json.Unmarshal(msg, &update) == nil {
				select {
				case out <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// SubscribeTrades subscribes to trade updates on the user channel.
// Requires API credentials for authentication.
func (c *Client) SubscribeTrades(ctx context.Context, apiKey, secret, passphrase string, markets ...string) <-chan TradeUpdate {
	initialDump := true
	req := SubscriptionRequest{
		Type:        ChannelUser,
		Operation:   OpSubscribe,
		AssetsIDs:   []string{},
		Markets:     markets,
		InitialDump: &initialDump,
		Auth: &AuthPayload{
			ApiKey:     apiKey,
			Secret:     secret,
			Passphrase: passphrase,
		},
	}

	raw := c.getUserConn(ctx).subscribe(req, EventTrade)
	out := make(chan TradeUpdate, channelBufferSize)
	go func() {
		defer close(out)
		for msg := range raw {
			var update TradeUpdate
			if json.Unmarshal(msg, &update) == nil {
				select {
				case out <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// Close shuts down all WebSocket connections and closes all subscription channels.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.marketConn != nil {
		c.marketConn.close()
		c.marketConn = nil
	}
	if c.userConn != nil {
		c.userConn.close()
		c.userConn = nil
	}
}
