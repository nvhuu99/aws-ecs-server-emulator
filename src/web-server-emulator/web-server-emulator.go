package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"
	"emulator/utils"
)

var (
	addr = flag.String("addr", ":80", "")
	cpu = flag.Float64("cpu", 40, "")
	mem = flag.Int("mem", 50, "")
	runTime = flag.Int("runtime", 30, "")
)

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
			go utils.ConsumeCPU(ctx, *cpu)
		}
		go utils.ConsumeMemory(ctx, *mem)
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
