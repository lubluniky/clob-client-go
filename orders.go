package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

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
	tickSize, err := c.GetTickSize(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting tick size: %w", err)
	}

	negRisk, err := c.GetNegRisk(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting neg risk: %w", err)
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
		FeeRateBps:    fmt.Sprintf("%d", args.FeeRateBps),
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
	tickSize, err := c.GetTickSize(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting tick size: %w", err)
	}

	negRisk, err := c.GetNegRisk(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("polymarket: getting neg risk: %w", err)
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
		FeeRateBps:    fmt.Sprintf("%d", args.FeeRateBps),
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
	req := PostOrderRequest{
		Order:     order,
		Owner:     c.creds.ApiKey,
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

// CreateAndPostOrder is a convenience method that creates, signs, and posts a
// limit order in a single call.
func (c *ClobClient) CreateAndPostOrder(ctx context.Context, args OrderArgs, orderType OrderType, postOnly bool) (*OrderResponse, error) {
	signed, err := c.CreateOrder(ctx, args)
	if err != nil {
		return nil, err
	}
	return c.PostOrder(ctx, *signed, orderType, postOnly)
}

// CancelOrder cancels a single order by ID.
func (c *ClobClient) CancelOrder(ctx context.Context, orderID string) error {
	reqBody := map[string]interface{}{"id": orderID}

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
	reqBody := map[string]interface{}{"ids": orderIDs}

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

// CancelAll cancels all open orders for the authenticated user.
func (c *ClobClient) CancelAll(ctx context.Context) error {
	headers, err := c.l2Headers("POST", EndpointCancelAll, "")
	if err != nil {
		return err
	}

	resp, err := c.http.Post(ctx, EndpointCancelAll, headers, nil)
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
	headers, err := c.l2Headers("GET", EndpointOrders, "")
	if err != nil {
		return func(yield func(Order, error) bool) {
			yield(Order{}, err)
		}
	}

	return paginate[Order](ctx, func(cursor string) (PaginatedResponse[Order], error) {
		query := make(map[string]string)
		if params.Market != "" {
			query["market"] = params.Market
		}
		if params.AssetID != "" {
			query["asset_id"] = params.AssetID
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
