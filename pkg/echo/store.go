package echo

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/shipperizer/miniature-monkey/v2/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Store struct {
	client *redis.Client

	tracer *tracing.Tracer
}

func (s *Store) Echo(ctx context.Context, message string) bool {
	_, span := s.tracer.Start(ctx, "store.Echo", trace.WithAttributes(attribute.String("message", message)))
	defer span.End()

	if message == "echo" {
		time.Sleep(5 * time.Millisecond)
	}

	return s.client.GetSet(ctx, message, true) == nil
}

func NewStore(client *redis.Client, tracer *tracing.Tracer) *Store {
	s := new(Store)

	s.client = client
	s.tracer = tracer

	return s
}
