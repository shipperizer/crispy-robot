package echo

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shipperizer/miniature-monkey/v2/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type EchoRequest struct {
	Message string `json:"message" default:"echo"`
}

type Blueprint struct {
	service *Service

	tracer *tracing.Tracer
}

func (b *Blueprint) Routes(router *chi.Mux) {
	router.Post("/api/v0/echo", b.echo)
}

func (b *Blueprint) echo(w http.ResponseWriter, r *http.Request) {
	echo := new(EchoRequest)

	if err := json.NewDecoder(r.Body).Decode(echo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, span := b.tracer.Start(r.Context(), "handler.Echo", trace.WithAttributes(attribute.String("message", echo.Message)))
	defer span.End()

	json.NewEncoder(w).Encode(map[string]bool{"echo": b.service.Echo(r.Context(), echo.Message)})
	w.WriteHeader(http.StatusOK)
}

func NewBlueprint(service *Service, tracer *tracing.Tracer) *Blueprint {
	b := new(Blueprint)

	b.service = service
	b.tracer = tracer

	return b
}
