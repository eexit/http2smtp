package ctx

import (
	"context"
	"testing"
)

func Test(t *testing.T) {
	tests := []struct {
		name string
		gc   func() context.Context
		want string
	}{
		{
			name: "context has not trace ID",
			gc:   func() context.Context { return context.Background() },
			want: "",
		},
		{
			name: "context has a trace ID",
			gc:   func() context.Context { return WithTraceID(context.Background(), "trace") },
			want: "trace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TraceID(tt.gc()); got != tt.want {
				t.Errorf("TraceID() = %v, want %v", got, tt.want)
			}
		})
	}
}
