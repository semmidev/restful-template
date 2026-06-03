// Package wideevent implements the "Canonical Log Line" / "Wide Event" pattern.
//
// Instead of emitting many scattered log statements per request, a single wide
// event is built up throughout the request lifecycle and emitted once at the end
// by the logger middleware. This gives every log line full context — HTTP metadata,
// authenticated identity, and domain-specific business fields — all in one place.
//
// Usage:
//
//	// In middleware (initialise):
//	ctx = wideevent.New(r.Context())
//
//	// In handlers / usecases (enrich):
//	wideevent.Add(ctx, "todo_id", id.String())
//	wideevent.Add(ctx, "todo_title", title)
//
//	// In middleware (emit, after handler returns):
//	wideevent.Emit(ctx, log, level, "request", extra...)
package wideevent

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type ctxKey struct{}

// WideEvent is a thread-safe map of key→value pairs accumulated throughout
// a single request's lifetime.
type WideEvent struct {
	mu     sync.Mutex
	fields []any // slog-style alternating key, value pairs
}

// New creates a fresh WideEvent and stores it in the returned context.
func New(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, &WideEvent{})
}

// Add enriches the wide event for this request with a key-value pair.
// Safe to call from concurrent goroutines.
// No-op when called outside of a request context (e.g. in unit tests that
// skip the middleware), so callers never need a nil check.
func Add(ctx context.Context, key string, value any) {
	ev, ok := ctx.Value(ctxKey{}).(*WideEvent)
	if !ok || ev == nil {
		return
	}
	ev.mu.Lock()
	ev.fields = append(ev.fields, key, value)
	ev.mu.Unlock()
}

func Fields(ctx context.Context) []any {
	ev, ok := ctx.Value(ctxKey{}).(*WideEvent)
	if !ok || ev == nil {
		return nil
	}
	ev.mu.Lock()
	out := make([]any, len(ev.fields))
	copy(out, ev.fields)
	ev.mu.Unlock()
	return out
}

// Emit logs the accumulated wide event at the given level, merging any
// extra key-value pairs (e.g. status, duration_ms) provided by the caller.
// extra must be in alternating key, value order, just like slog args.
func Emit(ctx context.Context, log *slog.Logger, level slog.Level, msg string, extra ...any) {
	if log == nil || !log.Enabled(ctx, level) {
		return
	}

	// Capture the program counter of Emit's caller to ensure the "source" field
	// (file/line) points to the middleware/handler, not this helper function.
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])

	accumulated := Fields(ctx)
	// extra fields (status, outcome, duration) come first so they appear at
	// the top of the JSON object for readability in most log viewers.
	args := make([]any, 0, len(extra)+len(accumulated))
	args = append(args, extra...)
	args = append(args, accumulated...)

	r.Add(args...)
	_ = log.Handler().Handle(ctx, r)
}
