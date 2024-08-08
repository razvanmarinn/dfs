package nodes

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	batchingprocessor "github.com/razvanmarinn/dfs/internal/batching_processor"
	"github.com/razvanmarinn/dfs/internal/load_balancer"
	pb "github.com/razvanmarinn/dfs/proto"
)

func (mn *MasterNode) GetFiles() map[string][]uuid.UUID {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	return mn.files
}

func (mn *MasterNode) GetFileBatches(file string) []uuid.UUID {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	return mn.files[file]
}

func (mn *MasterNode) GetBatchLocations(batchUUID uuid.UUID) []string {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	return mn.batchLocations[batchUUID]
}

func (mn *MasterNode) GetAllBatchLocations() map[uuid.UUID][]string {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	return mn.batchLocations
}

func (mn *MasterNode) Start() {
	fmt.Println("MasterNode started")

}

func (mn *MasterNode) Stop() {
}

func (mn *MasterNode) HealthCheck() bool {
	return true
}
func (mn *MasterNode) AddFile(file string, data []byte) {
	mn.lock.Lock()
	defer mn.lock.Unlock()

	bp := batchingprocessor.NewBatchProcessor(data)
	batches := bp.Process()

	mn.files[file] = make([]uuid.UUID, 0, len(batches))
	for _, batch := range batches {
		mn.files[file] = append(mn.files[file], batch.UUID)
		mn.batchLocations[batch.UUID] = []string{}

		// Write batch data to disk
		batchPath := filepath.Join(mn.batchDir, batch.UUID.String())
		if err := ioutil.WriteFile(batchPath, batch.Data, 0644); err != nil {
			fmt.Printf("Error writing batch %s to disk: %v\n", batch.UUID, err)
			continue
		}

		fmt.Printf("File %s added with batch %s (size: %d bytes)\n", file, batch.UUID, len(batch.Data))
	}
	fmt.Printf("File %s split into %d batches\n", file, len(batches))
}

func (mn *MasterNode) GetBatchData(batchID uuid.UUID) ([]byte, error) {
	mn.lock.Lock()
	defer mn.lock.Unlock()

	batchPath := filepath.Join(mn.batchDir, batchID.String())
	return ioutil.ReadFile(batchPath)
}

func (mn *MasterNode) CleanupBatch(batchID uuid.UUID) error {
	mn.lock.Lock()
	defer mn.lock.Unlock()

	batchPath := filepath.Join(mn.batchDir, batchID.String())
	return os.Remove(batchPath)
}

func (mn *MasterNode) UpdateBatchLocation(batchUUID uuid.UUID, workerNodeID string) {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	mn.batchLocations[batchUUID] = append(mn.batchLocations[batchUUID], workerNodeID)
	fmt.Printf("Batch %s location updated to %s\n", batchUUID, workerNodeID)
}

func (m *MasterNode) GetFileBackFromWorkers(filename string, lb *load_balancer.LoadBalancer) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	batchIDs := m.files[filename]
	fileData := make([]byte, 0)

	for _, batch := range batchIDs {
		// Get the worker ID allocated for this batch
		worker_id, err := m.GetWorkerAllocatedForBatch(batch)
		if err != nil {
			return nil, err
		}

		// Retrieve the client for the corresponding worker ID
		client := lb.GetClientByWorkerID(worker_id)
		if client == nil {
			return nil, fmt.Errorf("no client found for worker ID: %s", worker_id)
		}

		// Create a request to get the batch data
		req := &pb.GetBatchRequest{
			BatchId: &pb.UUID{Value: batch.String()},
		}

		// Make the request to the worker node
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		resp, err := client.GetBatch(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to get batch data from worker %s: %v", worker_id, err)
		}

		// Append the batch data to the complete file data
		fileData = append(fileData, resp.BatchData...)
	}

	return fileData, nil
}

func (m *MasterNode) GetWorkerAllocatedForBatch(batchID uuid.UUID) (string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	workers := m.batchLocations[batchID]
	if len(workers) == 0 {
		return "", fmt.Errorf("no workers allocated for batch %s", batchID)
	}

	workerID := workers[0]
	return workerID, nil

}
