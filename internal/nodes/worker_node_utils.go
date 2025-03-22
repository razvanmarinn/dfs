package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	pb "github.com/razvanmarinn/datalake/protobuf"
)

func (wn *WorkerNode) Start() {
	fmt.Println("WorkerNode started")
}

func (wn *WorkerNode) Stop() {
}

func (wn *WorkerNode) HealthCheck() bool {
	return true
}

func (w *WorkerNode) ReceiveBatch(ctx context.Context, req *pb.SendClientRequestToWorker) (*pb.WorkerResponse, error) {
	batchID := req.BatchId
	w.ReceivedBatches[batchID] = req.Data
	w.SaveBatch(batchID)
	log.Printf("Received and stored batch ID: %s, Data length: %d", batchID, len(req.Data))

	return &pb.WorkerResponse{
		Success: true,
	}, nil
}

func (w *WorkerNode) SaveBatch(batchID string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	dataToSave := w.ReceivedBatches[batchID]

	filePath := fmt.Sprintf("/data/batches/%s", batchID)

	if err := os.MkdirAll("/data/batches", os.ModePerm); err != nil {
		log.Print("Error creating directory: ", err)
		return
	}

	f, err := os.Create(filePath + ".bin")
	if err != nil {
		log.Print("Error creating file in persistent storage: ", err)
		return
	}
	defer f.Close()

	err = binary.Write(f, binary.LittleEndian, dataToSave)
	if err != nil {
		log.Print("Error writing to file in persistent storage: ", err)
	}
}

func (w *WorkerNode) GetWorkerID(ctx context.Context, req *pb.WorkerIDRequest) (*pb.WorkerIDResponse, error) {
	log.Printf("Received GetWorkerID request")
	return &pb.WorkerIDResponse{
		WorkerId: &pb.UUID{Value: w.ID},
	}, nil
}

func (wn *WorkerNode) RetrieveBatchForClient(ctx context.Context, req *pb.GetClientRequestToWorker) (*pb.WorkerBatchResponse, error) {
	batchID := req.BatchId
	batchData := wn.ReceivedBatches[batchID]

	return &pb.WorkerBatchResponse{
		BatchData: batchData,
	}, nil
}
