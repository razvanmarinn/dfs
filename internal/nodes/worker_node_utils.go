package nodes

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	pb "github.com/razvanmarinn/dfs/proto"
	master_pb "github.com/razvanmarinn/rcss/proto"
)

func (wn *WorkerNode) Start() {
	fmt.Println("WorkerNode started")
}

func (wn *WorkerNode) Stop() {
}

func (wn *WorkerNode) HealthCheck() bool {
	return true
}

func (w *WorkerNode) SendBatch(ctx context.Context, req *master_pb.ClientRequestToWorker) (*master_pb.WorkerResponse, error) {
	batchID := req.BatchId
	w.ReceivedBatches[batchID] = req.Data
	w.SaveBatch(batchID)
	log.Printf("Received and stored batch ID: %s, Data length: %d", batchID, len(req.Data))

	return &master_pb.WorkerResponse{
		Success: true,
	}, nil
}

func (w *WorkerNode) SaveBatch(batchID string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	dataToSave := w.ReceivedBatches[batchID]

	filePath := fmt.Sprintf("./internal/worker/batches/%s", batchID)

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

func (wn *WorkerNode) GetBatch(ctx context.Context, req *master_pb.Ttt) (*master_pb.WorkerBatchResponse, error) {
	batchID := req.BatchID
	batchData := wn.ReceivedBatches[batchID]

	return &master_pb.WorkerBatchResponse{
		BatchData: batchData,
	}, nil
}
