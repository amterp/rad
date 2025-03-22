# Benchmarks

This directory contains scripts/tools/infrastructure for benchmarking rad. The goal is to see changes in performance over time and being aware of improvements/regressions between versions.

## How To

We use [hyperfine](https://github.com/sharkdp/hyperfine) to conduct the benchmarks. [benchmark.rsl](./benchmark.rsl) can be invoked to run the benchmark suite & report. For example:

```
./benchmark.rsl --mock-response ".*:report.json"
```

## Benchmark History

### rad version 0.5.15 (Commit 4d3bee5) (2025-03-22)

*Full report: [reports/report-1742614777.json](reports/report-1742614777.json)*

```
Benchmark     Mean   Stddev  Min    Max    Runs 
concat        370.6  6.3     365.8  387.6  10    
for-loop-add  565.8  17.1    551.9  604.1  10    
math          496.9  2       493.6  500.3  10    
read-file     406.8  1.5     403.9  408.6  10    
Times in milliseconds
Apple M2 Pro (10 cores) 16 GB macOS 15.3.2
```
