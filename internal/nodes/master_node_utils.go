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

const batchDir = "batch_data"

func (mn *MasterNode) GetFiles() []FileMetadata {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	return mn.FileRegistry
}

func (mn *MasterNode) GetFileBatches(file string) []uuid.UUID {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	for _, fileMeta := range mn.FileRegistry {
		if fileMeta.Name == file {
			return fileMeta.Batches
		}
	}
	return nil
}

func (mn *MasterNode) GetBatchLocations(batchUUID uuid.UUID) []uuid.UUID {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	for _, fileMeta := range mn.FileRegistry {
		if locations, exists := fileMeta.BatchLocations[batchUUID]; exists {
			return locations
		}
	}
	return nil
}

func (mn *MasterNode) GetAllBatchLocations() map[uuid.UUID][]uuid.UUID {
	mn.lock.Lock()
	defer mn.lock.Unlock()

	result := make(map[uuid.UUID][]uuid.UUID)

	for _, fileMeta := range mn.FileRegistry {
		for batchID, locations := range fileMeta.BatchLocations {
			if existingLocations, exists := result[batchID]; exists {
				result[batchID] = append(existingLocations, locations...)
			} else {
				result[batchID] = locations
			}
		}
	}

	return result
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

	mn.FileRegistry = append(mn.FileRegistry, FileMetadata{
		Name:           file,
		Size:           int64(len(data)),
		Batches:        make([]uuid.UUID, 0, len(batches)),
		BatchSizes:     make(map[uuid.UUID]int),
		TotalSize:      len(data),
		BatchLocations: make(map[uuid.UUID][]uuid.UUID),
	})
	for _, batch := range batches {
		// mn.Files[file] = append(mn.Files[file], batch.UUID)
		// mn.BatchLocations[batch.UUID] = []string{}
		mn.FileRegistry[len(mn.FileRegistry)-1].Batches = append(mn.FileRegistry[len(mn.FileRegistry)-1].Batches, batch.UUID)
		mn.FileRegistry[len(mn.FileRegistry)-1].BatchSizes[batch.UUID] = len(batch.Data)
		mn.FileRegistry[len(mn.FileRegistry)-1].BatchLocations[batch.UUID] = []uuid.UUID{}

		// Write batch data to disk
		batchPath := filepath.Join(batchDir, batch.UUID.String() + ".bin")
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

	batchPath := filepath.Join(batchDir, batchID.String() + ".bin")
	return ioutil.ReadFile(batchPath)
}

func (mn *MasterNode) CleanupBatch(batchID uuid.UUID) error {
	mn.lock.Lock()
	defer mn.lock.Unlock()

	batchPath := filepath.Join(batchDir, batchID.String() + ".bin")
	return os.Remove(batchPath)
}

func (mn *MasterNode) UpdateBatchLocation(batchUUID uuid.UUID, workerNodeID string) {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	for i, fileMeta := range mn.FileRegistry {
		if _, exists := fileMeta.BatchLocations[batchUUID]; exists {
			workerNodeUUID, err := uuid.Parse(workerNodeID)
			if err != nil {
				fmt.Printf("Error parsing worker node ID %s: %v\n", workerNodeID, err)
				return
			}
			mn.FileRegistry[i].BatchLocations[batchUUID] = append(mn.FileRegistry[i].BatchLocations[batchUUID], workerNodeUUID)
		}
	}

	fmt.Printf("Batch %s location updated to %s\n", batchUUID, workerNodeID)
}

func (m *MasterNode) GetFileBackFromWorkers(filename string, lb *load_balancer.LoadBalancer) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	batchIDs := make([]uuid.UUID, 0)
	for _, fileMeta := range m.FileRegistry {
		if fileMeta.Name == filename {
			batchIDs = append(batchIDs, fileMeta.Batches...)
		}
	}
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

	// workers := m.BatchLocations[batchID]
	for _, fileMeta := range m.FileRegistry {
		if workers, exists := fileMeta.BatchLocations[batchID]; exists {
			if len(workers) == 0 {
				return "", fmt.Errorf("no workers allocated for batch %s", batchID)
			}
			workerID := workers[0]
			return workerID.String(), nil
		}
	}
	return "", fmt.Errorf("batch %s not found", batchID)
}
