# clob-client-go

Go client library for the [Polymarket CLOB API](https://docs.polymarket.com/).

Full-featured wrapper with EIP-712 signing, HMAC authentication, WebSocket streaming, precise decimal math, and automatic pagination.

## Install

```bash
go get github.com/lubluniky/clob-client-go
```

Requires Go 1.24+.

## Quick Start

### Public market data (no auth)

```go
package main

import (
    "context"
    "fmt"
    "log"

    polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
    ctx := context.Background()
    client := polymarket.NewClobClient()

    // Server time
    ts, _ := client.ServerTime(ctx)
    fmt.Println("Server time:", ts)

    // Iterate all markets (auto-paginated)
    for market, err := range client.GetMarkets(ctx) {
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("%s (tokens: %d)\n", market.Question, len(market.Tokens))
    }
}
```

### Authenticated operations

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"

    polymarket "github.com/lubluniky/clob-client-go"
)

func main() {
    ctx := context.Background()

    key, _ := crypto.HexToECDSA("your_private_key_hex")

    client := polymarket.NewClobClient(
        polymarket.WithSigner(key),
    )

    // Derive API key (L1 auth - EIP-712 wallet signing)
    creds, err := client.CreateOrDeriveApiKey(ctx)
    if err != nil {
        log.Fatal(err)
    }
    client.SetApiCreds(*creds)

    // Now use L2 endpoints
    for trade, err := range client.GetTrades(ctx, polymarket.TradeParams{}) {
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Trade %s: %s @ %s\n", trade.ID, trade.Size, trade.Price)
    }

    // Create and post a limit order
    signed, _ := client.CreateOrder(ctx, polymarket.OrderArgs{
        TokenID: "your_token_id",
        Price:   decimal.NewFromFloat(0.50),
        Size:    decimal.NewFromFloat(10),
        Side:    polymarket.Buy,
    })
    resp, _ := client.PostOrder(ctx, *signed, polymarket.GTC, false)
    fmt.Println("Order:", resp.ID, resp.Status)
}
```

### WebSocket streaming

```go
package main

import (
    "context"
    "fmt"
    "os/signal"
    "syscall"

    "github.com/lubluniky/clob-client-go/ws"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
    defer cancel()

    client := ws.NewClient()
    defer client.Close()

    books := client.SubscribeOrderBook(ctx, "token_id_here")
    prices := client.SubscribePrices(ctx, "token_id_here")

    for {
        select {
        case <-ctx.Done():
            return
        case book := <-books:
            fmt.Printf("Book: bids=%d asks=%d\n", len(book.Bids), len(book.Asks))
        case price := <-prices:
            fmt.Printf("Price changes: %d\n", len(price.PriceChanges))
        }
    }
}
```

## Authentication Levels

| Level | Method | Use Case |
|-------|--------|----------|
| **L0** | None | Public market data, orderbook, prices |
| **L1** | EIP-712 wallet signature | API key creation/derivation |
| **L2** | HMAC-SHA256 with API key | Orders, trades, account, RFQ |

## API Coverage

### Market Data (L0)
`GetOk`, `ServerTime`/`GetServerTime`, `GetMarkets`, `GetSamplingMarkets`, `GetMarket`, `GetSimplifiedMarkets`, `GetSamplingSimplifiedMarkets`, `GetOrderBook`, `GetOrderBooks`, `GetMidpoint`, `GetPrice`, `GetSpread`, `GetLastTradePrice`, `GetPricesHistory` + batch variants

### Orders (L2)
`CreateOrder`, `CreateMarketOrder`, `CalculateMarketPrice`, `PostOrder`, `PostOrders`, `CreateAndPostOrder`, `CreateAndPostMarketOrder`, `CancelOrder`, `CancelOrders`, `CancelMarketOrders`, `CancelAll`, `GetOrder`, `GetOpenOrders`

### Trades (L2)
`GetTrades`, `GetTradesPaginated`, `GetMarketTradesEvents`

### Account (L2)
`GetBalanceAllowance`, `UpdateBalanceAllowance`, `GetNotifications`, `DropNotifications`, `PostHeartbeat`, `GetClosedOnlyMode`

### Auth (L1/L2)
`CreateApiKey`, `DeriveApiKey`, `CreateOrDeriveApiKey`, `GetApiKeys`, `DeleteApiKey`, `CreateReadonlyApiKey`, `GetReadonlyApiKeys`, `DeleteReadonlyApiKey`, `ValidateReadonlyApiKey`

### Builder (L2)
`CreateBuilderApiKey`, `GetBuilderApiKeys`, `RevokeBuilderApiKey`, `GetBuilderTrades`

### Scoring (L2)
`IsOrderScoring`, `AreOrdersScoring`

### RFQ (L2)
`CreateRfqRequest`, `CancelRfqRequest`, `GetRfqRequests`, `CreateRfqQuote`, `CancelRfqQuote`, `GetRfqRequesterQuotes`, `GetRfqQuoterQuotes`, `GetRfqBestQuote`, `AcceptRfqRequest`, `ApproveRfqQuote`, `GetRfqConfig`

### Rewards (L0/L2)
`GetEarningsForDay`/`GetEarningsForUserForDay`, `GetTotalEarnings`/`GetTotalEarningsForUserForDay`, `GetRewardPercentages`, `GetCurrentRewardsMarkets`/`GetCurrentRewards`, `GetRewardsForMarket`/`GetRawRewardsForMarket`, `GetUserMarketRewards`/`GetUserEarningsAndMarketsConfig`

### WebSocket
`SubscribeOrderBook`, `SubscribePrices`, `SubscribeLastTradePrice`, `SubscribeTickSizeChange`, `SubscribeOrders`, `SubscribeTrades`, `UnsubscribeMarket`, `UnsubscribeUser`

## Features

- **Precise decimals** via `shopspring/decimal` - no floating point bugs
- **Automatic pagination** with Go 1.23+ `iter.Seq2` range iterators
- **Retry with backoff** - exponential backoff, jitter, Retry-After support
- **EIP-712 signing** for wallet authentication (L1)
- **HMAC-SHA256 signing** for API key authentication (L2)
- **WebSocket** with auto-reconnect, heartbeat (PING/PONG), and subscription management
- **Tick-size-aware rounding** for all 4 tick sizes (0.1, 0.01, 0.001, 0.0001)
- **Context support** throughout for timeouts and cancellation

## Configuration

```go
client := polymarket.NewClobClient(
    polymarket.WithSigner(key),                          // ECDSA private key
    polymarket.WithAddress("0x..."),                     // Optional explicit address for L2 when signer is absent
    polymarket.WithFunderAddress("0x..."),               // Optional maker/funder address for signed orders
    polymarket.WithCreds(polymarket.ApiCreds{...}),      // API credentials
    polymarket.WithSignatureType(polymarket.EOA),        // Default signature type
    polymarket.WithTickSizeTTL(time.Minute),             // Tick-size cache TTL (<=0 disables expiry)
    polymarket.WithBaseURL("https://clob.polymarket.com"), // Custom base URL
    polymarket.WithChainID(137),                         // Chain ID (137=Polygon, 80002=Amoy)
    polymarket.WithHTTPOptions(
        transport.WithTimeout(30 * time.Second),
        transport.WithMaxRetries(5),
    ),
)
```

WebSocket lifecycle can be tied to a caller context without breaking existing behavior:

```go
wsClient := ws.NewClient(ws.WithConnectionContext(ctx))
```

## License

[MIT](LICENSE)
