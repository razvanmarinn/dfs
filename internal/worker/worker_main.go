package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/razvanmarinn/dfs/internal/nodes"
	pb "github.com/razvanmarinn/dfs/proto"
	master_pb "github.com/razvanmarinn/rcss/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {
	log.Println("Starting worker node...")

	// Initialize worker node state
	state := nodes.NewWorkerNodeState()

	// Load state from file
	if err := state.LoadStateFromFile(); err != nil {
		log.Fatalf("failed to load state: %v", err)
	}

	// Set up the gRPC server
	address := ":50052"
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

	worker := nodes.NewWorkerNodeWithState(state)
	worker.Start()

	state.SetID(worker.ID)

	// Register both services
	pb.RegisterLBServiceServer(grpcServer, worker)
	master_pb.RegisterWorkerServiceServer(grpcServer, worker)

	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	log.Println("gRPC server started successfully")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	// Ensure state is saved on shutdown
	go func() {
		defer wg.Done()
		sig := <-sigs
		log.Printf("Received signal: %v", sig)
		log.Println("Shutting down worker node...")

		// Save state
		if _, err := state.UpdateState(worker); err != nil {
			log.Printf("failed to update state: %v", err)
		}
		if err := state.SaveState(); err != nil {
			log.Printf("failed to save state: %v", err)
		}

		grpcServer.GracefulStop()
		log.Println("Worker node stopped")
	}()

	log.Println("Worker node is running. Press Ctrl+C to stop.")
	wg.Wait()
	log.Println("Main function exiting")
}
