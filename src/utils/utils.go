package utils

import (
	"context"
	"time"
)

func ConsumeMemory(ctx context.Context, sizeMB int) {
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

func ConsumeCPU(ctx context.Context, targetLoad float64) {
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