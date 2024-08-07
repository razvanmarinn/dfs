package nodes

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	batchingprocessor "github.com/razvanmarinn/dfs/internal/batching_processor"
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

