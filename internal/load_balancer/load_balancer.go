package load_balancer

import (
	"fmt"
	"log"
	"sync"

	pb "github.com/razvanmarinn/dfs/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LoadBalancer struct {
	clients     []pb.BatchServiceClient
	batchCounts []int
	currentNode int
	mu          sync.Mutex
}


const MAXIMUM_BATCHES_PER_WORKER = 35

func NewLoadBalancer(numWorkers int, basePort int) *LoadBalancer {
	lb := &LoadBalancer{
		clients:     make([]pb.BatchServiceClient, numWorkers),
		batchCounts: make([]int, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
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
		lb.clients[i] = pb.NewBatchServiceClient(conn)
		log.Printf("Successfully connected to worker node %d", i+1)
	}

	return lb
}

func (lb *LoadBalancer) GetNextClient() (int, pb.BatchServiceClient) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	clientID := lb.currentNode
	client := lb.clients[clientID]

	lb.batchCounts[clientID]++
	fmt.Printf("Batch added to Worker %d (Batch count: %d)\n", clientID+1, lb.batchCounts[clientID])

	if lb.batchCounts[clientID] >= MAXIMUM_BATCHES_PER_WORKER {
		lb.currentNode = (lb.currentNode + 1) % len(lb.clients)
	}

	return clientID + 1, client
}

func (lb *LoadBalancer) PrintStatus() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	fmt.Println("Load Balancer Status:")
	for i, count := range lb.batchCounts {
		fmt.Printf("Worker %d: %d batches\n", i+1, count)
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
