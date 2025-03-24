package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"
)

func emulateMemory(ctx context.Context, sizeMB int) {
	memUsage := make([][]byte, sizeMB)
	for i := range memUsage {
		memUsage[i] = make([]byte, 1024*1024) // Allocate 1MB per slice
		for j := range memUsage[i] {
			memUsage[i][j] = byte(j % 256)
		}
	}

	idleTime := 100 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(idleTime)
	}
}

func emulateCPU(ctx context.Context, targetLoad float64) {
	cycleTime := 100 * time.Millisecond
	busyTime := time.Duration(float64(cycleTime) * (targetLoad / 100.0))
	idleTime := cycleTime - busyTime
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		start := time.Now()
		for time.Since(start) < busyTime {
		}
		time.Sleep(idleTime)
	}
}

func main() {
	// declare and validate inputs
	targetUsage := flag.Float64("cpu", 40, "The desired cpu usage in percentage. The default is 40 percent.")
	targetMemory := flag.Int("memory", 256, "The desired memory usage in megabytes. The default is 256 MB.")
	runTime := flag.Int("runtime", 10, "The total run time in seconds. The default is 10 seconds.")
	flag.Parse()
	if *targetUsage < 1 || *targetUsage > 100 {
		fmt.Println("error: CPU target must be between 1 and 100.")
		return
	}
	if *targetMemory <= 0 {
		fmt.Println("error: memory target must be greater than 0 MB.")
		return
	}
	if *runTime <= 0 {
		fmt.Println("error: runtime must be greater than 0 seconds.")
		return
	}
	// create a deadline based on run time
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*runTime) * time.Second)
	defer cancel()
	// start stress function on all CPU cores
	for i := 0; i < runtime.NumCPU(); i++ {
		go emulateCPU(ctx, *targetUsage)
	}
	go emulateMemory(ctx, *targetMemory)
	// wait for timeout
	fmt.Printf("emulating running for %v seconds: %v percent of cpu - %v MB of memory\n", *runTime, *targetUsage, *targetMemory)
	<-ctx.Done()
	fmt.Printf("cpu stress emulator stopped.\n")
}
