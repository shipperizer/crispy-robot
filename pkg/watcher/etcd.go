package watcher

import (
	"context"
	"time"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/shipperizer/miniature-monkey/v2/logging"
	"github.com/shipperizer/miniature-monkey/v2/tracing"
	etcd "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Watcher struct {
	client *etcd.Client

	keyPrefix string
	index     bleve.Index
	tracer    *tracing.Tracer
	logger    logging.LoggerInterface
}

func (w *Watcher) watch() {
	w.logger.Info("starting watch")

	ctx := context.Background()
	watchChan := w.client.Watch(etcd.WithRequireLeader(ctx), w.keyPrefix, etcd.WithPrefix())

	for r := range watchChan {
		w.logger.Info("listening watch chan")

		for _, ev := range r.Events {
			w.logger.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			_, span := w.tracer.Start(
				ctx,
				"watcher.watch",
				trace.WithAttributes(
					attribute.String("type", string(ev.Type)),
					attribute.String("key", string(ev.Kv.Key)),
					attribute.String("value", string(ev.Kv.Value)),
				),
			)
			// TODO @shipperizer add unmarshalling for value
			if err := w.index.Index(string(ev.Kv.Key), string(ev.Kv.Value)); err != nil {
				w.logger.Error(err)
			}
			span.End()
		}
	}
}

func NewWatcher(keyPrefix string, index bleve.Index, client *etcd.Client, tracer *tracing.Tracer, logger logging.LoggerInterface) *Watcher {
	w := new(Watcher)

	w.client = client
	w.keyPrefix = keyPrefix
	w.index = index
	w.tracer = tracer
	w.logger = logger

	// TODO @shipperizer add signalling check for graceful stop
	go w.watch()

	return w
}

type Scanner struct {
	client *etcd.Client

	keyPrefix string
	index     bleve.Index
	tracer    *tracing.Tracer
	logger    logging.LoggerInterface
}

func (s *Scanner) scan() {
	s.logger.Info("starting scan")
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {
		ctx := context.Background()

		s.logger.Info("ticker time")
		r, err := s.client.Get(etcd.WithRequireLeader(ctx), s.keyPrefix, etcd.WithPrefix())

		if err != nil {
			s.logger.Infof("ERROR %s \n", err)
		}

		for _, ev := range r.Kvs {
			_, span := s.tracer.Start(
				ctx,
				"scanner.scan",
				trace.WithAttributes(
					attribute.String("key", string(ev.Key)),
					attribute.String("value", string(ev.Value)),
				),
			)
			s.logger.Infof("%s %q : %q\n", ev, ev.Key, ev.Value)
			// TODO @shipperizer add unmarshalling for value
			if err := s.index.Index(string(ev.Key), string(ev.Value)); err != nil {
				s.logger.Error(err)
			}
			span.End()

		}
	}
}

func NewScanner(keyPrefix string, index bleve.Index, client *etcd.Client, tracer *tracing.Tracer, logger logging.LoggerInterface) *Scanner {
	s := new(Scanner)

	s.client = client
	s.keyPrefix = keyPrefix
	s.index = index
	s.tracer = tracer
	s.logger = logger

	// TODO @shipperizer add signalling check for graceful stop
	go s.scan()

	return s
}
