package ccontext

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// ctxKey используется как ключ для хранения значений в контексте.
type ctxKey string

const (
	methodNameKey      ctxKey = "methodName"
	userDataKey        ctxKey = "userData"
	methodHashKey      ctxKey = "methodHash"
	methodStartTimeKey ctxKey = "methodStartTime"
	traceStartKey      ctxKey = "traceStart"
	graphqlTypeKey     ctxKey = "graphqlRequestType"
)

func WithMethodName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, methodNameKey, name)
}

func WithMethodStartTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, methodStartTimeKey, t)
}

func WithHash(ctx context.Context, hash uint32) context.Context {
	return context.WithValue(ctx, methodHashKey, hash)
}

func MethodNameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(methodNameKey).(string); ok {
		return v
	}
	return ""
}

func MethodHashFromContext(ctx context.Context) uint32 {
	if v, ok := ctx.Value(methodHashKey).(uint32); ok {
		return v
	}
	return 0
}

func TraceStartTimeFromContext(ctx context.Context) time.Time {
	if v, ok := ctx.Value(traceStartKey).(time.Time); ok {
		return v
	}
	return time.Time{}
}

func MethodStartTimeFromContext(ctx context.Context) time.Time {
	if v, ok := ctx.Value(methodStartTimeKey).(time.Time); ok {
		return v
	}
	return time.Time{}
}

func SpanFromContext(ctx context.Context) *trace.SpanContext {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.TraceID().IsValid() {
		return &sc
	}
	return nil
}
