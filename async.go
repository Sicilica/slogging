package slogging

import (
	"context"
	"log/slog"
)

type asyncHandler struct {
	handler slog.Handler
}

// NewAsyncHandler wraps a log handler and causes all logs to be written out asynchronously.
// This can be useful in cases where writing logs may take more time than we want to wait for.
func NewAsyncHandler(handler slog.Handler) slog.Handler {
	return asyncHandler{
		handler: handler,
	}
}

func (ah asyncHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return ah.handler.Enabled(ctx, l)
}

func (ah asyncHandler) Handle(ctx context.Context, r slog.Record) error {
	go ah.handler.Handle(ctx, r)
	return nil
}

func (ah asyncHandler) WithAttrs(a []slog.Attr) slog.Handler {
	return NewAsyncHandler(ah.handler.WithAttrs(a))
}

func (ah asyncHandler) WithGroup(g string) slog.Handler {
	return NewAsyncHandler(ah.handler.WithGroup(g))
}
