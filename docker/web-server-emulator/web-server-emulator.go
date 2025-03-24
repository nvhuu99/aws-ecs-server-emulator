package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

var (
	addr = flag.String("addr", ":80", "")
	cpu = flag.Float64("cpu", 40, "")
	mem = flag.Int("mem", 50, "")
	runTime = flag.Int("runtime", 30, "")
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

func serveHome(w http.ResponseWriter, r *http.Request) {
	// get the client IP from X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	realIP := r.RemoteAddr
	if xff != "" {
		realIP = xff
	}
	// start the emulation
	go func() {
		// create a deadline based on run time
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*runTime) * time.Second)
		defer cancel()
		// start stress function on all CPU cores
		for i := 0; i < runtime.NumCPU(); i++ {
			go emulateCPU(ctx, *cpu)
		}
		go emulateMemory(ctx, *mem)
		// wait for the emulation finishing
		<-ctx.Done()
		// log the status
		fmt.Printf("emulation completed - ip: %v - cpu: %v - mem: %v - runtime: %v\n", realIP, *cpu, *mem, *runTime)
	}()
	
	// response immediately
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf("client ip: %s<br>", realIP)))
	w.Write([]byte(fmt.Sprintf("emulating running for %v seconds: %v percent of cpu - %v MB of memory.<br>", *runTime, *cpu, *mem)))
	// log the request
	fmt.Printf("served request for client ip %v\n", realIP)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", serveHome)
	fmt.Println("serving web server emulator")
	http.ListenAndServe(*addr, nil)
}
