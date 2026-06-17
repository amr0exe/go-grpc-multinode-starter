package main

import (
	"flag"
	"grpcdemo/internal/primary"
	"grpcdemo/internal/replica"
)

func main() {
	node := flag.String("node", "replica", "node type")
	port := flag.String("port", ":5001", "port to run on")

	flag.Parse()

	switch *node {
	case "primary":
		primaryNode := primary.NewClient()
		primaryNode.Start(*port)
	case "replica":
		replicaNode := replica.NewServer(*port)
		replicaNode.Start()
	}
}