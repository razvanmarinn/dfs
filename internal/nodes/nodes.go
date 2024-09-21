package nodes

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/razvanmarinn/dfs/internal/load_balancer"
	pb "github.com/razvanmarinn/dfs/proto"
)

const fileBatchesTopic = "file-batches"
const acknowledgmentsTopic = "acknowledgments"

var singleInstance *MasterNode
var lock sync.Mutex

type Node interface {
	Start()
	Stop()
	HealthCheck() bool
}

type MasterNode struct {
	ID           string
	FileRegistry []FileMetadata
	LoadBalancer *load_balancer.LoadBalancer
	lock         sync.Mutex
}

type FileMetadata struct {
	Name           string                    `json:"name"`
	Size           int64                     `json:"size"`
	Hash           uint32                    `json:"hash"`
	Batches        []uuid.UUID               `json:"batches"` // list of all the batches for this file in UUID
	BatchSizes     map[uuid.UUID]int         `json:"batchSizes"`
	TotalSize      int                       `json:"totalSize"`
	BatchLocations map[uuid.UUID][]uuid.UUID `json:"batchLocations"` // mapping from the uuid of the batches to the UUUID of the
	// worker node which it belongs ( it can be stored on multiple worker nodes)
}

type WorkerNode struct {
	ID              string
	lock            sync.Mutex
	ReceivedBatches map[string][]byte
	pb.UnimplementedBatchServiceServer
}

func NewMasterNode() *MasterNode {
	return &MasterNode{
		ID:           uuid.New().String(),
		FileRegistry: make([]FileMetadata, 0),
	}
}

func NewWorkerNode() *WorkerNode {
	return &WorkerNode{
		ID:              uuid.New().String(),
		ReceivedBatches: make(map[string][]byte),
	}
}

func NewWorkerNodeWithState(state *WorkerNodeState) *WorkerNode {
	if state.ID == "" {
		return NewWorkerNode()
	}
	return &WorkerNode{
		ID:              state.ID,
		ReceivedBatches: state.getBytesFromPaths(),
	}
}

func NewMasterNodeWithState(state *MasterNodeState) *MasterNode {

	if state.ID == "" {
		return NewMasterNode()
	}

	return &MasterNode{
		ID:           state.ID,
		FileRegistry: state.Files,
	}
}

func GetMasterNodeInstance() *MasterNode {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			state := NewMasterNodeState()
			if err := state.LoadStateFromFile(); err != nil {
				log.Fatalf("Failed to load master node state: %v", err)
			}
			singleInstance = NewMasterNodeWithState(state)
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}

	return singleInstance
}
