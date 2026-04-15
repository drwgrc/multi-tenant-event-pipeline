package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const RequestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func NewCorrelationID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}

	return hex.EncodeToString(buffer)
}

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	if !ok || requestID == "" {
		return "", false
	}

	return requestID, true
}
