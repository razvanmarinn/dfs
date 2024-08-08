package nodes

import (
	"encoding/json"
	"os"
	"sync"
)

const stateFile = "worker_node_state.json"

// WorkerNodeState represents the state of a worker node
type WorkerNodeState struct {
	ID              string            `json:"id"`
	ReceivedBatches map[string][]byte `json:"received_batches"`
	lock            sync.Mutex
}

// NewWorkerNodeState initializes a new WorkerNodeState
func NewWorkerNodeState() *WorkerNodeState {
	return &WorkerNodeState{
		ReceivedBatches: make(map[string][]byte),
	}
}

// GetState returns the current state as a byte slice
func (w *WorkerNodeState) GetState() ([]byte, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	data, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// LoadState sets the state from a byte slice
func (w *WorkerNodeState) LoadState(data []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if err := json.Unmarshal(data, w); err != nil {
		return err
	}
	return nil
}

// UpdateState updates the state with new data
func (w *WorkerNodeState) UpdateState(data []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.LoadState(data)
}

// SaveState saves the current state to a file
func (w *WorkerNodeState) SaveState() error {
	data, err := w.GetState()
	if err != nil {
		return err
	}

	file, err := os.Create(stateFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// LoadStateFromFile loads the state from a file
func (w *WorkerNodeState) LoadStateFromFile() error {
	file, err := os.Open(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// If file does not exist, consider it as no state
			return nil
		}
		return err
	}
	defer file.Close()

	data := make([]byte, 0)
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	data = make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		return err
	}

	return w.LoadState(data)
}

// SetID sets the ID for the worker node
func (w *WorkerNodeState) SetID(id string) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.ID = id
}
