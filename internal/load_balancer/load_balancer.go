package load_balancer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/razvanmarinn/dfs/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LoadBalancer struct {
	clients     map[string]pb.BatchServiceClient // Map from worker UUID to a client
	batchCounts map[string]int                   // Map from worker UUID to batch count
	mu          sync.Mutex
}

const MAXIMUM_BATCHES_PER_WORKER = 35

func NewLoadBalancer(numWorkers int, basePort int) *LoadBalancer {
	lb := &LoadBalancer{
		clients:     make(map[string]pb.BatchServiceClient),
		batchCounts: make(map[string]int),
	}

	for i := 0; i < numWorkers; i++ {
		// Dial the worker at the specified port
		conn, err := grpc.Dial(
			fmt.Sprintf("localhost:%d", basePort+i),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(64*1024*1024),
				grpc.MaxCallSendMsgSize(64*1024*1024),
			),
		)
		if err != nil {
			log.Fatalf("did not connect to worker %d: %v", i+1, err)
		}

		client := pb.NewBatchServiceClient(conn)

		// Make an initial request to get the worker ID
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req := &pb.WorkerIDRequest{} // No parameters needed
		resp, err := client.GetWorkerID(ctx, req)
		if err != nil {
			log.Fatalf("failed to get worker ID for worker %d: %v", i+1, err)
		}

		workerID := resp.WorkerId.Value
		lb.clients[workerID] = client
		lb.batchCounts[workerID] = 0 

		log.Printf("Successfully connected to worker node %d with UUID %s", i+1, workerID)
	}

	return lb
}

func (lb *LoadBalancer) GetNextClient() (string, pb.BatchServiceClient) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var selectedWorkerID string
	var selectedClient pb.BatchServiceClient

	// Simple round-robin or load-based selection strategy
	for workerID, client := range lb.clients {
		if lb.batchCounts[workerID] < MAXIMUM_BATCHES_PER_WORKER {
			selectedWorkerID = workerID
			selectedClient = client
			lb.batchCounts[workerID]++
			fmt.Printf("Batch added to Worker %s (Batch count: %d)\n", workerID, lb.batchCounts[workerID])
			break
		}
	}

	// Optionally, update currentNode for round-robin selection if needed

	return selectedWorkerID, selectedClient
}

func (lb *LoadBalancer) PrintStatus() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	fmt.Println("Load Balancer Status:")
	for workerID, count := range lb.batchCounts {
		fmt.Printf("Worker %s: %d batches\n", workerID, count)
	}
	fmt.Println()
}

func (lb *LoadBalancer) Close() {
	for _, client := range lb.clients {
		if conn, ok := client.(grpc.ClientConnInterface); ok {
			// conn.Close()
			fmt.Println("Connection closed", conn)
		}
	}
}

func (lb *LoadBalancer) GetClientByWorkerID(workerID string) pb.BatchServiceClient {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	client, exists := lb.clients[workerID]
	if !exists {
		log.Fatalf("No client found for worker ID: %s", workerID)
	}

	return client
}
