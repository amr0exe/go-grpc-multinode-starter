package main

import (
	"context"
	"flag"
	"fmt"
	"grpcdemo/pb"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPingServiceServer
	port string
}

func (s *server) SendPing(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	log.Printf("[Replica %s] Received ping from client: %s", s.port, in.GetGreetign())

	return &pb.PingResponse{
		Reply: "Pong!!!!!!!!!" + in.Greetign,
	}, nil
}

func main() {
	port := flag.String("port", ":5001", "port for replicas")
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPingServiceServer(grpcServer, &server{})

	fmt.Printf("gRPC server is running on port %s ...\n", *port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
