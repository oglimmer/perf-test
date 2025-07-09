package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	primeRange     int
	memoryPercent  float64
	chunkSizeMB    int
	reportInterval int
	cpuThreads     int
	full           bool
	disableCPU     bool
	disableDisk    bool
	diskPath       string
}

type CPUStats struct {
	mu               sync.RWMutex
	totalPrimesFound int
	totalTime        time.Duration
	lastReport       time.Time
}

func formatWithCommas(n float64) string {
	str := strconv.FormatFloat(n, 'f', 0, 64)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

func main() {
	var config Config

	// Parse command line arguments
	flag.IntVar(&config.primeRange, "prime-range", 10000000, "Range for prime number testing (default: 10M)")
	flag.Float64Var(&config.memoryPercent, "memory-percent", 0.9, "Percentage of memory to allocate (0.1-0.95)")
	flag.IntVar(&config.chunkSizeMB, "chunk-size", 100, "Memory chunk size in MB")
	flag.IntVar(&config.reportInterval, "report-interval", 5, "Seconds between benchmark reports")
	flag.IntVar(&config.cpuThreads, "cpu-threads", 0, "Number of CPU threads (0 = auto: cores-1)")
	flag.BoolVar(&config.full, "full", false, "Show full output with detailed information")
	flag.BoolVar(&config.disableCPU, "disable-cpu", false, "Disable CPU testing")
	flag.BoolVar(&config.disableDisk, "disable-disk", false, "Disable disk testing")
	flag.StringVar(&config.diskPath, "disk-path", os.TempDir(), "Path for disk benchmark files")
	flag.Parse()

	// Validate parameters
	if config.memoryPercent < 0.1 || config.memoryPercent > 0.95 {
		fmt.Println("Memory percent must be between 0.1 and 0.95")
		os.Exit(1)
	}

	cpuCores := runtime.NumCPU()
	if config.cpuThreads == 0 {
		config.cpuThreads = cpuCores - 1
		if config.cpuThreads < 1 {
			config.cpuThreads = 1
		}
	}

	if config.full {
		fmt.Printf("CPU cores detected: %d\n", cpuCores)
		fmt.Printf("Using %d threads for CPU benchmarking\n", config.cpuThreads)
		fmt.Printf("Prime range: %d\n", config.primeRange)
		fmt.Printf("Memory allocation: %.0f%%\n", config.memoryPercent*100)
		fmt.Printf("Chunk size: %d MB\n", config.chunkSizeMB)
		fmt.Printf("Report interval: %d seconds\n", config.reportInterval)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	stopChan := make(chan struct{})

	// Create shared CPU stats for quiet mode
	cpuStats := &CPUStats{lastReport: time.Now()}

	// Start CPU benchmarking threads
	if !config.disableCPU {
		for i := 0; i < config.cpuThreads; i++ {
			go func(threadID int) {
				benchmarkPrimality(threadID, stopChan, config, cpuStats)
			}(i)
		}
	}

	// Memory allocation and filesystem benchmarking
	if !config.disableDisk {
		go func() {
			memoryAndFilesystemBenchmark(stopChan, config)
		}()
	}

	// Wait for interrupt signal
	<-sigChan
	if config.full {
		fmt.Println("\nReceived interrupt signal, shutting down...")
	}
	close(stopChan)

	// Give goroutines time to finish current operations
	time.Sleep(2 * time.Second)
	if config.full {
		fmt.Println("Performance test completed")
	}
}

func benchmarkPrimality(threadID int, stopChan <-chan struct{}, config Config, cpuStats *CPUStats) {
	if config.full {
		fmt.Printf("CPU Thread %d: Starting\n", threadID)
	}

	iteration := 0
	lastReport := time.Now()
	totalTime := time.Duration(0)

	for {
		select {
		case <-stopChan:
			if config.full {
				fmt.Printf("CPU Thread %d: Completed %d iterations\n", threadID, iteration)
			}
			return
		default:
			start := time.Now()
			primeCount := 0

			for i := 2; i < config.primeRange; i++ {
				if isPrime(i) {
					primeCount++
				}
			}

			duration := time.Since(start)
			iteration++
			totalTime += duration

			// Update shared stats for default (quiet) mode
			if !config.full {
				cpuStats.mu.Lock()
				cpuStats.totalTime += duration
				cpuStats.totalPrimesFound += primeCount
				shouldReport := time.Since(cpuStats.lastReport) >= time.Duration(config.reportInterval)*time.Second
				if shouldReport {
					// Calculate total primes/sec by multiplying average by number of threads
					avgPrimesPerSec := float64(cpuStats.totalPrimesFound) / cpuStats.totalTime.Seconds()
					totalPrimesPerSec := avgPrimesPerSec * float64(config.cpuThreads)
					cpuStats.lastReport = time.Now()
					cpuStats.mu.Unlock()

					fmt.Printf("CPU: %s total primes/sec\n", formatWithCommas(totalPrimesPerSec))
				} else {
					cpuStats.mu.Unlock()
				}
			} else {
				// Report at intervals for full mode
				if time.Since(lastReport) >= time.Duration(config.reportInterval)*time.Second {
					avgTime := totalTime / time.Duration(iteration)
					primesPerSec := float64(primeCount) / duration.Seconds()
					fmt.Printf("CPU Thread %d: %d iterations, avg %.2fms/iter, %s primes/sec\n",
						threadID, iteration, avgTime.Seconds()*1000, formatWithCommas(primesPerSec))
					lastReport = time.Now()
				}
			}
		}
	}
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func memoryAndFilesystemBenchmark(stopChan <-chan struct{}, config Config) {
	if config.full {
		fmt.Println("Memory: Starting allocation and filesystem benchmark")
	}

	// Allocate memory
	targetMemory := int64(float64(getAvailableMemory(config)) * config.memoryPercent)
	if config.full {
		fmt.Printf("Memory: Target allocation: %d MB\n", targetMemory/(1024*1024))
	}

	var memoryChunks [][]byte
	chunkSize := config.chunkSizeMB * 1024 * 1024
	allocated := int64(0)

	start := time.Now()
	for allocated < targetMemory {
		select {
		case <-stopChan:
			if config.full {
				fmt.Printf("Memory: Stopping allocation at %d MB\n", allocated/(1024*1024))
			}
			return
		default:
			chunk := make([]byte, chunkSize)
			// Fill with random data to ensure actual allocation
			for i := range chunk {
				chunk[i] = byte(i % 256)
			}
			memoryChunks = append(memoryChunks, chunk)
			allocated += int64(chunkSize)
		}
	}

	allocationDuration := time.Since(start)
	if config.full {
		fmt.Printf("Memory: Allocated %d MB in %v\n", allocated/(1024*1024), allocationDuration)
	}

	// Now benchmark filesystem using the allocated memory (continuous loop)
	filesystemBenchmark(memoryChunks, stopChan, config)
}

func getAvailableMemory(config Config) int64 {
	if runtime.GOOS == "linux" {
		return getLinuxMemory(config)
	} else if runtime.GOOS == "darwin" {
		return getDarwinMemory(config)
	}

	fmt.Println("Unsupported OS, using 8GB memory")
	// Fallback for other systems
	return 8 * 1024 * 1024 * 1024 // 8GB default
}

func getLinuxMemory(config Config) int64 {
	// Read /proc/meminfo to get actual available memory
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Println("Error reading /proc/meminfo", err)
		return 8 * 1024 * 1024 * 1024 // 8GB default
	}

	lines := strings.Split(string(data), "\n")
	var memAvailable int64

	for _, line := range lines {
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					memAvailable = kb * 1024 // Convert KB to bytes
					break
				}
			}
		}
	}

	// If MemAvailable is not found or is 0, fall back to MemFree + Buffers + Cached
	if memAvailable == 0 {
		var memFree, buffers, cached int64

		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					continue
				}

				switch {
				case strings.HasPrefix(line, "MemFree:"):
					memFree = kb * 1024
				case strings.HasPrefix(line, "Buffers:"):
					buffers = kb * 1024
				case strings.HasPrefix(line, "Cached:"):
					cached = kb * 1024
				}
			}
		}

		memAvailable = memFree + buffers + cached
	}

	// If still 0 or negative, use default
	if memAvailable <= 0 {
		fmt.Println("Failed to find available memory, using 8GB memory")
		return 8 * 1024 * 1024 * 1024 // 8GB default
	}

	if config.full {
		fmt.Println("Found available memory:", memAvailable)
	}
	return memAvailable
}

func getDarwinMemory(config Config) int64 {
	// Use vm_stat command to get memory information on macOS
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running vm_stat:", err)
		return 8 * 1024 * 1024 * 1024 // 8GB default
	}

	lines := strings.Split(string(output), "\n")
	var pageSize, freePages, inactivePages int64

	// Get page size first
	for _, line := range lines {
		if strings.Contains(line, "page size of") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "of" && i+1 < len(fields) {
					size, err := strconv.ParseInt(fields[i+1], 10, 64)
					if err == nil {
						pageSize = size
						break
					}
				}
			}
		}
	}

	// Default page size if not found
	if pageSize == 0 {
		pageSize = 4096 // 4KB default page size
	}

	// Parse memory statistics
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			valueStr := strings.TrimSuffix(fields[len(fields)-1], ".")
			value, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				continue
			}

			if strings.HasPrefix(line, "Pages free:") {
				freePages = value
			} else if strings.HasPrefix(line, "Pages inactive:") {
				inactivePages = value
			}
		}
	}

	// Calculate available memory (free + inactive pages)
	availableMemory := (freePages + inactivePages) * pageSize

	// If calculation failed, use default
	if availableMemory <= 0 {
		fmt.Println("Failed to find available memory, using 8GB memory")
		return 8 * 1024 * 1024 * 1024 // 8GB default
	}

	if config.full {
		fmt.Println("Found available memory:", availableMemory)
	}
	return availableMemory
}

func filesystemBenchmark(memoryChunks [][]byte, stopChan <-chan struct{}, config Config) {
	if config.full {
		fmt.Printf("Disk: Starting filesystem benchmark in path: %s\n", config.diskPath)
	}

	if len(memoryChunks) == 0 {
		fmt.Println("Disk: No memory chunks available for filesystem test")
		return
	}

	// Create temporary file for benchmarking
	tempFile, err := os.CreateTemp(config.diskPath, "perf_test_*.tmp")
	if err != nil {
		fmt.Printf("Disk: Error creating temp file: %v\n", err)
		return
	}

	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("Disk: Error removing temp file: %v\n", err)
		}
	}(tempFile.Name())

	defer func(tempFile *os.File) {
		err := tempFile.Close()
		if err != nil {
			fmt.Printf("Disk: Error closing temp file: %v\n", err)
		}
	}(tempFile)

	iteration := 0
	lastReport := time.Now()
	totalWriteMBps := float64(0)
	totalReadMBps := float64(0)

	for {
		select {
		case <-stopChan:
			if config.full {
				fmt.Printf("Disk: Completed %d iterations\n", iteration)
			}
			return
		default:
			iteration++

			// Write benchmark
			_, err := tempFile.Seek(0, 0)
			if err != nil {
				fmt.Printf("Disk: Error seeking file: %v\n", err)
				return
			}
			err = tempFile.Truncate(0)
			if err != nil {
				fmt.Printf("Disk: Error truncating file: %v\n", err)
				return
			}

			writeStart := time.Now()
			totalBytesWritten := int64(0)

			for _, chunk := range memoryChunks {
				select {
				case <-stopChan:
					return
				default:
					// Fill chunk with random data
					_, err := rand.Read(chunk)
					if err != nil {
						return
					}

					n, err := tempFile.Write(chunk)
					if err != nil {
						fmt.Printf("Disk: Write error: %v\n", err)
						break
					}
					totalBytesWritten += int64(n)
				}
			}

			err = tempFile.Sync()
			if err != nil {
				fmt.Printf("Disk: Error syncing file: %v\n", err)
				return
			}
			writeDuration := time.Since(writeStart)
			writeMBps := float64(totalBytesWritten) / (1024 * 1024) / writeDuration.Seconds()
			totalWriteMBps += writeMBps

			// Read benchmark
			_, err = tempFile.Seek(0, 0)
			if err != nil {
				fmt.Printf("Disk: Error seeking file: %v\n", err)
				return
			}

			readStart := time.Now()
			totalBytesRead := int64(0)
			buffer := make([]byte, config.chunkSizeMB*1024*1024)

		readLoop:
			for {
				select {
				case <-stopChan:
					break readLoop
				default:
					n, err := tempFile.Read(buffer)
					if n == 0 {
						break readLoop
					}
					if err != nil && err.Error() != "EOF" {
						fmt.Printf("Disk: Read error: %v\n", err)
						break readLoop
					}
					totalBytesRead += int64(n)
				}
			}

			readDuration := time.Since(readStart)
			readMBps := float64(totalBytesRead) / (1024 * 1024) / readDuration.Seconds()
			totalReadMBps += readMBps

			// Report at intervals or every 5 iterations
			if time.Since(lastReport) >= time.Duration(config.reportInterval)*time.Second || iteration%5 == 0 {
				avgWriteMBps := totalWriteMBps / float64(iteration)
				avgReadMBps := totalReadMBps / float64(iteration)
				fmt.Printf("Disk: avg write %.2f MB/s, avg read %.2f MB/s\n", avgWriteMBps, avgReadMBps)
				lastReport = time.Now()
			}
		}
	}
}
