package echo

import (
	"context"

	"github.com/shipperizer/miniature-monkey/v2/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	store *Store

	tracer *tracing.Tracer
}

func (s *Service) Echo(ctx context.Context, message string) bool {
	_, span := s.tracer.Start(ctx, "service.Echo", trace.WithAttributes(attribute.String("message", message)))
	defer span.End()

	return s.store.Echo(ctx, message)
}
func NewService(store *Store, tracer *tracing.Tracer) *Service {
	s := new(Service)

	s.store = store
	s.tracer = tracer

	return s
}
