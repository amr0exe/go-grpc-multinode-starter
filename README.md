# Multi_Node_Comm

```bash
# Initiate with local_build
make build

# Then, start by listening from replicas/server
make r1
make r2
make r3

# start primary_node/client
make pm

# testing
sudo apt install httpie # Optional, for curl

# mutation on primary_node
http PUT :8080/set key=foo value=bar
http GET :8080/kv/foo

# test for data on replica nodes
http GET :8081/kv/foo
http GET :8082/kv/foo
http GET :8083/kv/foo
```
