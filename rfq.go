package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// CreateRfqRequest creates a new RFQ (Request For Quote) request.
// Requires L2 authentication.
func (c *ClobClient) CreateRfqRequest(ctx context.Context, params CreateRfqRequestParams) (*RfqRequest, error) {
	bodyBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("polymarket: marshalling rfq request: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointRfqRequest, string(bodyBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointRfqRequest, headers, params)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result RfqRequest
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing rfq request response: %w", err)
	}
	return &result, nil
}

// CancelRfqRequest cancels an existing RFQ request by its ID.
// Requires L2 authentication.
func (c *ClobClient) CancelRfqRequest(ctx context.Context, requestID string) error {
	reqBody := map[string]interface{}{"requestId": requestID}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling cancel rfq request: %w", err)
	}

	headers, err := c.l2Headers("DELETE", EndpointRfqRequest, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointRfqRequest, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// GetRfqRequests returns all RFQ requests for the authenticated user.
// Requires L2 authentication.
func (c *ClobClient) GetRfqRequests(ctx context.Context) ([]RfqRequest, error) {
	headers, err := c.l2Headers("GET", EndpointRfqRequests, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRfqRequests, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var requests []RfqRequest
	if err := json.Unmarshal(raw, &requests); err != nil {
		return nil, fmt.Errorf("polymarket: parsing rfq requests: %w", err)
	}
	return requests, nil
}

// CreateRfqQuote creates a new quote in response to an RFQ request.
// Requires L2 authentication.
func (c *ClobClient) CreateRfqQuote(ctx context.Context, params CreateRfqQuoteParams) (*RfqQuote, error) {
	bodyBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("polymarket: marshalling rfq quote: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointRfqQuote, string(bodyBytes))
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Post(ctx, EndpointRfqQuote, headers, params)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result RfqQuote
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing rfq quote response: %w", err)
	}
	return &result, nil
}

// CancelRfqQuote cancels an existing RFQ quote by its ID.
// Requires L2 authentication.
func (c *ClobClient) CancelRfqQuote(ctx context.Context, quoteID string) error {
	reqBody := map[string]interface{}{"quoteId": quoteID}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling cancel rfq quote: %w", err)
	}

	headers, err := c.l2Headers("DELETE", EndpointRfqQuote, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointRfqQuote, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// GetRfqRequesterQuotes returns all quotes for RFQ requests made by the
// authenticated user (i.e., quotes received as a requester).
// Requires L2 authentication.
func (c *ClobClient) GetRfqRequesterQuotes(ctx context.Context) ([]RfqQuote, error) {
	headers, err := c.l2Headers("GET", EndpointRfqRequesterQuotes, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRfqRequesterQuotes, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var quotes []RfqQuote
	if err := json.Unmarshal(raw, &quotes); err != nil {
		return nil, fmt.Errorf("polymarket: parsing requester quotes: %w", err)
	}
	return quotes, nil
}

// GetRfqQuoterQuotes returns all quotes submitted by the authenticated user
// (i.e., quotes created as a quoter).
// Requires L2 authentication.
func (c *ClobClient) GetRfqQuoterQuotes(ctx context.Context) ([]RfqQuote, error) {
	headers, err := c.l2Headers("GET", EndpointRfqQuoterQuotes, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRfqQuoterQuotes, headers, nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var quotes []RfqQuote
	if err := json.Unmarshal(raw, &quotes); err != nil {
		return nil, fmt.Errorf("polymarket: parsing quoter quotes: %w", err)
	}
	return quotes, nil
}

// GetRfqBestQuote returns the best available quote for an RFQ request.
// Requires L2 authentication.
func (c *ClobClient) GetRfqBestQuote(ctx context.Context, requestID string) (*RfqQuote, error) {
	query := map[string]string{"requestId": requestID}

	headers, err := c.l2Headers("GET", EndpointRfqBestQuote, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointRfqBestQuote, headers, query)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var quote RfqQuote
	if err := json.Unmarshal(raw, &quote); err != nil {
		return nil, fmt.Errorf("polymarket: parsing best rfq quote: %w", err)
	}
	return &quote, nil
}

// AcceptRfqRequest accepts an RFQ request, signaling willingness to trade.
// Requires L2 authentication.
func (c *ClobClient) AcceptRfqRequest(ctx context.Context, requestID string) error {
	reqBody := map[string]interface{}{"requestId": requestID}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling accept rfq request: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointRfqRequestAccept, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Post(ctx, EndpointRfqRequestAccept, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// ApproveRfqQuote approves an RFQ quote, finalizing the trade agreement.
// Requires L2 authentication.
func (c *ClobClient) ApproveRfqQuote(ctx context.Context, quoteID string) error {
	reqBody := map[string]interface{}{"quoteId": quoteID}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling approve rfq quote: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointRfqQuoteApprove, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Post(ctx, EndpointRfqQuoteApprove, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// GetRfqConfig returns the current RFQ feature configuration.
// This is a public (L0) endpoint.
func (c *ClobClient) GetRfqConfig(ctx context.Context) (*RfqConfig, error) {
	resp, err := c.http.Get(ctx, EndpointRfqConfig, c.l0Headers(), nil)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var config RfqConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("polymarket: parsing rfq config: %w", err)
	}
	return &config, nil
}
