package main

import (
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/razvanmarinn/dfs/src/nodes"
)

const inputDir = "../input/"

func main() {
    server := nodes.NewMasterNode()
    server.Start()

    var wg sync.WaitGroup

    // Run forever
    for {
        processFiles(server, &wg)

        wg.Wait()
        fmt.Println("Files and batch locations:")
        for file, batchUUIDs := range server.GetFiles() {
            fmt.Printf("File: %s\n", file)
            for _, batchUUID := range batchUUIDs {
                batchLocations := server.GetBatchLocations(batchUUID)
                fmt.Printf("  Batch UUID: %s\n", batchUUID)
                fmt.Printf("  Locations: %v\n", batchLocations)
            }
        }

        // Print the counters
        fmt.Printf("Sent messages count: %d\n", server.SentMessagesCount)
        fmt.Printf("Received acknowledgments count: %d\n", server.ReceivedAcksCount)

        // Sleep for a while before checking for new files
        time.Sleep(10 * time.Second)
    }
}

func processFiles(server *nodes.MasterNode, wg *sync.WaitGroup) {
    entries, err := os.ReadDir(inputDir)
    if err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }

    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }

        filePath := inputDir + entry.Name()
        data, err := os.ReadFile(filePath)
        if err != nil {
            fmt.Printf("error reading file %s: %v\n", filePath, err)
            continue
        }

        wg.Add(1)
        go func(fileName string, fileData []byte) {
            defer wg.Done()
            ackChan := server.AddFile(fileName, fileData, wg)

            // Wait for the acknowledgement
            <-ackChan

            fmt.Printf("Processed and acknowledged file: %s\n", fileName)
        }(entry.Name(), data)
    }
}