package logger

import (
	"context"
	"log/slog"

	"gitlab.com/go-init/go-init-common/default/ccontext"
)

const (
	methodName = "methodName"
	traceId    = "traceId"
	spanId     = "spanId"
	// snake_case — удобно для Loki / Grafana (корреляция traces ↔ logs)
	traceIDSnake = "trace_id"
	spanIDSnake  = "span_id"
)

// ContextHandler обогащает логи контекстной информацией.
type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.enrichWithContext(ctx)...) // Добавляем атрибуты из контекста
	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) enrichWithContext(ctx context.Context) (attrs []slog.Attr) {
	if mname := ccontext.MethodNameFromContext(ctx); mname != "" {
		attrs = append(attrs, slog.String(methodName, mname))
	}
	if sc := ccontext.SpanFromContext(ctx); sc != nil {
		tid := sc.TraceID().String()
		sid := sc.SpanID().String()
		attrs = append(attrs,
			slog.String(traceId, tid),
			slog.String(spanId, sid),
			slog.String(traceIDSnake, tid),
			slog.String(spanIDSnake, sid),
		)
	}
	return attrs
}
