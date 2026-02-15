package client

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

func TestPostOrderResponseOrderIDNormalization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != EndpointPostOrder || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"orderID":"ord-ts","status":"ok"}`))
	}))
	defer srv.Close()

	key := testSigner(t)
	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(testCreds()))
	resp, err := client.PostOrder(context.Background(), sampleSignedOrder(), GTC, false)
	if err != nil {
		t.Fatalf("post order: %v", err)
	}
	if resp.ID != "ord-ts" || resp.OrderID != "ord-ts" {
		t.Fatalf("id normalization mismatch: %+v", resp)
	}
}

func TestBalanceAllowanceQueries(t *testing.T) {
	var gotBalance url.Values
	var gotUpdate url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EndpointBalanceAllowance:
			gotBalance = r.URL.Query()
			_, _ = w.Write([]byte(`{"balance":"1","allowance":"2"}`))
		case EndpointUpdateBalanceAllowance:
			gotUpdate = r.URL.Query()
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	key := testSigner(t)
	client := NewClobClient(
		WithBaseURL(srv.URL),
		WithSigner(key),
		WithCreds(testCreds()),
		WithSignatureType(PolyProxy),
	)

	_, err := client.GetBalanceAllowance(context.Background(), BalanceAllowanceParams{
		SignatureType: SignatureUnset,
	})
	if err != nil {
		t.Fatalf("get balance allowance: %v", err)
	}
	if got := gotBalance.Get("signature_type"); got != "1" {
		t.Fatalf("signature_type mismatch: got %q", got)
	}
	if gotBalance.Has("asset_type") || gotBalance.Has("token_id") {
		t.Fatalf("unexpected optional query fields: %v", gotBalance.Encode())
	}

	if err := client.UpdateBalanceAllowance(context.Background(), BalanceAllowanceParams{
		AssetType:     string(AssetTypeCollateral),
		TokenID:       "123",
		SignatureType: SignatureUnset,
	}); err != nil {
		t.Fatalf("update balance allowance: %v", err)
	}

	if got := gotUpdate.Get("asset_type"); got != string(AssetTypeCollateral) {
		t.Fatalf("asset_type mismatch: got %q", got)
	}
	if got := gotUpdate.Get("token_id"); got != "123" {
		t.Fatalf("token_id mismatch: got %q", got)
	}
	if got := gotUpdate.Get("signature_type"); got != "1" {
		t.Fatalf("signature_type mismatch: got %q", got)
	}
}

func TestCreateOrderUsesClientSignatureTypeAndFunder(t *testing.T) {
	key := testSigner(t)
	funder := common.HexToAddress("0x00000000000000000000000000000000000000F0")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EndpointTickSize:
			_, _ = w.Write([]byte(`{"minimum_tick_size":0.01}`))
		case EndpointNegRisk:
			_, _ = w.Write([]byte(`{"neg_risk":false}`))
		case EndpointFeeRate:
			_, _ = w.Write([]byte(`{"base_fee":0}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClobClient(
		WithBaseURL(srv.URL),
		WithSigner(key),
		WithCreds(testCreds()),
		WithSignatureType(PolyProxy),
		WithFunderAddress(funder.Hex()),
	)

	order, err := client.CreateOrder(context.Background(), OrderArgs{
		TokenID: "1",
		Price:   decimal.RequireFromString("0.5"),
		Size:    decimal.RequireFromString("10"),
		Side:    Buy,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	if order.SignatureType != PolyProxy {
		t.Fatalf("signature type mismatch: got %d", order.SignatureType)
	}
	if order.Maker != funder.Hex() {
		t.Fatalf("maker/funder mismatch: got %s want %s", order.Maker, funder.Hex())
	}
	if order.Signer != crypto.PubkeyToAddress(key.PublicKey).Hex() {
		t.Fatalf("signer mismatch: got %s", order.Signer)
	}
}

func TestTickSizeCacheTTLAndClear(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != EndpointTickSize {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		call++
		if call == 1 {
			_, _ = w.Write([]byte(`{"minimum_tick_size":0.01}`))
			return
		}
		_, _ = w.Write([]byte(`{"minimum_tick_size":0.001}`))
	}))
	defer srv.Close()

	client := NewClobClient(WithBaseURL(srv.URL), WithTickSizeTTL(25*time.Millisecond))

	ts1, err := client.GetTickSize(context.Background(), "1")
	if err != nil {
		t.Fatalf("tick size first call: %v", err)
	}
	ts2, err := client.GetTickSize(context.Background(), "1")
	if err != nil {
		t.Fatalf("tick size second call: %v", err)
	}
	if ts1 != "0.01" || ts2 != "0.01" {
		t.Fatalf("expected cached tick size 0.01, got %q and %q", ts1, ts2)
	}
	if call != 1 {
		t.Fatalf("expected one tick-size API call, got %d", call)
	}

	time.Sleep(35 * time.Millisecond)
	ts3, err := client.GetTickSize(context.Background(), "1")
	if err != nil {
		t.Fatalf("tick size after ttl: %v", err)
	}
	if ts3 != "0.001" {
		t.Fatalf("ttl refresh mismatch: got %q", ts3)
	}
	if call != 2 {
		t.Fatalf("expected two tick-size API calls after ttl, got %d", call)
	}

	client.ClearTickSizeCache("1")
	if _, err := client.GetTickSize(context.Background(), "1"); err != nil {
		t.Fatalf("tick size after clear cache: %v", err)
	}
	if call != 3 {
		t.Fatalf("expected cache clear to force another call, got %d", call)
	}
}

func TestGetOrderBookHashDeterministic(t *testing.T) {
	client := NewClobClient()
	ob := &OrderBookSummary{
		Market:         "m",
		AssetID:        "1",
		Timestamp:      "t",
		Bids:           []PriceLevel{{Price: "0.45", Size: "10"}},
		Asks:           []PriceLevel{{Price: "0.55", Size: "12"}},
		MinOrderSize:   "1",
		TickSize:       "0.01",
		NegRisk:        false,
		LastTradePrice: "0.5",
	}

	hash, err := client.GetOrderBookHash(ob)
	if err != nil {
		t.Fatalf("get orderbook hash: %v", err)
	}
	const want = "08334c4690df6b480e5966509b8042cde57fdadd"
	if hash != want {
		t.Fatalf("hash mismatch: got %s want %s", hash, want)
	}
	if ob.Hash != want {
		t.Fatalf("orderbook hash field mismatch: got %s want %s", ob.Hash, want)
	}
}

func TestBuilderEndpointsContracts(t *testing.T) {
	var seen []string
	tradesCall := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == EndpointCreateBuilderApiKey:
			_, _ = w.Write([]byte(`{"apiKey":"bk","secret":"bs","passphrase":"bp"}`))
		case r.Method == http.MethodGet && r.URL.Path == EndpointGetBuilderApiKeys:
			_, _ = w.Write([]byte(`[{"apiKey":"bk"}]`))
		case r.Method == http.MethodDelete && r.URL.Path == EndpointRevokeBuilderApiKey:
			_, _ = w.Write([]byte(`{"ok":true}`))
		case r.Method == http.MethodGet && r.URL.Path == EndpointBuilderTrades:
			tradesCall++
			if tradesCall == 1 {
				_, _ = w.Write([]byte(`{"data":[{"id":"t1","market":"m","asset_id":"a","side":"BUY","size":"1","fee_rate_bps":"0","price":"0.5","status":"MATCHED","match_time":"mt","last_update":"lu","outcome":"YES","bucket_index":0,"owner":"o","maker_address":"ma","maker_orders":[],"transaction_hash":"tx","trader_side":"TAKER"}],"next_cursor":"NQ=="}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"t2","market":"m","asset_id":"a","side":"SELL","size":"2","fee_rate_bps":"0","price":"0.6","status":"MATCHED","match_time":"mt","last_update":"lu","outcome":"NO","bucket_index":0,"owner":"o","maker_address":"ma","maker_orders":[],"transaction_hash":"tx","trader_side":"TAKER"}],"next_cursor":"LTE="}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	key := testSigner(t)
	client := NewClobClient(WithBaseURL(srv.URL), WithSigner(key), WithCreds(testCreds()))

	created, err := client.CreateBuilderApiKey(context.Background())
	if err != nil || created.ApiKey != "bk" {
		t.Fatalf("create builder api key: %+v err=%v", created, err)
	}

	keys, err := client.GetBuilderApiKeys(context.Background())
	if err != nil || len(keys) != 1 || keys[0].ApiKey != "bk" {
		t.Fatalf("get builder api keys: %+v err=%v", keys, err)
	}

	if err := client.RevokeBuilderApiKey(context.Background()); err != nil {
		t.Fatalf("revoke builder api key: %v", err)
	}

	var gotIDs []string
	for trade, err := range client.GetBuilderTrades(context.Background(), TradeParams{}) {
		if err != nil {
			t.Fatalf("builder trades: %v", err)
		}
		gotIDs = append(gotIDs, trade.ID)
	}
	if strings.Join(gotIDs, ",") != "t1,t2" {
		t.Fatalf("unexpected builder trades: %v", gotIDs)
	}
	if tradesCall != 2 {
		t.Fatalf("expected paginated builder trades calls, got %d", tradesCall)
	}

	expected := []string{
		"POST " + EndpointCreateBuilderApiKey,
		"GET " + EndpointGetBuilderApiKeys,
		"DELETE " + EndpointRevokeBuilderApiKey,
		"GET " + EndpointBuilderTrades,
		"GET " + EndpointBuilderTrades,
	}
	if strings.Join(seen, "|") != strings.Join(expected, "|") {
		t.Fatalf("request sequence mismatch:\n got: %v\nwant: %v", seen, expected)
	}
}
