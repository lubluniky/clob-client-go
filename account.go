package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lubluniky/clob-client-go/internal/transport"
)

// GetBalanceAllowance returns the balance and allowance for a given asset.
// Requires L2 authentication.
func (c *ClobClient) GetBalanceAllowance(ctx context.Context, params BalanceAllowanceParams) (*BalanceAllowance, error) {
	query := map[string]string{
		"asset_type":     params.AssetType,
		"token_id":       params.TokenID,
		"signature_type": strconv.Itoa(int(params.SignatureType)),
	}

	headers, err := c.l2Headers("GET", EndpointBalanceAllowance, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointBalanceAllowance, headers, query)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var result BalanceAllowance
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("polymarket: parsing balance allowance: %w", err)
	}
	return &result, nil
}

// GetNotifications returns the authenticated user's notifications.
// Requires L2 authentication. The signatureType should match the client's
// signature type (0 for EOA, 1 for PolyProxy, 2 for PolyGnosisSafe).
func (c *ClobClient) GetNotifications(ctx context.Context, signatureType SignatureType) ([]Notification, error) {
	query := map[string]string{
		"signature_type": strconv.Itoa(int(signatureType)),
	}

	headers, err := c.l2Headers("GET", EndpointNotifications, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Get(ctx, EndpointNotifications, headers, query)
	if err != nil {
		return nil, err
	}

	raw, err := transport.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	if err := json.Unmarshal(raw, &notifications); err != nil {
		return nil, fmt.Errorf("polymarket: parsing notifications: %w", err)
	}
	return notifications, nil
}

// DropNotifications deletes notifications by their IDs.
// Requires L2 authentication.
func (c *ClobClient) DropNotifications(ctx context.Context, ids []string) error {
	reqBody := map[string]interface{}{"ids": ids}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling drop notifications request: %w", err)
	}

	headers, err := c.l2Headers("DELETE", EndpointNotifications, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Delete(ctx, EndpointNotifications, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}

// PostHeartbeat sends a heartbeat signal to keep an active session alive.
// Requires L2 authentication.
func (c *ClobClient) PostHeartbeat(ctx context.Context, heartbeatID string) error {
	reqBody := map[string]interface{}{"id": heartbeatID}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("polymarket: marshalling heartbeat request: %w", err)
	}

	headers, err := c.l2Headers("POST", EndpointHeartbeat, string(bodyBytes))
	if err != nil {
		return err
	}

	resp, err := c.http.Post(ctx, EndpointHeartbeat, headers, reqBody)
	if err != nil {
		return err
	}

	_, err = transport.ParseResponse(resp)
	return err
}
