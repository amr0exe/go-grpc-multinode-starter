package handler

import (
	"context"
	"encoding/json"
	"grpcdemo/internal/store"
	ty "grpcdemo/internal/types"
	"grpcdemo/pb"
	"log"
	"net/http"
	"sync"
	"time"
)

type KVStore struct {
	db       *store.KV
	Replicas []pb.PingServiceClient
}

func NewKVHandler(db *store.KV, rpl []pb.PingServiceClient) *KVStore {
	return &KVStore{
		db:       db,
		Replicas: rpl,
	}
}

func (e *KVStore) HandleSET(w http.ResponseWriter, r *http.Request) {
	var req ty.KVRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
		http.Error(w, "Invalid json payload or missing key", http.StatusBadRequest)
		return
	}

	// local mutation
	e.db.Put(req.Key, req.Value)
	log.Printf("[Primary] Mutated Locally: %s - %s. Replicating ...", req.Key, req.Value)

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i, client := range e.Replicas {
		wg.Add(1)

		go func(addr string, c pb.PingServiceClient) {
			defer wg.Done()
			res, err := c.ReplicatePut(ctx, &pb.PutRequest{Key: req.Key, Value: req.Value})
			if err != nil {
				log.Printf("[Primary] ERROR updateing replica %s - %v", addr, err)
				return
			}
			if res.GetSuccess() {
				log.Printf("[Primary] Replica %s acknowledged write", addr)
			}
		}(ty.ReplicaAddr[i], client)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "committed_and_replicated"})
}

func (e *KVStore) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	val, found := e.db.Get(key)
	if !found {
		http.Error(w, "key not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
}
