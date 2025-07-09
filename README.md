# Performance Test Tool

A comprehensive system performance benchmarking tool written in Go that tests CPU, memory, and disk I/O performance.

## Features

- **CPU Benchmarking**: Multi-threaded prime number calculation with configurable thread count
- **Memory Benchmarking**: Allocates and manages memory chunks to test RAM performance
- **Disk I/O Benchmarking**: Tests filesystem read/write performance using temporary files
- **Configurable Parameters**: Customize test parameters via command-line flags
- **Selective Testing**: Disable CPU or disk testing with command-line flags
- **Dual Output Modes**: Quiet mode (default) for essential metrics, full mode for detailed output
- **Graceful Shutdown**: Handles interrupt signals cleanly

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/your-username/perf-test/releases).

### Build from Source

```bash
git clone https://github.com/your-username/perf-test.git
cd perf-test
go build -o perf-test main.go
```

## Usage

### Basic Usage

Run with default settings (quiet mode):
```bash
./perf-test
```

### Command Line Options

```bash
./perf-test [options]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-prime-range` | 10000000 | Range for prime number testing |
| `-memory-percent` | 0.9 | Percentage of memory to allocate (0.1-0.95) |
| `-chunk-size` | 100 | Memory chunk size in MB |
| `-report-interval` | 5 | Seconds between benchmark reports |
| `-cpu-threads` | 0 | Number of CPU threads (0 = auto: cores-1) |
| `-full` | false | Show full output with detailed information |
| `-disable-cpu` | false | Disable CPU testing |
| `-disable-disk` | false | Disable disk testing |

### Examples

**Run with full detailed output:**
```bash
./perf-test -full
```

**Custom configuration:**
```bash
./perf-test -prime-range 5000000 -memory-percent 0.8 -cpu-threads 4 -full
```

**Light memory usage test:**
```bash
./perf-test -memory-percent 0.3 -chunk-size 50
```

**Run only CPU testing:**
```bash
./perf-test -disable-disk
```

**Run only disk testing:**
```bash
./perf-test -disable-cpu
```

## Output Modes

### Quiet Mode (Default)
- Shows only essential performance metrics
- CPU: Aggregated primes/sec across all threads
- Disk: Average read/write speeds

### Full Mode (`-full` flag)
- Detailed startup information
- Per-thread CPU performance metrics
- Memory allocation progress
- Filesystem benchmark details
- Shutdown messages

## Sample Output

### Quiet Mode
```
CPU: 1250000 total primes/sec
Disk: avg write 245.67 MB/s, avg read 412.33 MB/s
```

### Full Mode
```
CPU cores detected: 8
Using 7 threads for CPU benchmarking
Prime range: 10000000
Memory allocation: 90%
Chunk size: 100 MB
Report interval: 5 seconds
CPU Thread 0: Starting
CPU Thread 1: Starting
...
Memory: Starting allocation and filesystem benchmark
Memory: Target allocation: 7372 MB
Memory: Allocated 7372 MB in 2.3s
Disk: Starting filesystem benchmark
CPU Thread 0: 3 iterations, avg 1.67s/iter, 598743 primes/sec
Disk: avg write 245.67 MB/s, avg read 412.33 MB/s
```

## System Requirements

- Go 1.19+ (for building from source)
- Sufficient RAM for memory benchmarks (respects `-memory-percent` setting)
- Write permissions in current directory (for temporary files)

## Performance Notes

- The tool automatically detects CPU cores and uses cores-1 threads by default
- Memory allocation is gradual to avoid system instability
- Disk benchmarks use temporary files that are automatically cleaned up
- Use Ctrl+C for graceful shutdown

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Changelog

### v1.0.0
- Initial release
- CPU, memory, and disk benchmarking
- Quiet mode as default with optional full mode
- Multi-threaded CPU testing
- Graceful shutdown handling