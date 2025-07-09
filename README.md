# Performance Test Tool

A comprehensive system performance benchmarking tool written in Go that tests CPU, memory, and disk I/O performance.

## Features

- **CPU Benchmarking**: Multi-threaded prime number calculation with configurable thread count
- **Disk I/O Benchmarking**: Tests filesystem read/write performance using temporary files

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

## System Requirements

- Go 1.19+ (for building from source)

## License

MIT License - see [LICENSE](LICENSE) file for details.

