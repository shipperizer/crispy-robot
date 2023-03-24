package search

import (
	"encoding/json"
	"net/http"

	bleve "github.com/blevesearch/bleve/v2"
	chi "github.com/go-chi/chi/v5"
	"github.com/shipperizer/miniature-monkey/v2/logging"
	etcd "go.etcd.io/etcd/client/v3"
)

type SearchRequest struct {
	Term string `json:"term" default:"test"`
}

type Blueprint struct {
	client *etcd.Client

	keyPrefix string
	index     bleve.Index
	logger    logging.LoggerInterface
}

func (b *Blueprint) Routes(router *chi.Mux) {
	router.Post("/api/v0/search", b.search)
	router.Post("/api/v0/etcd", b.etcd)
}

func (b *Blueprint) search(w http.ResponseWriter, r *http.Request) {
	search := new(SearchRequest)

	if err := json.NewDecoder(r.Body).Decode(search); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	searchResult, err := b.index.Search(
		bleve.NewSearchRequest(
			bleve.NewFuzzyQuery(search.Term),
		),
	)

	if err != nil {
		b.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(searchResult)
	w.WriteHeader(http.StatusOK)
}

func (b *Blueprint) etcd(w http.ResponseWriter, r *http.Request) {
	search := new(SearchRequest)

	if err := json.NewDecoder(r.Body).Decode(search); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eR, err := b.client.Get(r.Context(), search.Term, etcd.WithPrefix())

	if err != nil {
		b.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b.logger.Info(eR.Kvs)
	json.NewEncoder(w).Encode(eR.Kvs)
	w.WriteHeader(http.StatusOK)

}

func NewBlueprint(keyPrefix string, index bleve.Index, client *etcd.Client, logger logging.LoggerInterface) *Blueprint {
	b := new(Blueprint)

	b.client = client
	b.keyPrefix = keyPrefix
	b.index = index
	b.logger = logger

	return b
}
