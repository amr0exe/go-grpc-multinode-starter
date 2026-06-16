package main

import (
	"context"
	"flag"
	"fmt"
	"grpcdemo/pb"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPingServiceServer
	port  string
	mu    sync.RWMutex
	store map[string]string
}

func (r *server) ReplicatePut(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	r.mu.Lock()
	r.store[in.GetKey()] = in.GetValue()
	r.mu.Unlock()

	log.Printf("[Replica %s] State Updated:  %s = %s", r.port, in.GetKey(), in.GetValue())
	return &pb.PutResponse{Success: true}, nil
}

func main() {
	port := flag.String("port", ":5001", "port for replicas")
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}

	server := &server{
		port: *port,
		store: make(map[string]string),
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPingServiceServer(grpcServer, server)

	fmt.Printf("replica server is running on port %s ...\n", *port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
