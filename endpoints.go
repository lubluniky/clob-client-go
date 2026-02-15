package client

const (
	// Server
	EndpointTime = "/time"

	// Markets
	EndpointMarkets                   = "/markets"
	EndpointMarket                    = "/markets/" // append conditionID
	EndpointSimplifiedMarkets         = "/simplified-markets"
	EndpointSamplingSimplifiedMarkets = "/sampling-simplified-markets"
	EndpointSamplingMarkets           = "/sampling-markets"

	// Order Book & Pricing
	EndpointOrderBook        = "/book"
	EndpointOrderBooks       = "/books"
	EndpointMidpoint         = "/midpoint"
	EndpointMidpoints        = "/midpoints"
	EndpointPrice            = "/price"
	EndpointPrices           = "/prices"
	EndpointSpread           = "/spread"
	EndpointSpreads          = "/spreads"
	EndpointLastTradePrice   = "/last-trade-price"
	EndpointLastTradesPrices = "/last-trades-prices"
	EndpointTickSize         = "/tick-size"
	EndpointNegRisk          = "/neg-risk"
	EndpointFeeRate          = "/fee-rate"

	// Orders
	EndpointPostOrder          = "/order"
	EndpointPostOrders         = "/orders"
	EndpointOrder              = "/data/order/" // append orderID
	EndpointOrders             = "/data/orders"
	EndpointCancelOrder        = "/order"
	EndpointCancelOrders       = "/orders"
	EndpointCancelAll          = "/cancel-all"
	EndpointCancelMarketOrders = "/cancel-market-orders"

	// Trades
	EndpointTrades             = "/data/trades"
	EndpointMarketTradesEvents = "/live-activity/events/" // append conditionID

	// Auth
	EndpointCreateApiKey           = "/auth/api-key"
	EndpointGetApiKeys             = "/auth/api-keys"
	EndpointDeleteApiKey           = "/auth/api-key"
	EndpointDeriveApiKey           = "/auth/derive-api-key"
	EndpointClosedOnly             = "/auth/ban-status/closed-only"
	EndpointCreateReadonlyApiKey   = "/auth/readonly-api-key"
	EndpointGetReadonlyApiKeys     = "/auth/readonly-api-keys"
	EndpointDeleteReadonlyApiKey   = "/auth/readonly-api-key"
	EndpointValidateReadonlyApiKey = "/auth/validate-readonly-api-key"
	EndpointCreateBuilderApiKey    = "/auth/builder-api-key"
	EndpointGetBuilderApiKeys      = "/auth/builder-api-key"
	EndpointRevokeBuilderApiKey    = "/auth/builder-api-key"

	// Balance
	EndpointBalanceAllowance       = "/balance-allowance"
	EndpointUpdateBalanceAllowance = "/balance-allowance/update"

	// Notifications
	EndpointNotifications = "/notifications"

	// Order Scoring
	EndpointOrderScoring  = "/order-scoring"
	EndpointOrdersScoring = "/orders-scoring"

	// Heartbeat
	EndpointHeartbeat = "/v1/heartbeats"

	// Builder
	EndpointBuilderTrades = "/builder/trades"

	// Rewards
	EndpointRewardsUser            = "/rewards/user"
	EndpointRewardsUserTotal       = "/rewards/user/total"
	EndpointRewardsUserPercentages = "/rewards/user/percentages"
	EndpointRewardsMarketsCurrent  = "/rewards/markets/current"
	EndpointRewardsMarket          = "/rewards/markets/" // append conditionID
	EndpointRewardsUserMarkets     = "/rewards/user/markets"

	// Price History
	EndpointPriceHistory = "/prices-history"

	// RFQ
	EndpointRfqRequest         = "/rfq/request"
	EndpointRfqRequests        = "/rfq/data/requests"
	EndpointRfqQuote           = "/rfq/quote"
	EndpointRfqRequesterQuotes = "/rfq/data/requester/quotes"
	EndpointRfqQuoterQuotes    = "/rfq/data/quoter/quotes"
	EndpointRfqBestQuote       = "/rfq/data/best-quote"
	EndpointRfqRequestAccept   = "/rfq/request/accept"
	EndpointRfqQuoteApprove    = "/rfq/quote/approve"
	EndpointRfqConfig          = "/rfq/config"
)
