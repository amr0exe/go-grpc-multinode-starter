package store

import (
	"encoding/json"
	"net/http"
	"sync"
)

type KV struct {
	mu sync.RWMutex
	db map[string]string
}

func NewStore() *KV {
	return &KV{
		db: make(map[string]string),
	}
}

func (r *KV) Put(key, value string) {
	r.mu.Lock()
	r.db[key] = value
	r.mu.Unlock()
}

func (r *KV) Get(key string) (string, bool) {
	r.mu.RLock()
	val, found := r.db[key]
	r.mu.RUnlock()
	return val, found
}

func (r *KV) HandleGet(w http.ResponseWriter, rq *http.Request) {
	key := rq.PathValue("key")
	val, found := r.Get(key)
	if !found {
		http.Error(w, "key not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
}
