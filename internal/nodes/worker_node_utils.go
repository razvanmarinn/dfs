package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"

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
	w.ReceivedBatches[batchID] = req.BatchData
	w.SaveBatch(batchID)
	log.Printf("Received and stored batch ID: %s, Data length: %d", batchID, len(req.BatchData))

	return &pb.BatchResponse{
		Success:  true,
		WorkerId: &pb.UUID{Value: w.ID},
	}, nil
}

func (w *WorkerNode) SaveBatch(batchID string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	dataToSave := w.ReceivedBatches[batchID]

	filePath := fmt.Sprintf("batches/%s", batchID)

	f, err := os.Create(filePath + ".bin")
	if err != nil {
		log.Print("Error creating temp file: ", err)
	}
	err = binary.Write(f, binary.LittleEndian, dataToSave)
	if err != nil {
		log.Print("Error writing to file: ", err)
	}
	f.Close()

}

func (w *WorkerNode) GetWorkerID(ctx context.Context, req *pb.WorkerIDRequest) (*pb.WorkerIDResponse, error) {
	return &pb.WorkerIDResponse{
		WorkerId: &pb.UUID{Value: w.ID},
	}, nil
}


func (wn *WorkerNode) GetBatch(ctx context.Context, req *pb.GetBatchRequest) (*pb.GetBatchResponse, error) {
	batchID := req.BatchId.Value
	batchData := wn.ReceivedBatches[batchID]

	return &pb.GetBatchResponse{
		BatchData: batchData,
	}, nil
}
