package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/razvanmarinn/dfs/proto"

	"github.com/google/uuid"
	"github.com/razvanmarinn/dfs/internal/load_balancer"
	"github.com/razvanmarinn/dfs/internal/nodes"
)

const (
	inputDir  = "input/"
	stateFile = "master_node_state.json"
)

func processFiles(server *nodes.MasterNode, lb *load_balancer.LoadBalancer) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		log.Printf("error reading directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := inputDir + entry.Name()
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("error reading file %s: %v\n", filePath, err)
			continue
		}

		fileName := entry.Name()
		log.Printf("Processing file: %s", fileName)

		server.AddFile(fileName, data)
		batches := server.GetFileBatches(fileName)

		for _, batchID := range batches {
			batchData, err := server.GetBatchData(batchID)
			if err != nil {
				log.Printf("Error reading batch %s: %v", batchID, err)
				continue
			}

			log.Printf("Sending batch %s for file %s (size: %d bytes)", batchID, fileName, len(batchData))
			_, client := lb.GetNextClient()

			success, worker_id := sendBatch(client, batchID, batchData)
			if success {
				log.Printf("Successfully sent batch %s for file %s", batchID, fileName)
				server.UpdateBatchLocation(batchID, worker_id)

				// Clean up the batch file after successful send
				if err := server.CleanupBatch(batchID); err != nil {
					log.Printf("Error cleaning up batch %s: %v", batchID, err)
				}
			} else {
				log.Printf("Failed to send batch %s for file %s", batchID, fileName)
			}
		}
	}
}

func sendBatch(client pb.BatchServiceClient, batchID uuid.UUID, batchData []byte) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.BatchRequest{
		BatchId:   &pb.UUID{Value: batchID.String()},
		BatchData: batchData,
	}

	log.Printf("Sending batch request for batch ID: %s", batchID)
	res, err := client.SendBatch(ctx, req)
	if err != nil {
		log.Printf("Error sending batch: %v", err)
		return false, ""
	}

	log.Printf("Batch sent successfully. Worker ID: %s", res.WorkerId.Value)
	return res.Success, res.WorkerId.Value
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	state := nodes.NewMasterNodeState()

	if err := state.LoadStateFromFile(); err != nil {
		log.Fatalf("Failed to load master node state: %v", err)
	}
	tempDir := "batch_data"
	err := os.Mkdir(tempDir, 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to create temp directory: %v", err)
		}
	}
	// defer os.RemoveAll(tempDir)

	server := nodes.NewMasterNodeWithState(state)

	server.Start()

	numWorkers := 1
	basePort := 50051
	lb := load_balancer.NewLoadBalancer(numWorkers, basePort)
	defer lb.Close()

	processFiles(server, lb)

	for _, file := range server.GetFiles() {
		fmt.Printf("File: %s\n", file.Name)
		for _, batch := range file.Batches {
			locations := server.GetBatchLocations(batch)
			fmt.Printf("Batch: %s, Locations: %v\n", batch, locations)
		}
	}

	// Save state periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := state.SaveState(); err != nil {
				log.Printf("Failed to save master node state: %v", err)
			}
		}
	}()

	// Handle graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		sig := <-sigs
		log.Printf("Received signal: %v", sig)
		log.Println("Shutting down master node...")

		// Save state
		// update the state of the master node

		state.UpdateState(server)
		if err := state.SaveState(); err != nil {
			log.Printf("Failed to save master node state: %v", err)
		}

		server.Stop()
		log.Println("Master node stopped")
	}()

	log.Println("Master node is running. Press Ctrl+C to stop.")
	wg.Wait()
	log.Println("Main function exiting")
}
