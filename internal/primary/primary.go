package primary

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

type KVRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Client struct {
	mu       sync.RWMutex
	store    map[string]string
	replicas []pb.PingServiceClient
	conns    []*grpc.ClientConn
}

func NewClient() *Client {
	return &Client{
		store: make(map[string]string),
	}
}

func (e *Client) Start(httpPort string) {
	for _, addr := range replicaAddr {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("[Primary] Failed to connect to replica: %s - %v", addr, err)
		}

		e.conns = append(e.conns, conn)
		e.replicas = append(e.replicas, pb.NewPingServiceClient(conn))
	}

	defer func() {
		for _, conn := range e.conns {
			conn.Close()
		}
	}()

	http.HandleFunc("PUT /set", e.handleSET)
	http.HandleFunc("GET /kv/{key}", e.handleGET)

	log.Printf("[Primary] HTTP server running on port %s", httpPort)
	if err := http.ListenAndServe(httpPort, nil); err != nil {
		log.Fatalf("Http server execution failed: %v", err)
	}
}

func (e *Client) handleSET(w http.ResponseWriter, r *http.Request) {
	var req KVRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
		http.Error(w, "Invalid json payload or missing key", http.StatusBadRequest)
		return
	}

	// local mutation
	e.mu.Lock()
	e.store[req.Key] = req.Value
	e.mu.Unlock()
	log.Printf("[Primary] Mutated Locally: %s - %s. Replicating ...", req.Key, req.Value)

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i, client := range e.replicas {
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
		}(replicaAddr[i], client)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "committed_and_replicated"})
}

func (e *Client) handleGET(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	e.mu.RLock()
	val, found := e.store[key]
	e.mu.RUnlock()

	if !found {
		http.Error(w, "Key not found", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": val})
}
