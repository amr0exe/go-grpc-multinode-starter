package types

// This package contains types and contants required throughout project

// Address for replica nodes
var ReplicaAddr = []string{
	"localhost:5001",
	"localhost:5002",
	"localhost:5003",
}

// json-format for PUT request on KV store
type KVRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
