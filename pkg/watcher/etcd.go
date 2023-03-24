package watcher

import (
	"context"
	"time"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/shipperizer/miniature-monkey/v2/logging"
	etcd "go.etcd.io/etcd/client/v3"
)

type Watcher struct {
	client *etcd.Client

	keyPrefix string
	index     bleve.Index
	logger    logging.LoggerInterface
}

func (w *Watcher) watch() {
	w.logger.Info("starting watch")

	watchChan := w.client.Watch(etcd.WithRequireLeader(context.Background()), w.keyPrefix, etcd.WithPrefix())

	for r := range watchChan {

		w.logger.Info("listening wtach chan")

		for _, ev := range r.Events {
			w.logger.Infof("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			// TODO @shipperizer add unmarshalling for value
			if err := w.index.Index(string(ev.Kv.Key), string(ev.Kv.Value)); err != nil {
				w.logger.Error(err)
			}
		}
	}
}

func NewWatcher(keyPrefix string, index bleve.Index, client *etcd.Client, logger logging.LoggerInterface) *Watcher {
	w := new(Watcher)

	w.client = client
	w.keyPrefix = keyPrefix
	w.index = index
	w.logger = logger

	// TODO @shipperizer add signalling check for graceful stop
	go w.watch()

	return w
}

type Scanner struct {
	client *etcd.Client

	keyPrefix string
	index     bleve.Index
	logger    logging.LoggerInterface
}

func (s *Scanner) scan() {
	s.logger.Info("starting scan")
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {
		s.logger.Info("ticker time")
		r, err := s.client.Get(etcd.WithRequireLeader(context.Background()), s.keyPrefix, etcd.WithPrefix())

		if err != nil {
			s.logger.Infof("ERROR %s \n", err)
		}

		for _, ev := range r.Kvs {
			s.logger.Infof("%s %q : %q\n", ev, ev.Key, ev.Value)
			// TODO @shipperizer add unmarshalling for value
			if err := s.index.Index(string(ev.Key), string(ev.Value)); err != nil {
				s.logger.Error(err)
			}

		}
	}
}

func NewScanner(keyPrefix string, index bleve.Index, client *etcd.Client, logger logging.LoggerInterface) *Scanner {
	s := new(Scanner)

	s.client = client
	s.keyPrefix = keyPrefix
	s.index = index
	s.logger = logger

	// TODO @shipperizer add signalling check for graceful stop
	go s.scan()

	return s
}
