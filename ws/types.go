package ws

// All WS message types matching the Polymarket WebSocket API.
// Tagged by "event_type" field in JSON.

// Event type constants
const (
	EventBook           = "book"
	EventPriceChange    = "price_change"
	EventTickSizeChange = "tick_size_change"
	EventLastTradePrice = "last_trade_price"
	EventTrade          = "trade"
	EventOrder          = "order"
)

// Channel constants
const (
	ChannelMarket = "market"
	ChannelUser   = "user"
)

// Operation constants
const (
	OpSubscribe   = "subscribe"
	OpUnsubscribe = "unsubscribe"
)

// SubscriptionRequest is the message sent to subscribe/unsubscribe.
type SubscriptionRequest struct {
	Type        string       `json:"type"`
	Operation   string       `json:"operation,omitempty"`
	AssetsIDs   []string     `json:"assets_ids"`
	Markets     []string     `json:"markets"`
	InitialDump *bool        `json:"initial_dump,omitempty"`
	Auth        *AuthPayload `json:"auth,omitempty"`
}

// AuthPayload is embedded in user channel subscription requests.
type AuthPayload struct {
	ApiKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// RawMessage is used for initial JSON parsing to extract event_type.
type RawMessage struct {
	EventType string `json:"event_type"`
}

// BookUpdate represents an orderbook snapshot/update from the "book" event.
type BookUpdate struct {
	AssetID   string      `json:"asset_id"`
	Market    string      `json:"market"`
	Timestamp string      `json:"timestamp"`
	Bids      []BookLevel `json:"bids"`
	Asks      []BookLevel `json:"asks"`
	Hash      string      `json:"hash,omitempty"`
}

// BookLevel represents a single price level in an orderbook update.
type BookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// PriceChange represents a "price_change" event.
type PriceChange struct {
	Market       string             `json:"market"`
	Timestamp    string             `json:"timestamp"`
	PriceChanges []PriceChangeEntry `json:"price_changes"`
}

// PriceChangeEntry is a single asset's price change within a PriceChange event.
type PriceChangeEntry struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Size    string `json:"size,omitempty"`
	Side    string `json:"side"`
	Hash    string `json:"hash,omitempty"`
	BestBid string `json:"best_bid,omitempty"`
	BestAsk string `json:"best_ask,omitempty"`
}

// TickSizeChange represents a "tick_size_change" event.
type TickSizeChange struct {
	AssetID     string `json:"asset_id"`
	Market      string `json:"market"`
	OldTickSize string `json:"old_tick_size"`
	NewTickSize string `json:"new_tick_size"`
	Timestamp   string `json:"timestamp"`
}

// LastTradePrice represents a "last_trade_price" event.
type LastTradePrice struct {
	AssetID    string `json:"asset_id"`
	Market     string `json:"market"`
	Price      string `json:"price"`
	Side       string `json:"side,omitempty"`
	Size       string `json:"size,omitempty"`
	FeeRateBps string `json:"fee_rate_bps,omitempty"`
	Timestamp  string `json:"timestamp"`
}

// OrderUpdate represents an "order" event from the user channel.
type OrderUpdate struct {
	ID              string   `json:"id"`
	Market          string   `json:"market"`
	AssetID         string   `json:"asset_id"`
	Side            string   `json:"side"`
	Price           string   `json:"price"`
	Type            string   `json:"type,omitempty"`
	Outcome         string   `json:"outcome,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	OriginalSize    string   `json:"original_size,omitempty"`
	SizeMatched     string   `json:"size_matched,omitempty"`
	Timestamp       string   `json:"timestamp,omitempty"`
	AssociateTrades []string `json:"associate_trades,omitempty"`
	Status          string   `json:"status,omitempty"`
}

// TradeUpdate represents a "trade" event from the user channel.
type TradeUpdate struct {
	ID              string      `json:"id"`
	Market          string      `json:"market"`
	AssetID         string      `json:"asset_id"`
	Side            string      `json:"side"`
	Size            string      `json:"size"`
	Price           string      `json:"price"`
	Status          string      `json:"status"`
	Type            string      `json:"type,omitempty"`
	LastUpdate      string      `json:"last_update,omitempty"`
	MatchTime       string      `json:"match_time,omitempty"`
	Timestamp       string      `json:"timestamp,omitempty"`
	Outcome         string      `json:"outcome,omitempty"`
	Owner           string      `json:"owner,omitempty"`
	TakerOrderID    string      `json:"taker_order_id,omitempty"`
	MakerOrders     []MakerFill `json:"maker_orders,omitempty"`
	FeeRateBps      string      `json:"fee_rate_bps,omitempty"`
	TransactionHash string      `json:"transaction_hash,omitempty"`
	TraderSide      string      `json:"trader_side,omitempty"`
}

// MakerFill represents a maker-side fill within a trade update.
type MakerFill struct {
	AssetID       string `json:"asset_id"`
	MatchedAmount string `json:"matched_amount"`
	OrderID       string `json:"order_id"`
	Outcome       string `json:"outcome"`
	Owner         string `json:"owner"`
	Price         string `json:"price"`
}
