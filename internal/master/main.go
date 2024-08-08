package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	pb "github.com/razvanmarinn/dfs/proto"

	"github.com/google/uuid"
	"github.com/razvanmarinn/dfs/internal/load_balancer"
	"github.com/razvanmarinn/dfs/internal/nodes"
)

const inputDir = "../../input/"

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
	tempDir, err := ioutil.TempDir("", "batch_data")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	server := nodes.NewMasterNode(tempDir)
	server.Start()

	numWorkers := 1
	basePort := 50051
	lb := load_balancer.NewLoadBalancer(numWorkers, basePort)
	defer lb.Close()

	processFiles(server, lb)

	for fileName, batchIDs := range server.GetFiles() {
		fmt.Printf("File: %s\n", fileName)
		for _, batchID := range batchIDs {
			locations := server.GetBatchLocations(batchID)
			fmt.Printf("  Batch: %s, Locations: %v\n", batchID, locations)
		}
	}

	select {}
}
