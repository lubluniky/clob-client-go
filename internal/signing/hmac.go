package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// BuildHMACSignature creates an HMAC-SHA256 signature for L2 API authentication.
//
// Parameters:
//   - secret: base64 URL-safe encoded API secret
//   - timestamp: unix timestamp as string
//   - method: HTTP method or signing method string
//   - requestPath: API endpoint path (e.g., "/orders")
//   - body: optional request body string (empty string if no body)
//
// The message to sign is the concatenation: timestamp + method + requestPath + body.
//
// Returns the base64 URL-safe encoded signature.
func BuildHMACSignature(secret, timestamp, method, requestPath, body string) (string, error) {
	// 1. Decode secret from base64 URL-safe encoding.
	decodedSecret, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("signing: failed to decode secret: %w", err)
	}

	// 2. Build message: timestamp + method + requestPath + body
	message := timestamp + method + requestPath + body

	// 3. Create HMAC-SHA256 digest.
	mac := hmac.New(sha256.New, decodedSecret)
	mac.Write([]byte(message))

	// 4. Return base64 URL-safe encoded signature.
	return base64.URLEncoding.EncodeToString(mac.Sum(nil)), nil
}
