package client

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

func testSigner(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func testCreds() ApiCreds {
	return ApiCreds{
		ApiKey:        "test-api-key",
		ApiSecret:     base64.URLEncoding.EncodeToString([]byte("super-secret-key-material")),
		ApiPassphrase: "test-passphrase",
	}
}

func sampleSignedOrder() SignedOrder {
	return SignedOrder{
		Salt:          "1",
		Maker:         "0x0000000000000000000000000000000000000001",
		Signer:        "0x0000000000000000000000000000000000000001",
		Taker:         ZeroAddress,
		TokenID:       "1",
		MakerAmount:   "1000000",
		TakerAmount:   "2000000",
		Expiration:    "0",
		Nonce:         "1",
		FeeRateBps:    "0",
		Side:          Buy,
		SignatureType: EOA,
		Signature:     "0xdeadbeef",
	}
}

func TestPostOrderPayloadAndValidation(t *testing.T) {
	var requestBody PostOrderRequest
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != EndpointPostOrder || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		callCount++
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &requestBody); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"ord1","status":"ok"}`))
	}))
	defer srv.Close()

	key := testSigner(t)
	creds := testCreds()
	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))

	_, err := client.PostOrder(context.Background(), sampleSignedOrder(), FOK, true)
	if err == nil || !strings.Contains(err.Error(), "postOnly is only supported") {
		t.Fatalf("expected postOnly validation error, got: %v", err)
	}
	if callCount != 0 {
		t.Fatalf("expected no request on validation error, got %d", callCount)
	}

	_, err = client.PostOrder(context.Background(), sampleSignedOrder(), GTC, true)
	if err != nil {
		t.Fatalf("post order: %v", err)
	}
	if requestBody.Owner != creds.ApiKey {
		t.Fatalf("owner mismatch: got %q want %q", requestBody.Owner, creds.ApiKey)
	}
}

func TestPostOrdersPayload(t *testing.T) {
	var payload []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != EndpointPostOrders || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"o1","status":"ok"}]`))
	}))
	defer srv.Close()

	key := testSigner(t)
	creds := testCreds()
	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))

	postOnly := true
	_, err := client.PostOrders(context.Background(), []PostOrdersArgs{
		{
			Order:     sampleSignedOrder(),
			OrderType: GTD,
			PostOnly:  &postOnly,
		},
	}, false, false)
	if err != nil {
		t.Fatalf("post orders: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("unexpected payload len: %d", len(payload))
	}
	if got := payload[0]["owner"]; got != creds.ApiKey {
		t.Fatalf("owner mismatch: got %v want %s", got, creds.ApiKey)
	}
	if got := payload[0]["postOnly"]; got != true {
		t.Fatalf("postOnly mismatch: got %v", got)
	}
}

func TestCancelPayloads(t *testing.T) {
	var cancelOrder map[string]any
	var cancelOrders []string
	var cancelMarket map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case EndpointCancelOrder:
			if r.Method != http.MethodDelete {
				t.Fatalf("cancel order method: %s", r.Method)
			}
			_ = json.Unmarshal(body, &cancelOrder)
		case EndpointCancelOrders:
			if r.Method != http.MethodDelete {
				t.Fatalf("cancel orders method: %s", r.Method)
			}
			_ = json.Unmarshal(body, &cancelOrders)
		case EndpointCancelMarketOrders:
			if r.Method != http.MethodDelete {
				t.Fatalf("cancel market method: %s", r.Method)
			}
			_ = json.Unmarshal(body, &cancelMarket)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	key := testSigner(t)
	creds := testCreds()
	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))

	if err := client.CancelOrder(context.Background(), "ord-1"); err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if got := cancelOrder["orderID"]; got != "ord-1" {
		t.Fatalf("cancel order payload mismatch: %v", cancelOrder)
	}

	if err := client.CancelOrders(context.Background(), []string{"a", "b"}); err != nil {
		t.Fatalf("cancel orders: %v", err)
	}
	if len(cancelOrders) != 2 || cancelOrders[0] != "a" || cancelOrders[1] != "b" {
		t.Fatalf("cancel orders payload mismatch: %#v", cancelOrders)
	}

	if err := client.CancelMarketOrders(context.Background(), "mkt", "asset"); err != nil {
		t.Fatalf("cancel market orders: %v", err)
	}
	if cancelMarket["market"] != "mkt" || cancelMarket["asset_id"] != "asset" {
		t.Fatalf("cancel market payload mismatch: %#v", cancelMarket)
	}
}

func TestAccountPayloads(t *testing.T) {
	var notificationsQuery string
	var notificationsBody string
	var heartbeatBodies []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EndpointNotifications:
			if r.Method != http.MethodDelete {
				t.Fatalf("notifications method: %s", r.Method)
			}
			notificationsQuery = r.URL.RawQuery
			body, _ := io.ReadAll(r.Body)
			notificationsBody = string(body)
		case EndpointHeartbeat:
			if r.Method != http.MethodPost {
				t.Fatalf("heartbeat method: %s", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			heartbeatBodies = append(heartbeatBodies, string(body))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	key := testSigner(t)
	creds := testCreds()
	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))

	if err := client.DropNotifications(context.Background(), []string{"1", "2"}); err != nil {
		t.Fatalf("drop notifications: %v", err)
	}
	if notificationsQuery != "ids=1,2" {
		t.Fatalf("notifications query mismatch: %s", notificationsQuery)
	}
	if notificationsBody != "" {
		t.Fatalf("expected empty notifications body, got %q", notificationsBody)
	}

	if err := client.PostHeartbeat(context.Background(), "hb-1"); err != nil {
		t.Fatalf("post heartbeat: %v", err)
	}
	if err := client.PostHeartbeat(context.Background(), ""); err != nil {
		t.Fatalf("post heartbeat empty: %v", err)
	}
	if len(heartbeatBodies) != 2 {
		t.Fatalf("heartbeat calls mismatch: %d", len(heartbeatBodies))
	}
	if heartbeatBodies[0] != `{"heartbeat_id":"hb-1"}` {
		t.Fatalf("heartbeat body mismatch: %s", heartbeatBodies[0])
	}
	if heartbeatBodies[1] != `{"heartbeat_id":null}` {
		t.Fatalf("heartbeat null body mismatch: %s", heartbeatBodies[1])
	}
}

func TestBatchMarketDataPayloads(t *testing.T) {
	var bodies = map[string]string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies[r.URL.Path] = string(body)
		switch r.URL.Path {
		case EndpointMidpoints:
			_, _ = w.Write([]byte(`{"1":"0.5"}`))
		case EndpointPrices:
			_, _ = w.Write([]byte(`{"1":"0.6"}`))
		case EndpointSpreads:
			_, _ = w.Write([]byte(`{"1":{"spread":"0.1","bid":"0.45","ask":"0.55"}}`))
		case EndpointLastTradesPrices:
			_, _ = w.Write([]byte(`{"1":"0.51"}`))
		case EndpointOrderBooks:
			_, _ = w.Write([]byte(`[{"market":"m","asset_id":"1","timestamp":"t","bids":[],"asks":[],"min_order_size":"1","neg_risk":false,"tick_size":"0.01","last_trade_price":"0.5","hash":""}]`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClobClient(WithBaseURL(srv.URL))
	if _, err := client.GetMidpoints(context.Background(), []string{"1"}); err != nil {
		t.Fatalf("get midpoints: %v", err)
	}
	if _, err := client.GetPrices(context.Background(), []string{"1"}, Buy); err != nil {
		t.Fatalf("get prices: %v", err)
	}
	if _, err := client.GetSpreads(context.Background(), []string{"1"}); err != nil {
		t.Fatalf("get spreads: %v", err)
	}
	if _, err := client.GetLastTradesPrices(context.Background(), []string{"1"}); err != nil {
		t.Fatalf("get last trades prices: %v", err)
	}
	if _, err := client.GetOrderBooks(context.Background(), []BookParams{{TokenID: "1"}}); err != nil {
		t.Fatalf("get order books: %v", err)
	}

	if bodies[EndpointMidpoints] != `[{"token_id":"1"}]` {
		t.Fatalf("midpoints payload mismatch: %s", bodies[EndpointMidpoints])
	}
	if bodies[EndpointPrices] != `[{"token_id":"1","side":"BUY"}]` {
		t.Fatalf("prices payload mismatch: %s", bodies[EndpointPrices])
	}
	if bodies[EndpointSpreads] != `[{"token_id":"1"}]` {
		t.Fatalf("spreads payload mismatch: %s", bodies[EndpointSpreads])
	}
	if bodies[EndpointLastTradesPrices] != `[{"token_id":"1"}]` {
		t.Fatalf("last trades payload mismatch: %s", bodies[EndpointLastTradesPrices])
	}
	if bodies[EndpointOrderBooks] != `[{"token_id":"1"}]` {
		t.Fatalf("order books payload mismatch: %s", bodies[EndpointOrderBooks])
	}
}

func TestGetApiKeysParsing(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantLen int
		wantKey string
		wantErr bool
	}{
		{
			name:    "wrapped",
			body:    `{"apiKeys":[{"apiKey":"k1","secret":"s1","passphrase":"p1"}]}`,
			wantLen: 1,
			wantKey: "k1",
		},
		{
			name:    "array",
			body:    `[{"apiKey":"k2"}]`,
			wantLen: 1,
			wantKey: "k2",
		},
		{
			name:    "single",
			body:    `{"apiKey":"k3"}`,
			wantLen: 1,
			wantKey: "k3",
		},
		{
			name:    "invalid",
			body:    `{}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != EndpointGetApiKeys {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			key := testSigner(t)
			creds := testCreds()
			client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))
			keys, err := client.GetApiKeys(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("get api keys: %v", err)
			}
			if len(keys) != tc.wantLen {
				t.Fatalf("len mismatch: got %d want %d", len(keys), tc.wantLen)
			}
			if keys[0].ApiKey != tc.wantKey {
				t.Fatalf("api key mismatch: got %s want %s", keys[0].ApiKey, tc.wantKey)
			}
		})
	}
}

func TestL2AuthRequiresAddress(t *testing.T) {
	client := NewClobClient(WithCreds(testCreds()))
	_, err := client.GetApiKeys(context.Background())
	if err == nil {
		t.Fatalf("expected auth error")
	}
	if !strings.Contains(err.Error(), "address required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateMarketOrderUsesMarketPrice(t *testing.T) {
	key := testSigner(t)
	creds := testCreds()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EndpointTickSize:
			_, _ = w.Write([]byte(`{"minimum_tick_size":0.01}`))
		case EndpointNegRisk:
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case EndpointFeeRate:
			_, _ = w.Write([]byte(`{"base_fee":0}`))
		case EndpointOrderBook:
			_, _ = w.Write([]byte(`{"market":"m","asset_id":"1","timestamp":"t","bids":[{"price":"0.4","size":"100"}],"asks":[{"price":"0.5","size":"100"},{"price":"0.6","size":"100"}],"min_order_size":"1","neg_risk":false,"tick_size":"0.01","last_trade_price":"0.5","hash":""}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))
	order, err := client.CreateMarketOrder(context.Background(), MarketOrderArgs{
		TokenID:   "1",
		Amount:    decimal.RequireFromString("50"),
		Side:      Buy,
		OrderType: FAK,
	})
	if err != nil {
		t.Fatalf("create market order: %v", err)
	}
	if order.MakerAmount != "50000000" {
		t.Fatalf("maker amount mismatch: %s", order.MakerAmount)
	}
	if order.TakerAmount != "83333300" {
		t.Fatalf("taker amount mismatch: %s", order.TakerAmount)
	}
}

func TestCreateOrderValidationAndFeeMismatch(t *testing.T) {
	key := testSigner(t)
	creds := testCreds()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EndpointTickSize:
			_, _ = w.Write([]byte(`{"minimum_tick_size":0.01}`))
		case EndpointNegRisk:
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case EndpointFeeRate:
			_, _ = w.Write([]byte(`{"base_fee":10}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(creds))

	_, err := client.CreateOrder(context.Background(), OrderArgs{
		TokenID: "1",
		Price:   decimal.RequireFromString("0.999"),
		Size:    decimal.RequireFromString("1"),
		Side:    Buy,
	})
	if err == nil || !strings.Contains(err.Error(), "price") {
		t.Fatalf("expected price validation error, got: %v", err)
	}

	_, err = client.CreateOrder(context.Background(), OrderArgs{
		TokenID:    "1",
		Price:      decimal.RequireFromString("0.5"),
		Size:       decimal.RequireFromString("1"),
		Side:       Buy,
		FeeRateBps: 20,
	})
	if err == nil || !strings.Contains(err.Error(), "fee rate") {
		t.Fatalf("expected fee-rate mismatch error, got: %v", err)
	}
}
