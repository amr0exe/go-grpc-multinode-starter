package main

import (
	"context"
	"encoding/json"
	"grpcdemo/pb"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var replicaAddr = []string{
	"localhost:5001",
	"localhost:5002",
	"localhost:5003",
}

type PrimaryStore struct {
	mu       sync.RWMutex
	store    map[string]string
	replicas []pb.PingServiceClient // replicas are just client_conn through which we communicate
	conns    []*grpc.ClientConn     // purpose behind storing conn is just accessible way to close conn
}

type KVRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewPrimaryStore() *PrimaryStore {
	return &PrimaryStore{
		store: make(map[string]string),
	}
}

func (p *PrimaryStore) Put(key, value string) {
	p.mu.Lock()
	p.store[key] = value
	p.mu.Unlock()

	log.Printf("[PrimaryStore] Mutated Locally: %s -- %s. Broadcasting Mutation ... ", key, value)

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i, client := range p.replicas {
		wg.Add(1)

		go func(addr string, c pb.PingServiceClient) {
			defer wg.Done()

			res, err := c.ReplicatePut(ctx, &pb.PutRequest{Key: key, Value: value})
			if err != nil {
				log.Printf("[Primary] ERROR updating replica %s:%v", addr, err)
			}
			if res.GetSuccess() {
				log.Printf("[Primary] Replica %s acknowledged write", addr)
			}

		}(replicaAddr[i], client)
	}

	wg.Wait()
	log.Println("[PrimaryStore] Replication round finished.")
}

func (p *PrimaryStore) Get(key string) (string, bool) {
	p.mu.RLock()
	val, exists := p.store[key]
	p.mu.RUnlock()

	return val, exists
}

func main() {
	p := NewPrimaryStore()

	for _, addr := range replicaAddr {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}

		p.conns = append(p.conns, conn)
		p.replicas = append(p.replicas, pb.NewPingServiceClient(conn))
	}
	defer func() {
		for _, conn := range p.conns {
			conn.Close()
		}
	}()

	http.HandleFunc("PUT /set", func(w http.ResponseWriter, r *http.Request) {
		var req KVRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || req.Key == "" {
			http.Error(w, "Invalid json payload or missing key", http.StatusBadRequest)
			return
		}

		// local mutation
		p.Put(req.Key, req.Value)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "committed_and_replicated"})
	})

	http.HandleFunc("GET /kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")	

		val, found := p.Get(key)
		if !found {
			http.Error(w, "Key not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
	})

	log.Printf("[Primary_Node] Http server running on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Http Server failed: %v", err)
	}
}
