package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	mbPerSecond    = 100              // MB of RAM to allocate per second
	chunkSize      = mbPerSecond * 1024 * 1024 // bytes allocated per tick
	logFile        = "stress_test.log"
)

var (
	totalAllocated int64 // total MB allocated (atomic)
	cpuOpsCounter  int64 // cpu ops counter (atomic)
	stopChan       = make(chan struct{})
)

func main() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	// Set up file logger
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()
	logger := log.New(f, "", 0)

	header := fmt.Sprintf("=== Stress Test Started at %s ===", time.Now().Format(time.RFC3339))
	fmt.Println(header)
	logger.Println(header)

	info := fmt.Sprintf("System: %d CPUs detected | RAM target: %dMB/sec", numCPU, mbPerSecond)
	fmt.Println(info)
	logger.Println(info)

	// Handle Ctrl+C / SIGTERM gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	// --- CPU stress: one goroutine per logical CPU ---
	fmt.Printf("Spawning %d CPU stress goroutines...\n", numCPU)
	logger.Printf("Spawning %d CPU stress goroutines", numCPU)
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cpuStress(id)
		}(i)
	}

	// --- RAM stress: allocate 100MB every second ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		ramStress(logger)
	}()

	// --- Stats logger: print + log every second ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		statsLogger(logger)
	}()

	// Wait for signal
	sig := <-sigChan
	msg := fmt.Sprintf("\nReceived signal: %v — stopping stress test...", sig)
	fmt.Println(msg)
	logger.Println(msg)

	close(stopChan)
	wg.Wait()

	summary := fmt.Sprintf("=== Stress Test Ended at %s | Total RAM allocated: %dMB ===",
		time.Now().Format(time.RFC3339), atomic.LoadInt64(&totalAllocated))
	fmt.Println(summary)
	logger.Println(summary)
	fmt.Printf("Log written to: %s\n", logFile)
}

// cpuStress runs a tight math loop to saturate one CPU core.
func cpuStress(id int) {
	x := 1.0001 + float64(id)*0.000001
	for {
		select {
		case <-stopChan:
			return
		default:
			// Expensive floating-point work
			for i := 0; i < 100_000; i++ {
				x = math.Sqrt(x*x+1) * math.Sin(x) / math.Cos(x+0.1)
				if math.IsNaN(x) || math.IsInf(x, 0) {
					x = 1.0001
				}
			}
			atomic.AddInt64(&cpuOpsCounter, 100_000)
		}
	}
}

// ramStress allocates 100MB every second and holds the memory.
func ramStress(logger *log.Logger) {
	var held [][]byte // keep references so GC can't collect
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			chunk := make([]byte, chunkSize)
			// Touch every page so the OS actually commits the memory
			for i := range chunk {
				chunk[i] = byte(i & 0xFF)
			}
			held = append(held, chunk)
			mb := atomic.AddInt64(&totalAllocated, mbPerSecond)
			msg := fmt.Sprintf("[RAM] Allocated +%dMB  |  Total held: %dMB", mbPerSecond, mb)
			fmt.Println(msg)
			logger.Println(msg)
		}
	}
}

// statsLogger prints a summary line every second.
func statsLogger(logger *log.Logger) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	start := time.Now()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)

			elapsed := time.Since(start).Round(time.Second)
			ops := atomic.SwapInt64(&cpuOpsCounter, 0)
			ramMB := atomic.LoadInt64(&totalAllocated)
			sysMB := ms.Sys / 1024 / 1024
			gcRuns := ms.NumGC

			line := fmt.Sprintf(
				"[STATS] Elapsed: %-8s | RAM held: %4dMB | RSS ~%4dMB | CPU ops/s: %8d | GC runs: %d",
				elapsed, ramMB, sysMB, ops, gcRuns,
			)
			fmt.Println(line)
			logger.Println(line)
		}
	}
}
