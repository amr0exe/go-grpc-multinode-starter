# Multi_Node_Comm

```bash
# first start by listening from replicas/server
go run server/main.go --port=:5001
go run server/main.go --port=:5002
go run server/main.go --port=:5003

# start primary_node/client
go run client/main.go

# testing
sudo apt install httpie # Optional, for curl

http --json PUT :8080/set key=foo value=bar
http GET :8080/kv/foo
```
