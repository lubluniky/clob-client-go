package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	"github.com/lubluniky/clob-client-go/internal/orderbuilder"
	"github.com/lubluniky/clob-client-go/internal/transport"
)

// sideToInt converts a Side enum to the integer representation expected by the
// order builder (0 = Buy, 1 = Sell).
func sideToInt(s Side) int {
	if s == Buy {
		return 0
	}
	return 1
}

// CreateOrder builds and signs a limit order from the given OrderArgs.
// Returns a SignedOrder ready to be posted via PostOrder.
func (c *ClobClient) CreateOrder(ctx context.Context, args OrderArgs) (*SignedOrder, error) {
	if c.signer == nil {
		return nil, &AuthError{Message: "signer key required for creating orders"}
	}

	tickSize, err := c.GetTickSize(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting tick size: %w", err)
	}
	if err := orderbuilder.ValidatePrice(args.Price, tickSize); err != nil {
		return nil, &ValidationError{Field: "price", Message: err.Error()}
	}

	negRisk, err := c.GetNegRisk(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting neg risk: %w", err)
	}
	resolvedFeeRate, err := c.resolveFeeRateBps(ctx, args.TokenID, args.FeeRateBps)
	if err != nil {
		return nil, err
	}

	makerAmt, takerAmt, err := orderbuilder.CalculateLimitOrderAmounts(
		string(args.Side), args.Price, args.Size, tickSize,
	)
	if err != nil {
		return nil, fmt.Errorf("polymarket: calculating order amounts: %w", err)
	}

	taker := args.Taker
	if taker == "" {
		taker = ZeroAddress
	}

	signerAddr := crypto.PubkeyToAddress(c.signer.PublicKey)

	orderData := orderbuilder.OrderData{
		Maker:         c.address,
		Taker:         common.HexToAddress(taker),
		TokenID:       args.TokenID,
		MakerAmount:   makerAmt,
		TakerAmount:   takerAmt,
		Side:          sideToInt(args.Side),
		FeeRateBps:    fmt.Sprintf("%d", resolvedFeeRate),
		Nonce:         fmt.Sprintf("%d", args.Nonce),
		Signer:        signerAddr,
		Expiration:    fmt.Sprintf("%d", args.Expiration),
		SignatureType: int(EOA),
		Salt:          orderbuilder.GenerateSalt(),
	}

	sig, err := orderbuilder.SignOrder(c.signer, c.chainID, orderData, negRisk)
	if err != nil {
		return nil, fmt.Errorf("polymarket: signing order: %w", err)
	}

	return &SignedOrder{
		Salt:          orderData.Salt,
		Maker:         orderData.Maker.Hex(),
		Signer:        orderData.Signer.Hex(),
		Taker:         orderData.Taker.Hex(),
		TokenID:       orderData.TokenID,
		MakerAmount:   orderData.MakerAmount,
		TakerAmount:   orderData.TakerAmount,
		Expiration:    orderData.Expiration,
		Nonce:         orderData.Nonce,
		FeeRateBps:    orderData.FeeRateBps,
		Side:          args.Side,
		SignatureType: EOA,
		Signature:     sig,
	}, nil
}

// CreateMarketOrder builds and signs a market order (FOK/FAK) from the given
// MarketOrderArgs. Returns a SignedOrder ready to be posted via PostOrder.
func (c *ClobClient) CreateMarketOrder(ctx context.Context, args MarketOrderArgs) (*SignedOrder, error) {
	if c.signer == nil {
		return nil, &AuthError{Message: "signer key required for creating orders"}
	}
	if args.OrderType == "" {
		args.OrderType = FOK
	}

	tickSize, err := c.GetTickSize(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting tick size: %w", err)
	}
	if args.Price.LessThanOrEqual(decimal.Zero) {
		marketPrice, err := c.CalculateMarketPrice(ctx, args.TokenID, args.Side, args.Amount, args.OrderType)
		if err != nil {
			return nil, err
		}
		args.Price = marketPrice
	}
	if err := orderbuilder.ValidatePrice(args.Price, tickSize); err != nil {
		return nil, &ValidationError{Field: "price", Message: err.Error()}
	}

	negRisk, err := c.GetNegRisk(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting neg risk: %w", err)
	}
	resolvedFeeRate, err := c.resolveFeeRateBps(ctx, args.TokenID, args.FeeRateBps)
	if err != nil {
		return nil, err
	}

	makerAmt, takerAmt, err := orderbuilder.CalculateMarketOrderAmounts(
		string(args.Side), args.Amount, args.Price, tickSize,
	)
	if err != nil {
		return nil, fmt.Errorf("polymarket: calculating market order amounts: %w", err)
	}

	taker := args.Taker
	if taker == "" {
		taker = ZeroAddress
	}

	signerAddr := crypto.PubkeyToAddress(c.signer.PublicKey)

	orderData := orderbuilder.OrderData{
		Maker:         c.address,
		Taker:         common.HexToAddress(taker),
		TokenID:       args.TokenID,
		MakerAmount:   makerAmt,
		TakerAmount:   takerAmt,
		Side:          sideToInt(args.Side),
		FeeRateBps:    fmt.Sprintf("%d", resolvedFeeRate),
		Nonce:         fmt.Sprintf("%d", args.Nonce),
		Signer:        signerAddr,
		Expiration:    "0",
		SignatureType: int(EOA),
		Salt:          orderbuilder.GenerateSalt(),
	}

	sig, err := orderbuilder.SignOrder(c.signer, c.chainID, orderData, negRisk)
	if err != nil {
		return nil, fmt.Errorf("polymarket: signing market order: %w", err)
	}

	return &SignedOrder{
		Salt:          orderData.Salt,
		Maker:         orderData.Maker.Hex(),
		Signer:        orderData.Signer.Hex(),
		Taker:         orderData.Taker.Hex(),
		TokenID:       orderData.TokenID,
		MakerAmount:   orderData.MakerAmount,
		TakerAmount:   orderData.TakerAmount,
		Expiration:    orderData.Expiration,
		Nonce:         orderData.Nonce,
		FeeRateBps:    orderData.FeeRateBps,
		Side:          args.Side,
		SignatureType: EOA,
		Signature:     sig,
	}, nil
}

// PostOrder submits a signed order to the API. The orderType controls execution
// strategy (GTC, FOK, GTD, FAK). When postOnly is true the order will only be
// accepted if it would rest on the book (no immediate match).
func (c *ClobClient) PostOrder(ctx context.Context, order SignedOrder, orderType OrderType, postOnly bool) (*OrderResponse, error) {
	if postOnly && orderType != GTC && orderType != GTD {
		return nil, &ValidationError{Field: "postOnly", Message: "postOnly is only supported for GTC and GTD orders"}
	}

	owner := ""
	if c.creds != nil {
		owner = c.creds.ApiKey
	}

	req := PostOrderRequest{
		Order:     order,
		Owner:     owner,
		OrderType: orderType,
		PostOnly:  postOnly,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("polymarket: marshalling order request: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointPostOrder, string(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointPostOrder, headers, req)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result OrderResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing order response: %w", err)
	}
	return &result, nil
}

// PostOrders submits a batch of signed orders.
func (c *ClobClient) PostOrders(ctx context.Context, args []PostOrdersArgs, deferExec bool, defaultPostOnly bool) ([]OrderResponse, error) {
	type batchOrderRequest struct {
		Order     SignedOrder `json:"order"`
		Owner     string      `json:"owner"`
		OrderType OrderType   `json:"orderType"`
		PostOnly  bool        `json:"postOnly"`
		DeferExec bool        `json:"deferExec"`
	}

	owner := ""
	if c.creds != nil {
		owner = c.creds.ApiKey
	}

	payload := make([]batchOrderRequest, 0, len(args))
	for _, arg := range args {
		postOnly := defaultPostOnly
		if arg.PostOnly != nil {
			postOnly = *arg.PostOnly
		}
		if postOnly && arg.OrderType != GTC && arg.OrderType != GTD {
			return nil, &ValidationError{Field: "postOnly", Message: "postOnly is only supported for GTC and GTD orders"}
		}
		payload = append(payload, batchOrderRequest{
			Order:     arg.Order,
			Owner:     owner,
			OrderType: arg.OrderType,
			PostOnly:  postOnly,
			DeferExec: deferExec,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("polymarket: marshalling post orders request: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointPostOrders, string(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointPostOrders, headers, payload)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var results []OrderResponse
	if err := json.Unmarshal(raw, &results); err == nil {
		return results, nil
	}

	var single OrderResponse
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, fmt.Errorf("polymarket: parsing post orders response: %w", err)
	}
	return []OrderResponse{single}, nil
}

// CreateAndPostOrder is a convenience method that creates, signs, and posts a
// limit order in a single call.
func (c *ClobClient) CreateAndPostOrder(ctx context.Context, args OrderArgs, orderType OrderType, postOnly bool) (*OrderResponse, error) {
	signed, err := c.CreateOrder(ctx, args)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, *signed, orderType, postOnly)
}

// CreateAndPostMarketOrder creates, signs, and submits a market order.
func (c *ClobClient) CreateAndPostMarketOrder(ctx context.Context, args MarketOrderArgs, postOnly bool) (*OrderResponse, error) {
	orderType := args.OrderType
	if orderType == "" {
		orderType = FOK
	}
	signed, err := c.CreateMarketOrder(ctx, args)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, *signed, orderType, postOnly)
}

// CancelOrder cancels a single order by ID.
func (c *ClobClient) CancelOrder(ctx context.Context, orderID string) error {
	reqBody := OrderPayload{OrderID: orderID}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling cancel request: %w", err)
	}

	headers, err := c.l2Headers("DELETE", EndpointCancelOrder, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointCancelOrder, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// CancelOrders cancels multiple orders by their IDs.
func (c *ClobClient) CancelOrders(ctx context.Context, orderIDs []string) error {
	reqBody := orderIDs

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling cancel request: %w", err)
	}

	headers, err := c.l2Headers("DELETE", EndpointCancelOrders, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointCancelOrders, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// CancelMarketOrders cancels orders by market and/or asset id.
func (c *ClobClient) CancelMarketOrders(ctx context.Context, market, assetID string) error {
	reqBody := OrderMarketCancelParams{
		Market:  market,
		AssetID: assetID,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling cancel market orders request: %w", err)
	}

	headers, err := c.l2Headers("DELETE", EndpointCancelMarketOrders, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointCancelMarketOrders, headers, reqBody)
	if err != nil {
		return err
	}
	_, err = transport.ParseResponse(resp)
	return err
}

// CancelAll cancels all open orders for the authenticated user.
func (c *ClobClient) CancelAll(ctx context.Context) error {
	headers, err := c.l2Headers("DELETE", EndpointCancelAll, "")
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointCancelAll, headers, nil)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// GetOrder returns a single order by ID. Requires L2 authentication.
func (c *ClobClient) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	path := EndpointOrder + orderID

	headers, err := c.l2Headers("GET", path, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, path, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := json.Unmarshal(raw, &order); err != nil {
		return nil, fmt.Errorf("polymarket: parsing order: %w", err)
	}
	return &order, nil
}

// GetOpenOrders returns an iterator over open orders with auto-pagination.
// Requires L2 authentication. The optional params filter by market or asset.
func (c *ClobClient) GetOpenOrders(ctx context.Context, params OpenOrderParams) iter.Seq2[Order, error] {
	return paginate[Order](ctx, func(cursor string) (PaginatedResponse[Order], error) {
		headers, err := c.l2Headers("GET", EndpointOrders, "")
		if err != nil {
			return PaginatedResponse[Order]{}, err
		}
		query := make(map[string]string)
		if params.Market != "" {
			query["market"] = params.Market
		}
		if params.AssetID != "" {
			query["asset_id"] = params.AssetID
		}
		if params.ID != "" {
			query["id"] = params.ID
		}
		if cursor != "" {
			query["next_cursor"] = cursor
		}
		resp, err := c.http.Get(ctx, EndpointOrders, headers, query)
		if err != nil {
			return PaginatedResponse[Order]{}, err
		}
		body, err := transport.ParseResponse(resp)
		if err != nil {
			return PaginatedResponse[Order]{}, err
		}
		var page PaginatedResponse[Order]
		if err := json.Unmarshal(body, &page); err != nil {
			return PaginatedResponse[Order]{}, fmt.Errorf("polymarket: parsing orders: %w", err)
		}
		return page, nil
	})
}

// CalculateMarketPrice computes a matching market price from the current order book.
func (c *ClobClient) CalculateMarketPrice(ctx context.Context, tokenID string, side Side, amount decimal.Decimal, orderType OrderType) (decimal.Decimal, error) {
	book, err := c.GetOrderBook(ctx, tokenID)
	if err != nil {
		return decimal.Zero, err
	}
	if book == nil {
		return decimal.Zero, fmt.Errorf("polymarket: no orderbook")
	}

	switch side {
	case Buy:
		if len(book.Asks) == 0 {
			return decimal.Zero, fmt.Errorf("polymarket: no match")
		}
		sum := decimal.Zero
		for i := len(book.Asks) - 1; i >= 0; i-- {
			price, err := decimal.NewFromString(book.Asks[i].Price)
			if err != nil {
				return decimal.Zero, fmt.Errorf("polymarket: invalid ask price: %w", err)
			}
			size, err := decimal.NewFromString(book.Asks[i].Size)
			if err != nil {
				return decimal.Zero, fmt.Errorf("polymarket: invalid ask size: %w", err)
			}
			sum = sum.Add(size.Mul(price))
			if sum.GreaterThanOrEqual(amount) {
				return price, nil
			}
		}
		if orderType == FOK {
			return decimal.Zero, fmt.Errorf("polymarket: no match")
		}
		return decimal.NewFromString(book.Asks[0].Price)

	case Sell:
		if len(book.Bids) == 0 {
			return decimal.Zero, fmt.Errorf("polymarket: no match")
		}
		sum := decimal.Zero
		for i := len(book.Bids) - 1; i >= 0; i-- {
			price, err := decimal.NewFromString(book.Bids[i].Price)
			if err != nil {
				return decimal.Zero, fmt.Errorf("polymarket: invalid bid price: %w", err)
			}
			size, err := decimal.NewFromString(book.Bids[i].Size)
			if err != nil {
				return decimal.Zero, fmt.Errorf("polymarket: invalid bid size: %w", err)
			}
			sum = sum.Add(size)
			if sum.GreaterThanOrEqual(amount) {
				return price, nil
			}
		}
		if orderType == FOK {
			return decimal.Zero, fmt.Errorf("polymarket: no match")
		}
		return decimal.NewFromString(book.Bids[0].Price)
	default:
		return decimal.Zero, &ValidationError{Field: "side", Message: "must be BUY or SELL"}
	}
}

func (c *ClobClient) resolveFeeRateBps(ctx context.Context, tokenID string, userFeeRate int) (int, error) {
	feeRateStr, err := c.GetFeeRateBps(ctx, tokenID)
	if err != nil {
		return 0, err
	}
	if feeRateStr == "" {
		return userFeeRate, nil
	}
	feeRate, err := strconv.Atoi(feeRateStr)
	if err != nil {
		return 0, fmt.Errorf("polymarket: invalid fee rate from server: %s", feeRateStr)
	}
	if feeRate > 0 && userFeeRate > 0 && userFeeRate != feeRate {
		return 0, &ValidationError{
			Field:   "fee_rate_bps",
			Message: fmt.Sprintf("invalid user provided fee rate (%d), fee rate for the market must be %d", userFeeRate, feeRate),
		}
	}
	if feeRate > 0 {
		return feeRate, nil
	}
	return userFeeRate, nil
}
