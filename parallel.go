package slogging

import (
	"context"
	"log/slog"
)

type parallelHandler struct {
	handlers []slog.Handler
}

// NewParallelHandler creates a handler that dispatches events to all of the
// provided handlers (for example, if you want logs to be written to both a file
// and stdout).
//
// The resulting handler is Enabled when at least one child handler is Enabled.
func NewParallelHandler(handlers ...slog.Handler) slog.Handler {
	return parallelHandler{
		handlers: handlers,
	}
}

func (ph parallelHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range ph.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (ph parallelHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range ph.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (ph parallelHandler) WithAttrs(a []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(ph.handlers))
	for i, h := range ph.handlers {
		newHandlers[i] = h.WithAttrs(a)
	}
	return NewParallelHandler(newHandlers...)
}

func (ph parallelHandler) WithGroup(g string) slog.Handler {
	newHandlers := make([]slog.Handler, len(ph.handlers))
	for i, h := range ph.handlers {
		newHandlers[i] = h.WithGroup(g)
	}
	return NewParallelHandler(newHandlers...)
}
