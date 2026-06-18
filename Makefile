build:
	@go build -o app main.go

pm:
	@./app --node=primary --port=:8080

r1:
	@./app --node=replica --port=:5001 --http-port=:8081

r2:
	@./app --node=replica --port=:5002 --http-port=:8082

r3:
	@./app --node=replica --port=:5003 --http-port=:8083

test1:
	@http PUT :8080/set key=foo value=bar

