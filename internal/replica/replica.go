package replica

import (
	"context"
	"grpcdemo/pb"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedPingServiceServer
	port  string
	mu    sync.RWMutex
	store map[string]string
}

func NewServer(port string) *Server {
	return &Server{
		port:  port,
		store: make(map[string]string),
	}
}

func (s *Server) ReplicatePut(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	s.mu.Lock()
	s.store[in.GetKey()] = in.GetValue()
	s.mu.Unlock()

	log.Printf("[Replica %s] State Synchronized: %s - %s", s.port, in.GetKey(), in.GetValue())
	return &pb.PutResponse{Success: true}, nil
}

func (s *Server) Start() {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		log.Fatalf("Replica failed to listen on %s : %v", s.port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPingServiceServer(grpcServer, s)

	log.Printf("[Replica] gRPC server running on port %s", s.port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server execution failed: %v", err)
	}
}