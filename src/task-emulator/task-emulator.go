package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"emulator/utils"
)

func main() {
	// declare and validate inputs
	targetUsage := flag.Float64("cpu", 40, "The desired cpu usage in percentage. The default is 40 percent.")
	targetMemory := flag.Int("mem", 256, "The desired memory usage in megabytes. The default is 256 MB.")
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
		go utils.ConsumeCPU(ctx, *targetUsage)
	}
	go utils.ConsumeMemory(ctx, *targetMemory)
	// wait for timeout
	fmt.Printf("emulating running for %v seconds: %v percent of cpu - %v MB of memory\n", *runTime, *targetUsage, *targetMemory)
	<-ctx.Done()
	fmt.Printf("cpu stress emulator stopped.\n")
}
