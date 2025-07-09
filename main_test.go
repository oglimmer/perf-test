package main

import (
	"os"
	"runtime"
	"testing"
)

func TestIsPrime(t *testing.T) {
	tests := []struct {
		input    int
		expected bool
	}{
		{-1, false},
		{0, false},
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{6, false},
		{7, true},
		{8, false},
		{9, false},
		{10, false},
		{11, true},
		{13, true},
		{15, false},
		{17, true},
		{25, false},
		{29, true},
		{100, false},
		{101, true},
		{997, true},
		{1000, false},
	}

	for _, test := range tests {
		result := isPrime(test.input)
		if result != test.expected {
			t.Errorf("isPrime(%d) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestFormatWithCommas(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0"},
		{123, "123"},
		{1000, "1,000"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{12345678, "12,345,678"},
		{123456789, "123,456,789"},
		{1234567890, "1,234,567,890"},
		{999, "999"},
		{9999, "9,999"},
		{99999, "99,999"},
		{999999, "999,999"},
	}

	for _, test := range tests {
		result := formatWithCommas(test.input)
		if result != test.expected {
			t.Errorf("formatWithCommas(%g) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGetAvailableMemory(t *testing.T) {
	config := Config{full: false}

	// Test that it returns a positive value
	memory := getAvailableMemory(config)
	if memory <= 0 {
		t.Errorf("getAvailableMemory() returned %d, expected positive value", memory)
	}

	// Test that it returns a reasonable value (at least 1MB, at most 1TB)
	minMemory := int64(1024 * 1024)               // 1MB
	maxMemory := int64(1024 * 1024 * 1024 * 1024) // 1TB

	if memory < minMemory {
		t.Errorf("getAvailableMemory() returned %d, expected at least %d", memory, minMemory)
	}

	if memory > maxMemory {
		t.Errorf("getAvailableMemory() returned %d, expected at most %d", memory, maxMemory)
	}
}

func TestGetLinuxMemory(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux platform")
	}

	config := Config{full: false}
	memory := getLinuxMemory(config)

	if memory <= 0 {
		t.Errorf("getLinuxMemory() returned %d, expected positive value", memory)
	}
}

func TestGetDarwinMemory(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping Darwin-specific test on non-Darwin platform")
	}

	config := Config{full: false}
	memory := getDarwinMemory(config)

	if memory <= 0 {
		t.Errorf("getDarwinMemory() returned %d, expected positive value", memory)
	}
}

func TestConfigValidation(t *testing.T) {
	// Test memory percent validation bounds
	tests := []struct {
		memPercent float64
		shouldFail bool
	}{
		{0.05, true},  // Too low
		{0.1, false},  // Valid minimum
		{0.5, false},  // Valid middle
		{0.95, false}, // Valid maximum
		{0.96, true},  // Too high
		{1.0, true},   // Too high
		{-0.1, true},  // Negative
	}

	for _, test := range tests {
		config := Config{
			memoryPercent:  test.memPercent,
			primeRange:     1000,
			chunkSizeMB:    100,
			reportInterval: 5,
			cpuThreads:     1,
		}

		// This simulates the validation logic from main()
		isValid := config.memoryPercent >= 0.1 && config.memoryPercent <= 0.95

		if isValid == test.shouldFail {
			t.Errorf("Memory percent %f validation failed: expected shouldFail=%v, got isValid=%v",
				test.memPercent, test.shouldFail, isValid)
		}
	}
}

func TestCPUThreadsCalculation(t *testing.T) {
	cpuCores := runtime.NumCPU()

	// Test auto-calculation (cpuThreads = 0)
	config := Config{cpuThreads: 0}

	// Simulate the logic from main()
	if config.cpuThreads == 0 {
		config.cpuThreads = cpuCores - 1
		if config.cpuThreads < 1 {
			config.cpuThreads = 1
		}
	}

	if config.cpuThreads < 1 {
		t.Errorf("CPU threads calculation resulted in %d, expected at least 1", config.cpuThreads)
	}

	if cpuCores > 1 && config.cpuThreads != cpuCores-1 {
		t.Errorf("CPU threads calculation resulted in %d, expected %d (cores-1)", config.cpuThreads, cpuCores-1)
	}

	if cpuCores == 1 && config.cpuThreads != 1 {
		t.Errorf("CPU threads calculation resulted in %d, expected 1 for single-core system", config.cpuThreads)
	}
}

func BenchmarkIsPrime(b *testing.B) {
	// Benchmark isPrime function with various inputs
	primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, prime := range primes {
			isPrime(prime)
		}
	}
}

func BenchmarkFormatWithCommas(b *testing.B) {
	testValues := []float64{123, 1234, 12345, 123456, 1234567, 12345678, 123456789, 1234567890}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testValues {
			formatWithCommas(val)
		}
	}
}

// Test helper to check if temp directory is writable
func TestTempDirWritable(t *testing.T) {
	tempDir := os.TempDir()

	// Try to create a temp file
	tempFile, err := os.CreateTemp(tempDir, "test_*.tmp")
	if err != nil {
		t.Errorf("Cannot create temp file in %s: %v", tempDir, err)
		return
	}

	// Clean up
	tempFile.Close()
	os.Remove(tempFile.Name())
}
