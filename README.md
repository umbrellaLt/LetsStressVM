# LetsStressVM

A lightweight Go application to stress test CPU and RAM on Linux. It allocates **100 MB of RAM per second** (holding it to prevent garbage collection) while simultaneously saturating every CPU core with floating-point math. All activity is logged to a file in real time.

---

## Prerequisites (Ubuntu 22.04)

### 1. Update package list

```bash
sudo apt update
```

### 2. Install Go

```bash
sudo apt install -y golang-go
```

Verify the installation:

```bash
go version
# Expected output: go version go1.18.x linux/amd64 (or newer)
```

> **Note:** If you need a newer version of Go than what apt provides, install it manually:
> ```bash
> wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
> sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
> echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
> source ~/.bashrc
> ```

### 3. Install Make (optional, for using the Makefile)

```bash
sudo apt install -y make
```

### 4. Install Git (if not already present)

```bash
sudo apt install -y git
```

---

## Installation

```bash
git clone https://github.com/umbrellaLt/LetsStressVM.git
cd LetsStressVM
```

---

## Build

**Using Make:**

```bash
make build
```

**Or manually:**

```bash
go build -o stress_test stress.go
```

---

## Usage

**Using Make:**

```bash
make run
```

**Or run the binary directly:**

```bash
./stress_test
```

Stop the test at any time with **Ctrl+C** — the application shuts down gracefully and prints a final summary.

---

## What it does

| Component | Behaviour |
|-----------|-----------|
| **RAM stress** | Allocates 100 MB every second and touches every memory page to force the OS to commit it. All chunks are held in memory for the duration of the test. |
| **CPU stress** | Spawns one goroutine per logical CPU core. Each goroutine runs a tight loop of `sqrt`, `sin`, and `cos` floating-point operations to fully saturate the core. |
| **Logging** | Writes timestamped output to `stress_test.log` alongside the binary. Each allocation event and a stats summary (RAM held, RSS, CPU ops/sec, GC runs) is recorded every second. |

---

## Example output

```
=== Stress Test Started at 2024-05-08T10:00:00Z ===
System: 4 CPUs detected | RAM target: 100MB/sec
Spawning 4 CPU stress goroutines...
[RAM]   Allocated +100MB  |  Total held:  100MB
[STATS] Elapsed: 1s       | RAM held:  100MB | RSS ~  312MB | CPU ops/s:  8450123 | GC runs: 0
[RAM]   Allocated +100MB  |  Total held:  200MB
[STATS] Elapsed: 2s       | RAM held:  200MB | RSS ~  420MB | CPU ops/s:  8312456 | GC runs: 0
...
```

---

## Log file

A log file named `stress_test.log` is created in the same directory as the binary. It captures the full session including start time, per-second stats, and the final summary with total RAM allocated.

---

## Cleanup

```bash
make clean
```

This removes the compiled binary and the log file.

---

## Warning

This tool is intended for controlled stress testing environments. Running it on a production system or a machine with limited RAM **will exhaust available memory** and may cause the OOM killer to terminate processes. Use with caution.
