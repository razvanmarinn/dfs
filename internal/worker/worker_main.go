package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/razvanmarinn/dfs/internal/nodes"
	pb "github.com/razvanmarinn/dfs/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {
	log.Println("Starting worker node...")
	worker := nodes.NewWorkerNode()
	worker.Start()

	// Set up the gRPC server
	address := ":50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Worker node listening on %s", address)

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(64 * 1024 * 1024),
		grpc.MaxSendMsgSize(64 * 1024 * 1024), 
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    10 * time.Second,
			Timeout: 20 * time.Second,
		}),
		grpc.Creds(insecure.NewCredentials()),
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterBatchServiceServer(grpcServer, worker)

	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	log.Println("gRPC server started successfully")


	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker node is running. Press Ctrl+C to stop.")
	<-sigs

	log.Println("Shutting down worker node...")
	grpcServer.GracefulStop()
	log.Println("Worker node stopped")
}
