package main

import (
    "github.com/razvanmarinn/dfs/src/nodes"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    worker := nodes.NewWorkerNode()
    worker.Start()

    // Create a channel to receive OS signals
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    // Block until a signal is received
    <-sigs
}