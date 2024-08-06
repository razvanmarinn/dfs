package nodes

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/razvanmarinn/dfs/src/batches"
	batchingprocessor "github.com/razvanmarinn/dfs/src/batching_processor"
	"github.com/segmentio/kafka-go"
)

const fileBatchesTopic = "file-batches"
const acknowledgmentsTopic = "acknowledgments"

type Node interface {
	Start()
	Stop()
	HealthCheck() bool
}

type MasterNode struct {
	files             map[string][]uuid.UUID      // file name to batch UUIDs
	batchLocations    map[uuid.UUID][]string      // batch UUID to DataNode IDs
	ackChannels       map[uuid.UUID]chan struct{} // batch UUID to acknowledgment channels
	lock              sync.Mutex
	writer            *kafka.Writer
	reader            *kafka.Reader
	SentMessagesCount int
	ReceivedAcksCount int
}

type WorkerNode struct {
	ID     string
	writer *kafka.Writer
	reader *kafka.Reader
	lock   sync.Mutex
}

func NewMasterNode() *MasterNode {
	return &MasterNode{
		files:          make(map[string][]uuid.UUID),
		batchLocations: make(map[uuid.UUID][]string),
		ackChannels:    make(map[uuid.UUID]chan struct{}),
		writer: &kafka.Writer{
			Addr:  kafka.TCP("localhost:9092"),
			Topic: fileBatchesTopic,
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   acknowledgmentsTopic,
		}),
	}
}

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

func (mn *MasterNode) Start() {
	fmt.Println("MasterNode started")
	go mn.consumeAcknowledgments()
}

func (mn *MasterNode) Stop() {
	mn.writer.Close()
	mn.reader.Close()
}

func (mn *MasterNode) HealthCheck() bool {
	return true
}

func (mn *MasterNode) AddFile(file string, data []byte, wg *sync.WaitGroup) chan struct{} {
	mn.lock.Lock()
	defer mn.lock.Unlock()

	ackChan := make(chan struct{})
	bp := batchingprocessor.NewBatchProcessor(data)
	batches := bp.Process()
	wg.Add(len(batches))

	go func() {
		for _, batch := range batches {
			mn.files[file] = append(mn.files[file], batch.UUID)
			mn.batchLocations[batch.UUID] = []string{}
			mn.ackChannels[batch.UUID] = ackChan // Register the acknowledgment channel
			fmt.Printf("File %s added with batch %s\n", file, batch.UUID)
			go mn.sendBatchToKafka(batch, ackChan, wg)
		}
		wg.Wait()
		close(ackChan) // Close the channel to signal completion
	}()
	return ackChan
}

func (mn *MasterNode) sendBatchToKafka(batch batches.Batch, ackChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Attempt to send the message to Kafka
	err := mn.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(batch.UUID.String()),
		Value: batch.Data,
	})
	if err != nil {
		log.Fatalf("Failed to send message to Kafka: %s", err)
	}

	mn.lock.Lock()
	mn.SentMessagesCount++
	mn.lock.Unlock()

	// Wait for acknowledgment
	select {
	case <-ackChan:
		fmt.Printf("Acknowledgment received for batch %s\n", batch.UUID)
	case <-time.After(35 * time.Second):
		log.Printf("Timed out waiting for acknowledgment for batch %s\n", batch.UUID)
	}
}

func (mn *MasterNode) consumeAcknowledgments() {
	for {
		msg, err := mn.reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Failed to read acknowledgment: %s", err)
		}

		batchUUID, err := uuid.Parse(string(msg.Key))
		if err != nil {
			log.Fatalf("Failed to parse batch UUID: %s", err)
		}

		workerNodeID := string(msg.Value)
		mn.UpdateBatchLocation(batchUUID, workerNodeID)

		mn.lock.Lock()
		ackChan, exists := mn.ackChannels[batchUUID]
		mn.lock.Unlock()

		if exists {
			// Signal completion
			select {
			case ackChan <- struct{}{}:
				fmt.Printf("Acknowledgment received for batch %s from %s\n", batchUUID, workerNodeID)
				mn.lock.Lock()
				mn.ReceivedAcksCount++
				mn.lock.Unlock()
			default:
				// ackChan is closed or not ready to receive
			}
		} else {
			fmt.Printf("No acknowledgment channel found for batch %s\n", batchUUID)
		}
	}
}

func (mn *MasterNode) UpdateBatchLocation(batchUUID uuid.UUID, workerNodeID string) {
	mn.lock.Lock()
	defer mn.lock.Unlock()
	mn.batchLocations[batchUUID] = append(mn.batchLocations[batchUUID], workerNodeID)
	fmt.Printf("Batch %s location updated to %s\n", batchUUID, workerNodeID)
}

func NewWorkerNode() *WorkerNode {
	return &WorkerNode{
		ID: uuid.New().String(),
		writer: &kafka.Writer{
			Addr:         kafka.TCP("localhost:9092", "localhost:9093", "localhost:9094"),
			Topic:        acknowledgmentsTopic,
			RequiredAcks: kafka.RequireAll,
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
            Brokers:     []string{"localhost:9092"},
            Topic:       fileBatchesTopic,
            GroupID:     "worker-group", // Ensure a unique consumer group ID
            StartOffset: kafka.LastOffset, // Start reading from the latest offset
        }),
	}
}

func (wn *WorkerNode) Start() {
	fmt.Println("WorkerNode started")
	go wn.consumeBatches()
}

func (wn *WorkerNode) Stop() {
	wn.writer.Close()
	wn.reader.Close()
}

func (wn *WorkerNode) HealthCheck() bool {
	return true
}

func (wn *WorkerNode) consumeBatches() {
	for {
		msg, err := wn.reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Failed to read message: %s", err)
		}

		batchUUID := string(msg.Key)
		fmt.Printf("Processing batch %s\n", batchUUID)

		// Simulate processing
		time.Sleep(2 * time.Second)

		// Send acknowledgment
		ackMsg := kafka.Message{
			Key:   msg.Key,
			Value: []byte(wn.ID),
		}
		err = wn.writer.WriteMessages(context.Background(), ackMsg)
		if err != nil {
			log.Fatalf("Failed to send acknowledgment: %s", err)
		}
	}
}
