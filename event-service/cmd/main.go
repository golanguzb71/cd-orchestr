package main

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	handlers "jprq-event/internal/handler"
	pb "jprq-event/protos/pb"
	"log"
	"net"
	"net/http"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	grpcServer := grpc.NewServer()
	eventServer := handlers.NewEventServer(rdb)
	pb.RegisterEventServiceServer(grpcServer, eventServer)

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}
		fmt.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	wsHandler := handlers.NewWebSocketHandler(rdb, eventServer)

	http.HandleFunc("/tunnel", wsHandler.HandleTunnel)

	fmt.Println("WebSocket server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
