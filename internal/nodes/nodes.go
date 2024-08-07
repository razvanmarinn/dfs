package nodes

import (
	"sync"

	"github.com/google/uuid"
	pb "github.com/razvanmarinn/dfs/proto"
)

const fileBatchesTopic = "file-batches"
const acknowledgmentsTopic = "acknowledgments"

type Node interface {
	Start()
	Stop()
	HealthCheck() bool
}

type MasterNode struct {
    files          map[string][]uuid.UUID
    batchLocations map[uuid.UUID][]string
    batchDir       string
    lock           sync.Mutex
}

type WorkerNode struct {
	ID   string
	lock sync.Mutex
    receivedBatches map[string][]byte
	pb.UnimplementedBatchServiceServer
}


func NewMasterNode(batchDir string) *MasterNode {
    return &MasterNode{
        files:          make(map[string][]uuid.UUID),
        batchLocations: make(map[uuid.UUID][]string),
        batchDir:       batchDir,
    }
}


func NewWorkerNode() *WorkerNode {
	return &WorkerNode{
		ID: uuid.New().String(),
		receivedBatches: make(map[string][]byte),
	}
}