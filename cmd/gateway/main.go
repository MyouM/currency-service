package main

import (
	"currency-service/internal/config"
	handler "currency-service/internal/handler/gateway"
	"currency-service/internal/proto/currpb"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	conn, err := grpc.Dial("localhost:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error grpc dial: %v", err)
	}
	defer conn.Close()

	client := currpb.NewCurrencyServiceClient(conn)

	router := http.NewServeMux()

	/*logger, err := zap.NewProduction(db)
	if err != nil {
		log.Fatalf("error init logger: %v", err)
	}*/

	handler.GatewayHandlersInit(router, &cfg, client)

	fmt.Println("Server running on port 8082...")
	server := http.Server{
		Addr:    ":8082",
		Handler: router,
	}
	server.ListenAndServe()

}
