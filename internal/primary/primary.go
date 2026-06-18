package primary

import (
	"grpcdemo/internal/handler"
	"grpcdemo/internal/store"
	ty "grpcdemo/internal/types"
	"grpcdemo/pb"
	"log"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	store    *store.KV
	replicas []pb.PingServiceClient
	conns    []*grpc.ClientConn
}

func NewClient() *Client {
	return &Client{
		store: store.NewStore(),
	}
}

func (e *Client) Start(httpPort string) {
	for _, addr := range ty.ReplicaAddr {
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

	hd := handler.NewKVHandler(e.store, e.replicas)

	http.HandleFunc("PUT /set", hd.HandleSET)
	http.HandleFunc("GET /kv/{key}", hd.HandleGet)

	log.Printf("[Primary] HTTP server running on port %s", httpPort)
	if err := http.ListenAndServe(httpPort, nil); err != nil {
		log.Fatalf("Http server execution failed: %v", err)
	}
}
