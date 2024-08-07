package nodes

import (
	"context"
	"fmt"
	"log"

	pb "github.com/razvanmarinn/dfs/proto" // Adjust the import path as necessary
)

func (wn *WorkerNode) Start() {
	fmt.Println("WorkerNode started")
}

func (wn *WorkerNode) Stop() {
}

func (wn *WorkerNode) HealthCheck() bool {
	return true
}

func (w *WorkerNode) SendBatch(ctx context.Context, req *pb.BatchRequest) (*pb.BatchResponse, error) {
	batchID := req.BatchId.Value
	w.receivedBatches[batchID] = req.BatchData
	log.Printf("Received and stored batch ID: %s, Data length: %d", batchID, len(req.BatchData))

	return &pb.BatchResponse{
		Success:  true,
		WorkerId: &pb.UUID{Value: w.ID},
	}, nil
}
