package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIError is defined in transport to avoid circular imports with the parent
// client package. It mirrors client.APIError.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("polymarket: %s %s returned %d: %s", e.Method, e.Path, e.StatusCode, e.Message)
}

// HTTPClient is a resilient HTTP client with retry logic, exponential backoff,
// and jitter for communicating with the Polymarket CLOB API.
type HTTPClient struct {
	client     *http.Client
	baseURL    string
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// Option is a functional option for configuring HTTPClient.
type Option func(*HTTPClient)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *HTTPClient) {
		c.client.Timeout = d
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) Option {
	return func(c *HTTPClient) {
		c.maxRetries = n
	}
}

// WithBaseDelay sets the base delay for exponential backoff.
func WithBaseDelay(d time.Duration) Option {
	return func(c *HTTPClient) {
		c.baseDelay = d
	}
}

// WithMaxDelay sets the maximum delay cap for exponential backoff.
func WithMaxDelay(d time.Duration) Option {
	return func(c *HTTPClient) {
		c.maxDelay = d
	}
}

// NewHTTPClient creates a new HTTPClient with the given base URL and options.
// Default configuration: timeout=10s, maxRetries=3, baseDelay=100ms, maxDelay=5s.
func NewHTTPClient(baseURL string, opts ...Option) *HTTPClient {
	c := &HTTPClient{
		client:     &http.Client{Timeout: 10 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   5 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Get performs an HTTP GET request.
func (c *HTTPClient) Get(ctx context.Context, path string, headers http.Header, query map[string]string) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, headers, nil)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	return c.do(req)
}

// Post performs an HTTP POST request. The body is marshalled to JSON.
func (c *HTTPClient) Post(ctx context.Context, path string, headers http.Header, body interface{}) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodPost, path, headers, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// Delete performs an HTTP DELETE request. The body is marshalled to JSON.
func (c *HTTPClient) Delete(ctx context.Context, path string, headers http.Header, body interface{}) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, path, headers, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// DoJSON executes the given request through the retry-aware client and
// unmarshals the JSON response body into a value of type T.
func DoJSON[T any](c *HTTPClient, req *http.Request) (T, error) {
	var zero T

	resp, err := c.do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("polymarket: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return zero, &APIError{
			StatusCode: resp.StatusCode,
			Method:     req.Method,
			Path:       req.URL.Path,
			Message:    msg,
		}
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return zero, fmt.Errorf("polymarket: unmarshalling response: %w", err)
	}
	return result, nil
}

// do executes the HTTP request with retry logic, exponential backoff, and jitter.
// It buffers the request body upfront so retries can replay it.
func (c *HTTPClient) do(req *http.Request) (*http.Response, error) {
	// Buffer the request body so we can replay it on retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("polymarket: reading request body: %w", err)
		}
		req.Body.Close()
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Check for context cancellation before each attempt.
		if err := req.Context().Err(); err != nil {
			return nil, err
		}

		// Clone the request body for this attempt.
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}

		resp, err := c.client.Do(req)
		if err != nil {
			if !isRetryableError(err) {
				return nil, err
			}
			lastErr = err
			if attempt < c.maxRetries {
				if waitErr := c.backoff(req.Context(), attempt, 0); waitErr != nil {
					return nil, waitErr
				}
			}
			continue
		}

		// Non-retryable status: return immediately.
		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// For retryable statuses, drain and close the body so the
		// connection can be reused.
		drainBody(resp)
		lastErr = &APIError{
			StatusCode: resp.StatusCode,
			Method:     req.Method,
			Path:       req.URL.Path,
			Message:    http.StatusText(resp.StatusCode),
		}

		if attempt < c.maxRetries {
			retryAfter := parseRetryAfter(resp)
			if waitErr := c.backoff(req.Context(), attempt, retryAfter); waitErr != nil {
				return nil, waitErr
			}
		}
	}

	return nil, fmt.Errorf("polymarket: request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// backoff sleeps for an exponentially increasing duration with jitter, capped
// at maxDelay. If retryAfterSec is positive (from a Retry-After header), that
// value is used instead.
func (c *HTTPClient) backoff(ctx context.Context, attempt int, retryAfterSec int) error {
	var delay time.Duration
	if retryAfterSec > 0 {
		delay = time.Duration(retryAfterSec) * time.Second
	} else {
		// Exponential backoff: baseDelay * 2^attempt.
		exp := math.Pow(2, float64(attempt))
		delay = time.Duration(float64(c.baseDelay) * exp)
		if delay > c.maxDelay {
			delay = c.maxDelay
		}
		// Apply jitter: multiply by a random factor in [0.75, 1.25].
		jitter := 0.75 + rand.Float64()*0.5
		delay = time.Duration(float64(delay) * jitter)
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// parseRetryAfter extracts the Retry-After header value (in seconds) from an
// HTTP response. Returns 0 if the header is absent or unparseable.
func parseRetryAfter(resp *http.Response) int {
	val := resp.Header.Get("Retry-After")
	if val == "" {
		return 0
	}
	secs, err := strconv.Atoi(val)
	if err != nil || secs < 0 {
		return 0
	}
	return secs
}

// drainBody reads and closes the response body so the underlying connection
// can be returned to the pool.
func drainBody(resp *http.Response) {
	if resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// isRetryableError reports whether a network-level error is transient and the
// request should be retried. It checks for connection refused, connection reset,
// timeout, and DNS resolution failures.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for timeout errors via the net.Error interface.
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Check for DNS errors. Only retry temporary ones; a "not found"
	// (NXDOMAIN) response is permanent and should not be retried.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return !dnsErr.IsNotFound
	}

	// Check for operational errors (connection refused, connection reset).
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Fallback: check the error string for common transient patterns.
	msg := err.Error()
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") {
		return true
	}

	return false
}

// isRetryableStatus reports whether an HTTP status code indicates a transient
// failure that should be retried. Returns true for 429 (Too Many Requests) and
// all 5xx server errors.
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// newRequest builds an *http.Request with the full URL, JSON body, and headers.
func (c *HTTPClient) newRequest(ctx context.Context, method, path string, headers http.Header, body interface{}) (*http.Request, error) {
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("polymarket: marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("polymarket: creating request: %w", err)
	}

	// Copy all provided headers into the request first, so caller values
	// take precedence over defaults.
	for key, vals := range headers {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}

	// Set Content-Type as a default only if the caller didn't already provide one.
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// ParseResponse reads the response body and checks for API errors.
// On success (2xx), it returns the raw body bytes.
// On error, it returns an *APIError with the status code, method, path, and
// message extracted from the response body.
func ParseResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("polymarket: reading response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}

	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	return nil, &APIError{
		StatusCode: resp.StatusCode,
		Method:     resp.Request.Method,
		Path:       resp.Request.URL.Path,
		Message:    msg,
	}
}
