package testutil

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// BenchmarkConfig holds configuration for performance tests
type BenchmarkConfig struct {
	MemoryProfilePath string
	CPUProfilePath    string
	ProfileDuration   time.Duration
	MaxMemoryMB       int
	MaxAllocMB        int
}

// DefaultBenchmarkConfig returns default benchmark configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		MemoryProfilePath: "mem.prof",
		CPUProfilePath:    "cpu.prof",
		ProfileDuration:   30 * time.Second,
		MaxMemoryMB:       100, // 100MB memory limit
		MaxAllocMB:        50,  // 50MB allocation limit
	}
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	AllocMB      float64
	TotalAllocMB float64
	SysMB        float64
	NumGC        uint32
}

// GetMemoryStats returns current memory statistics
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		AllocMB:      bToMB(m.Alloc),
		TotalAllocMB: bToMB(m.TotalAlloc),
		SysMB:        bToMB(m.Sys),
		NumGC:        m.NumGC,
	}
}

// bToMB converts bytes to megabytes
func bToMB(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

// StartCPUProfile starts CPU profiling if enabled
func StartCPUProfile(config *BenchmarkConfig) func() {
	if config.CPUProfilePath == "" {
		return func() {} // No-op if not configured
	}

	f, err := os.Create(config.CPUProfilePath)
	if err != nil {
		fmt.Printf("Could not create CPU profile: %v\n", err)
		return func() {}
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		fmt.Printf("Could not start CPU profile: %v\n", err)
		return func() {}
	}

	return func() {
		pprof.StopCPUProfile()
		f.Close()
		fmt.Printf("CPU profile saved to %s\n", config.CPUProfilePath)
	}
}

// WriteMemoryProfile writes memory profile if enabled
func WriteMemoryProfile(config *BenchmarkConfig) {
	if config.MemoryProfilePath == "" {
		return
	}

	f, err := os.Create(config.MemoryProfilePath)
	if err != nil {
		fmt.Printf("Could not create memory profile: %v\n", err)
		return
	}
	defer f.Close()

	runtime.GC() // Force GC before profiling
	if err := pprof.WriteHeapProfile(f); err != nil {
		fmt.Printf("Could not write memory profile: %v\n", err)
		return
	}

	fmt.Printf("Memory profile saved to %s\n", config.MemoryProfilePath)
}

// BenchmarkRunner provides utilities for running performance benchmarks
type BenchmarkRunner struct {
	config    *BenchmarkConfig
	startTime time.Time
	startMem  MemoryStats
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner(config *BenchmarkConfig) *BenchmarkRunner {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}

	return &BenchmarkRunner{
		config: config,
	}
}

// Start begins the benchmark measurement
func (br *BenchmarkRunner) Start() func() {
	br.startTime = time.Now()
	br.startMem = GetMemoryStats()

	// Start CPU profiling if configured
	stopCPUProfile := StartCPUProfile(br.config)

	return func() {
		stopCPUProfile()
		WriteMemoryProfile(br.config)
		br.printResults()
	}
}

// printResults prints benchmark results
func (br *BenchmarkRunner) printResults() {
	duration := time.Since(br.startTime)
	endMem := GetMemoryStats()

	fmt.Printf("\n=== Benchmark Results ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Memory Usage:\n")
	fmt.Printf("  Start: %.2f MB\n", br.startMem.AllocMB)
	fmt.Printf("  End: %.2f MB\n", endMem.AllocMB)
	fmt.Printf("  Peak: %.2f MB\n", endMem.SysMB)
	fmt.Printf("  Total Allocated: %.2f MB\n", endMem.TotalAllocMB)
	fmt.Printf("  GC Runs: %d\n", endMem.NumGC)
	fmt.Printf("========================\n")
}

// AssertMemoryUsage validates memory usage against limits
func (br *BenchmarkRunner) AssertMemoryUsage(t *testing.T) {
	stats := GetMemoryStats()

	if br.config.MaxMemoryMB > 0 && stats.AllocMB > float64(br.config.MaxMemoryMB) {
		t.Errorf("Memory usage %.2f MB exceeds limit %d MB",
			stats.AllocMB, br.config.MaxMemoryMB)
	}

	if br.config.MaxAllocMB > 0 && stats.TotalAllocMB > float64(br.config.MaxAllocMB) {
		t.Errorf("Total allocation %.2f MB exceeds limit %d MB",
			stats.TotalAllocMB, br.config.MaxAllocMB)
	}
}

// LeakDetector helps detect memory leaks in tests
type LeakDetector struct {
	initialStats MemoryStats
	threshold    float64 // MB
}

// NewLeakDetector creates a new memory leak detector
func NewLeakDetector(thresholdMB float64) *LeakDetector {
	return &LeakDetector{
		initialStats: GetMemoryStats(),
		threshold:    thresholdMB,
	}
}

// Check validates that memory usage hasn't increased beyond threshold
func (ld *LeakDetector) Check(t *testing.T, description string) {
	runtime.GC() // Force GC to clean up
	runtime.GC() // Run twice to ensure cleanup

	currentStats := GetMemoryStats()
	increase := currentStats.AllocMB - ld.initialStats.AllocMB

	if increase > ld.threshold {
		t.Errorf("Memory leak detected in %s: increased by %.2f MB (threshold: %.2f MB)",
			description, increase, ld.threshold)

		t.Logf("Initial memory: %.2f MB", ld.initialStats.AllocMB)
		t.Logf("Current memory: %.2f MB", currentStats.AllocMB)
		t.Logf("GC runs: %d", currentStats.NumGC)
	}
}

// BenchmarkWithTimeout runs a benchmark function with timeout
func BenchmarkWithTimeout(b *testing.B, timeout time.Duration, fn func()) {
	done := make(chan bool, 1)

	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		// Completed successfully
	case <-time.After(timeout):
		b.Fatalf("Benchmark timed out after %v", timeout)
	}
}

// MeasureLatency measures function execution latency
func MeasureLatency(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// ConcurrencyTest runs a function concurrently and measures performance
func ConcurrencyTest(t *testing.T, concurrency int, iterations int, fn func(workerID int)) {
	results := make(chan time.Duration, concurrency)
	start := time.Now()

	// Start workers
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			workerStart := time.Now()

			for j := 0; j < iterations; j++ {
				fn(workerID)
			}

			results <- time.Since(workerStart)
		}(i)
	}

	// Collect results
	var totalWorkerTime time.Duration
	for i := 0; i < concurrency; i++ {
		workerTime := <-results
		totalWorkerTime += workerTime
	}

	totalTime := time.Since(start)
	avgWorkerTime := totalWorkerTime / time.Duration(concurrency)

	t.Logf("Concurrency test results:")
	t.Logf("  Workers: %d", concurrency)
	t.Logf("  Iterations per worker: %d", iterations)
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Average worker time: %v", avgWorkerTime)
	t.Logf("  Efficiency: %.2f%%", float64(avgWorkerTime)/float64(totalTime)*100)
}

// RequireNoGoroutineLeaks checks for goroutine leaks
func RequireNoGoroutineLeaks(t *testing.T, initialCount int, description string) {
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Allow goroutines to finish

	currentCount := runtime.NumGoroutine()

	if currentCount > initialCount {
		require.Equal(t, initialCount, currentCount,
			"Goroutine leak detected in %s: started with %d, ended with %d",
			description, initialCount, currentCount)
	}
}
