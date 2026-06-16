package main

import (
	"context"
	"fmt"
	"grpcdemo/pb"
	"log"
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

type Primary struct {
	replicas []pb.PingServiceClient // replicas are just client_conn through which we communicate
	conns    []*grpc.ClientConn     // purpose behind storing conn is just accessible way to close conn
}

func main() {
	p := &Primary{}

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

	fmt.Println("Primary node initialized. Simulating a data mutation ...")
	mutationData := "User_ID_42 updated name to 'Bob'"
	p.replicateMutation(mutationData)
}

func (p *Primary) replicateMutation(data string) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	log.Printf("[Primary] Broacastign Mutation: %v", data)

	for i, replica := range p.replicas {
		wg.Add(1)

		go func(replicaIdx int, client pb.PingServiceClient) {
			defer wg.Done()

			addr := replicaAddr[replicaIdx]
			resp, err := client.SendPing(ctx, &pb.PingRequest{Greetign: data})
			if err != nil {
				log.Printf("[Primary] Error: failed to replicate to %s : %v", addr, err)
				return
			}

			log.Printf("[Primary] Ack from %s : %s", addr, resp.GetReply())
		}(i, replica)
	}

	wg.Wait()
	log.Println("[Primary] Replication round finished.")
}
