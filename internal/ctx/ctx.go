package ctx

import "context"

type key int

const traceIDKey key = 0

// TraceID returns the traceID from the given context if any,
// it returns an empty string if the context has no value
func TraceID(gc context.Context) string {
	if traceID, ok := gc.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithTraceID returns a context with a TraceID
func WithTraceID(gc context.Context, traceID string) context.Context {
	return context.WithValue(gc, traceIDKey, traceID)
}
