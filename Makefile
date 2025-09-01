run:
	go run cmd/migrator/main.go
	go run cmd/currency/main.go &
	go run cmd/gateway/main.go &
