package slogging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
)

var levelToColor = map[slog.Level]string{
	slog.LevelDebug: "\033[36m",
	slog.LevelInfo:  "\033[32m",
	slog.LevelWarn:  "\033[33m",
	slog.LevelError: "\033[31m",
}

type prettyHandler struct {
	mut *sync.Mutex

	fixedAttrs  map[string]string
	groupPrefix string
	opts        slog.HandlerOptions
	w           io.Writer
}

// NewPrettyHandler creates a handler that writes colored, multi-line output to
// the specified writer.
// It is designed to be human readable, not machine readable.
func NewPrettyHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	o := slog.HandlerOptions{}
	if opts != nil {
		o = *opts
	}
	if o.Level == nil {
		o.Level = slog.LevelInfo
	}
	return &prettyHandler{
		mut:  &sync.Mutex{},
		opts: o,
		w:    w,
	}
}

func (ph *prettyHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return ph.opts.Level.Level() <= l
}

func (ph *prettyHandler) Handle(ctx context.Context, r slog.Record) error {
	go func() {
		ph.mut.Lock()
		defer ph.mut.Unlock()

		levelColor, ok := levelToColor[r.Level]
		if !ok {
			levelColor = "\033[0m"
		}
		fmt.Fprintf(ph.w, "%s%s [%s] %s\033[0m", levelColor, r.Time.Format("15:04:05"), r.Level.String(), r.Message)

		if ph.opts.AddSource {
			frames := runtime.CallersFrames([]uintptr{r.PC})
			for {
				frame, more := frames.Next()
				fmt.Fprintf(ph.w, " \033[30mat %s:%d\033[0m\n", frame.File, frame.Line)
				if !more {
					break
				}
			}
		} else {
			fmt.Fprint(ph.w, "\n")
		}

		explicitKeys := make(map[string]struct{}, r.NumAttrs())
		r.Attrs(func(a slog.Attr) bool {
			fullKey := fmt.Sprintf("%s%s", ph.groupPrefix, a.Key)
			explicitKeys[fullKey] = struct{}{}
			fmt.Fprint(ph.w, prettyLogAttr(fullKey, a.Value))
			return true
		})

		for k, v := range ph.fixedAttrs {
			if _, ok := explicitKeys[k]; !ok {
				fmt.Fprint(ph.w, v)
			}
		}
	}()

	return nil
}

func (ph *prettyHandler) WithAttrs(a []slog.Attr) slog.Handler {
	newAttrs := make(map[string]string, len(ph.fixedAttrs)+len(a))
	for k, v := range ph.fixedAttrs {
		newAttrs[k] = v
	}
	for _, a := range a {
		fullKey := fmt.Sprintf("%s%s", ph.groupPrefix, a.Key)
		newAttrs[fullKey] = prettyLogAttr(fullKey, a.Value)
	}
	return &prettyHandler{
		mut:         ph.mut,
		fixedAttrs:  newAttrs,
		groupPrefix: ph.groupPrefix,
		opts:        ph.opts,
		w:           ph.w,
	}
}

func (ph *prettyHandler) WithGroup(g string) slog.Handler {
	return &prettyHandler{
		mut:         ph.mut,
		fixedAttrs:  ph.fixedAttrs,
		groupPrefix: fmt.Sprintf("%s%s.", ph.groupPrefix, g),
		opts:        ph.opts,
		w:           ph.w,
	}
}

func prettyLogAttr(key string, val slog.Value) string {
	return fmt.Sprintf("    \033[30m%s: %s\033[0m\n", key, prettyLogValue(val))
}

func prettyLogValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return fmt.Sprintf("\"%s\"", v.String())
	case slog.KindAny:
		if b, ok := v.Any().([]byte); ok {
			return fmt.Sprintf("bytes(%d)", len(b))
		}
		fallthrough
	default:
		return v.String()
	}
}
