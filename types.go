package client

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

// Side represents a buy or sell side.
type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// OrderType represents the time-in-force / execution strategy for an order.
type OrderType string

const (
	GTC OrderType = "GTC"
	FOK OrderType = "FOK"
	GTD OrderType = "GTD"
	FAK OrderType = "FAK"
)

// SignatureType represents the type of signature used for signing orders.
type SignatureType int

const (
	EOA            SignatureType = 0
	PolyProxy      SignatureType = 1
	PolyGnosisSafe SignatureType = 2
)

// TickSize represents valid tick size strings for markets.
type TickSize string

const (
	TickSizeTenth         TickSize = "0.1"
	TickSizeHundredth     TickSize = "0.01"
	TickSizeThousandth    TickSize = "0.001"
	TickSizeTenThousandth TickSize = "0.0001"
)

// ---------------------------------------------------------------------------
// Market types
// ---------------------------------------------------------------------------

// Market represents the full market object returned by the CLOB API.
type Market struct {
	EnableOrderBook        bool            `json:"enable_order_book"`
	Active                 bool            `json:"active"`
	Closed                 bool            `json:"closed"`
	Archived               bool            `json:"archived"`
	AcceptingOrders        bool            `json:"accepting_orders"`
	AcceptingOrderTimestamp *string         `json:"accepting_order_timestamp,omitempty"`
	MinimumOrderSize       decimal.Decimal `json:"minimum_order_size"`
	MinimumTickSize        decimal.Decimal `json:"minimum_tick_size"`
	ConditionID            string          `json:"condition_id"`
	QuestionID             *string         `json:"question_id,omitempty"`
	Question               string          `json:"question"`
	Description            string          `json:"description"`
	MarketSlug             string          `json:"market_slug"`
	EndDateISO             *string         `json:"end_date_iso,omitempty"`
	GameStartTime          *string         `json:"game_start_time,omitempty"`
	SecondsDelay           int             `json:"seconds_delay"`
	FPMM                   *string         `json:"fpmm,omitempty"`
	MakerBaseFee           decimal.Decimal `json:"maker_base_fee"`
	TakerBaseFee           decimal.Decimal `json:"taker_base_fee"`
	NotificationsEnabled   bool            `json:"notifications_enabled"`
	NegRisk                bool            `json:"neg_risk"`
	NegRiskMarketID        *string         `json:"neg_risk_market_id,omitempty"`
	NegRiskRequestID       *string         `json:"neg_risk_request_id,omitempty"`
	Icon                   string          `json:"icon"`
	Image                  string          `json:"image"`
	Rewards                Rewards         `json:"rewards"`
	Is5050Outcome          bool            `json:"is_50_50_outcome"`
	Tokens                 []Token         `json:"tokens"`
	Tags                   []string        `json:"tags"`
}

// Rewards holds the reward configuration for a market.
type Rewards struct {
	MinSize          decimal.Decimal `json:"min_size"`
	MaxSpread        decimal.Decimal `json:"max_spread"`
	EventStartDate   *string         `json:"event_start_date,omitempty"`
	EventEndDate     *string         `json:"event_end_date,omitempty"`
	InGameMultiplier decimal.Decimal `json:"in_game_multiplier"`
	RewardEpoch      int             `json:"reward_epoch"`
}

// Token represents a single outcome token within a market.
type Token struct {
	TokenID string          `json:"token_id"`
	Outcome string          `json:"outcome"`
	Price   decimal.Decimal `json:"price"`
	Winner  bool            `json:"winner"`
}

// SimplifiedMarket is a lightweight representation of a market with only
// its condition ID and associated tokens.
type SimplifiedMarket struct {
	ConditionID string  `json:"condition_id"`
	Tokens      []Token `json:"tokens"`
}

// ---------------------------------------------------------------------------
// OrderBook types
// ---------------------------------------------------------------------------

// OrderBookSummary represents an order-book snapshot for a single asset.
type OrderBookSummary struct {
	Market         string       `json:"market"`
	AssetID        string       `json:"asset_id"`
	Timestamp      string       `json:"timestamp"`
	Bids           []PriceLevel `json:"bids"`
	Asks           []PriceLevel `json:"asks"`
	MinOrderSize   string       `json:"min_order_size"`
	NegRisk        bool         `json:"neg_risk"`
	TickSize       string       `json:"tick_size"`
	LastTradePrice string       `json:"last_trade_price"`
	Hash           string       `json:"hash"`
}

// PriceLevel represents a single price level (bid or ask) in the order book.
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// ---------------------------------------------------------------------------
// Order types
// ---------------------------------------------------------------------------

// OrderArgs holds the parameters needed to construct a limit order.
type OrderArgs struct {
	TokenID    string
	Price      decimal.Decimal
	Size       decimal.Decimal
	Side       Side
	FeeRateBps int
	Nonce      int
	Expiration int
	Taker      string
}

// MarketOrderArgs holds the parameters needed to construct a market order
// (FOK or FAK).
type MarketOrderArgs struct {
	TokenID    string
	Amount     decimal.Decimal
	Side       Side
	Price      decimal.Decimal // optional; uses worst price from book if zero
	FeeRateBps int
	Nonce      int
	Taker      string
	OrderType  OrderType // FOK or FAK
}

// SignedOrder is the EIP-712 signed order structure sent to the exchange.
type SignedOrder struct {
	Salt          string        `json:"salt"`
	Maker         string        `json:"maker"`
	Signer        string        `json:"signer"`
	Taker         string        `json:"taker"`
	TokenID       string        `json:"tokenId"`
	MakerAmount   string        `json:"makerAmount"`
	TakerAmount   string        `json:"takerAmount"`
	Expiration    string        `json:"expiration"`
	Nonce         string        `json:"nonce"`
	FeeRateBps    string        `json:"feeRateBps"`
	Side          Side          `json:"side"`
	SignatureType SignatureType  `json:"signatureType"`
	Signature     string        `json:"signature"`
}

// PostOrderRequest is the body sent to POST /order.
type PostOrderRequest struct {
	Order     SignedOrder `json:"order"`
	Owner     string     `json:"owner"`
	OrderType OrderType  `json:"orderType"`
	PostOnly  bool       `json:"postOnly,omitempty"`
}

// OrderResponse is returned after posting an order.
type OrderResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

// Order represents a full order record as returned by the data API.
type Order struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	Owner           string   `json:"owner"`
	MakerAddress    string   `json:"maker_address"`
	Market          string   `json:"market"`
	AssetID         string   `json:"asset_id"`
	Side            string   `json:"side"`
	OriginalSize    string   `json:"original_size"`
	SizeMatched     string   `json:"size_matched"`
	Price           string   `json:"price"`
	AssociateTrades []string `json:"associate_trades"`
	Outcome         string   `json:"outcome"`
	CreatedAt       int64    `json:"created_at"`
	Expiration      string   `json:"expiration"`
	OrderType       string   `json:"order_type"`
}

// OpenOrderParams holds optional filter parameters when fetching open orders.
type OpenOrderParams struct {
	Market  string
	AssetID string
}

// ---------------------------------------------------------------------------
// Trade types
// ---------------------------------------------------------------------------

// Trade represents a completed trade record.
type Trade struct {
	ID              string       `json:"id"`
	TakerOrderID    string       `json:"taker_order_id"`
	Market          string       `json:"market"`
	AssetID         string       `json:"asset_id"`
	Side            string       `json:"side"`
	Size            string       `json:"size"`
	FeeRateBps      string       `json:"fee_rate_bps"`
	Price           string       `json:"price"`
	Status          string       `json:"status"`
	MatchTime       string       `json:"match_time"`
	LastUpdate      string       `json:"last_update"`
	Outcome         string       `json:"outcome"`
	BucketIndex     int          `json:"bucket_index"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	MakerOrders     []MakerOrder `json:"maker_orders"`
	TransactionHash string       `json:"transaction_hash"`
	TraderSide      string       `json:"trader_side"`
}

// MakerOrder represents a maker-side fill within a trade.
type MakerOrder struct {
	OrderID       string `json:"order_id"`
	Owner         string `json:"owner"`
	MakerAddress  string `json:"maker_address"`
	MatchedAmount string `json:"matched_amount"`
	Price         string `json:"price"`
	FeeRateBps    string `json:"fee_rate_bps"`
	AssetID       string `json:"asset_id"`
	Outcome       string `json:"outcome"`
	Side          string `json:"side"`
}

// TradeParams holds optional filter parameters when fetching trades.
type TradeParams struct {
	Market  string
	AssetID string
	Maker   string
	Before  string
	After   string
}

// ---------------------------------------------------------------------------
// Auth types
// ---------------------------------------------------------------------------

// ApiCreds holds the three-part credentials for CLOB API authentication.
type ApiCreds struct {
	ApiKey        string `json:"apiKey"`
	ApiSecret     string `json:"secret"`
	ApiPassphrase string `json:"passphrase"`
}

// ApiKeyResponse is returned when creating or listing API keys.
type ApiKeyResponse struct {
	ApiKey    string `json:"apiKey"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// ---------------------------------------------------------------------------
// Account types
// ---------------------------------------------------------------------------

// BalanceAllowance holds balance and allowance for an asset.
type BalanceAllowance struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

// BalanceAllowanceParams holds the parameters for querying balance/allowance.
type BalanceAllowanceParams struct {
	AssetType     string
	TokenID       string
	SignatureType SignatureType
}

// Notification represents a single notification from the API.
// Fields use json.Number because the API returns numeric values for id and type.
type Notification struct {
	ID      json.Number `json:"id"`
	Type    json.Number `json:"type"`
	Message string      `json:"message"`
}

// ---------------------------------------------------------------------------
// Rewards types
// ---------------------------------------------------------------------------

// EarningsResponse wraps the server's rewards/earnings response.
type EarningsResponse struct {
	Data interface{} `json:"data"`
}

// RewardPercentage holds reward rate information for a market asset.
type RewardPercentage struct {
	Market      string `json:"market"`
	AssetID     string `json:"asset_id"`
	RewardRate  string `json:"reward_rate"`
	DailyReward string `json:"daily_reward"`
}

// ---------------------------------------------------------------------------
// RFQ types
// ---------------------------------------------------------------------------

// RfqRequest represents an RFQ (Request For Quote) object.
type RfqRequest struct {
	RequestID       string  `json:"requestId"`
	UserAddress     string  `json:"userAddress"`
	ProxyAddress    string  `json:"proxyAddress"`
	Token           string  `json:"token"`
	Complement      string  `json:"complement"`
	Condition       string  `json:"condition"`
	Side            string  `json:"side"`
	SizeIn          string  `json:"sizeIn"`
	SizeOut         string  `json:"sizeOut"`
	Price           float64 `json:"price"`
	AcceptedQuoteID string  `json:"acceptedQuoteId"`
	State           string  `json:"state"`
	Expiry          string  `json:"expiry"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

// RfqQuote represents a quote submitted in response to an RFQ request.
type RfqQuote struct {
	QuoteID      string  `json:"quoteId"`
	RequestID    string  `json:"requestId"`
	UserAddress  string  `json:"userAddress"`
	ProxyAddress string  `json:"proxyAddress"`
	Complement   string  `json:"complement"`
	Condition    string  `json:"condition"`
	Token        string  `json:"token"`
	Side         string  `json:"side"`
	SizeIn       string  `json:"sizeIn"`
	SizeOut      string  `json:"sizeOut"`
	Price        float64 `json:"price"`
	State        string  `json:"state"`
	Expiry       string  `json:"expiry"`
	MatchType    string  `json:"matchType"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// CreateRfqRequestParams holds the parameters for creating a new RFQ request.
type CreateRfqRequestParams struct {
	AssetIn   string `json:"assetIn"`
	AssetOut  string `json:"assetOut"`
	AmountIn  string `json:"amountIn"`
	AmountOut string `json:"amountOut"`
	UserType  int    `json:"userType"`
}

// CreateRfqQuoteParams holds the parameters for creating a new RFQ quote.
type CreateRfqQuoteParams struct {
	RequestID string `json:"requestId"`
	AssetIn   string `json:"assetIn"`
	AssetOut  string `json:"assetOut"`
	AmountIn  string `json:"amountIn"`
	AmountOut string `json:"amountOut"`
}

// RfqConfig holds the RFQ feature configuration as returned by the API.
type RfqConfig struct {
	Enabled bool `json:"enabled"`
}

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

// PaginationParams holds cursor-based pagination parameters.
type PaginationParams struct {
	NextCursor string `json:"next_cursor,omitempty"`
}

// PaginatedResponse wraps a page of results with a cursor for the next page.
type PaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor"`
}

// ---------------------------------------------------------------------------
// Spread / Midpoint / Price response types
// ---------------------------------------------------------------------------

// MidpointResponse holds the midpoint price for a market asset.
type MidpointResponse struct {
	Mid string `json:"mid"`
}

// PriceResponse holds a single price value.
type PriceResponse struct {
	Price string `json:"price"`
}

// SpreadResponse holds spread, bid, and ask for a market asset.
type SpreadResponse struct {
	Spread string `json:"spread"`
	Bid    string `json:"bid"`
	Ask    string `json:"ask"`
}

// LastTradePriceResponse holds the last traded price for a market asset.
type LastTradePriceResponse struct {
	Price string `json:"price"`
}

// ServerTimeResponse holds the server timestamp.
type ServerTimeResponse struct {
	Timestamp int64 `json:"timestamp"`
}

// ---------------------------------------------------------------------------
// Contract configuration
// ---------------------------------------------------------------------------

// ContractConfig holds the deployed contract addresses for a chain.
type ContractConfig struct {
	Exchange          string
	NegRiskExchange   string
	NegRiskAdapter    string
	Collateral        string
	ConditionalTokens string
}

// PolygonContracts holds the mainnet Polygon contract addresses.
var PolygonContracts = ContractConfig{
	Exchange:          "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
	NegRiskExchange:   "0xC5d563A36AE78145C45a50134d48A1215220f80a",
	NegRiskAdapter:    "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	Collateral:        "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
	ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
}

// AmoyContracts holds the Amoy testnet contract addresses.
var AmoyContracts = ContractConfig{
	Exchange:          "0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40",
	NegRiskExchange:   "0xC5d563A36AE78145C45a50134d48A1215220f80a",
	NegRiskAdapter:    "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	Collateral:        "0x9c4e1703476e875070ee25b56a58b008cfb8fa78",
	ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
}

// Well-known constants.
const (
	ZeroAddress    = "0x0000000000000000000000000000000000000000"
	PolygonChainID = 137
	AmoyChainID    = 80002
)

// ---------------------------------------------------------------------------
// Trade events
// ---------------------------------------------------------------------------

// TradeEvent represents an event from the live-activity/events endpoint.
type TradeEvent struct {
	ConditionID string `json:"conditionId"`
}

// ---------------------------------------------------------------------------
// Order scoring
// ---------------------------------------------------------------------------

// OrderScoringResponse indicates whether an order is scoring for rewards.
type OrderScoringResponse struct {
	Scoring bool `json:"scoring"`
}
