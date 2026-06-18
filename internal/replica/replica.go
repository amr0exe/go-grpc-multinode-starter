package replica

import (
	"context"
	"grpcdemo/internal/handler"
	"grpcdemo/internal/store"
	"grpcdemo/pb"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedPingServiceServer
	port     string
	httpPort string
	store    *store.KV
}

func NewServer(port string, httpPort string) *Server {
	return &Server{
		port:     port,
		httpPort: httpPort,
		store:    store.NewStore(),
	}
}

func (s *Server) ReplicatePut(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	s.store.Put(in.GetKey(), in.GetValue())
	log.Printf("[Replica %s] State Synchronized: %s - %s", s.port, in.GetKey(), in.GetValue())
	return &pb.PutResponse{Success: true}, nil
}

func (s *Server) Start() {
	go func() {
		lis, err := net.Listen("tcp", s.port)
		if err != nil {
			log.Fatalf("Replica failed to listen on %s : %v", s.port, err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterPingServiceServer(grpcServer, s)

		log.Printf("[Replica] gRPC server listening on port %s", s.port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server execution failed: %v", err)
		}
	}()

	hd := handler.NewKVHandler(s.store, nil)

	// handle GET-req in replica nodes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /kv/{key}", hd.HandleGet)

	log.Printf("[Replica] HTTP readonly server listening on port %s", s.httpPort)
	if err := http.ListenAndServe(s.httpPort, mux); err != nil {
		log.Fatalf("Replica HTTP failed: %v", err)
	}
}
