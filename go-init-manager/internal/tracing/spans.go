package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartGraphQLOperationSpan — внутренний span доменной GraphQL-операции (под HTTP POST /graphql).
func StartGraphQLOperationSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	base := []attribute.KeyValue{attribute.String("graphql.operation.name", operation)}
	return Tracer().Start(ctx, "graphql."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(append(base, attrs...)...),
	)
}

// StartKafkaProduceSpan — дочерний span при публикации в Kafka (producer).
func StartKafkaProduceSpan(ctx context.Context, topic, correlationID string) (context.Context, trace.Span) {
	return Tracer().Start(ctx, "kafka.Produce",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", topic),
			attribute.String("messaging.message_id", correlationID),
		),
	)
}

// StartKafkaConsumeSpan — корневой span обработки сообщения консьюмером (нет родительского HTTP).
func StartKafkaConsumeSpan(ctx context.Context, topic string) (context.Context, trace.Span) {
	return Tracer().Start(ctx, "kafka.Consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", topic),
		),
	)
}

// StartDBSpan — операция с БД внутри переданного контекста.
func StartDBSpan(ctx context.Context, op string) (context.Context, trace.Span) {
	return Tracer().Start(ctx, "db."+op,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("db.system", "postgresql")),
	)
}

// EndError завершает span с ошибкой (если err != nil).
func EndError(span trace.Span, err error) {
	if err == nil || span == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
