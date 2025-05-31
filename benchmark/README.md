# Benchmarks

This directory contains scripts/tools/infrastructure for benchmarking rad. The goal is to see changes in performance over time and being aware of improvements/regressions between versions.

## How To

We use [hyperfine](https://github.com/sharkdp/hyperfine) to conduct the benchmarks. [benchmark.rad](./benchmark.rad) can be invoked to run the benchmark suite & report. For example:

```
./benchmark.rad --mock-response ".*:report.json"
```

## Benchmark History

### rad version 0.5.15 (Commit 00d3b98) (2025-03-29)

*Full report: [reports/report-1743229305.json](reports/report-1743229305.json)*

```
Benchmark     Mean   Stddev  Min    Max    Runs 
concat        510.1  2.9     507.2  516.3  10    
for-loop-add  418.3  7.2     410.2  432.1  10    
math          597    21.8    576.1  638.3  10    
read-file     551.1  19.7    529.3  592.2  10    
Times in milliseconds
Apple M2 Pro (10 cores) 16 GB macOS 15.3.2
```

### rad version 0.5.15 (Commit a6bb304) (2025-03-29)

*Full report: [reports/report-1743228997.json](reports/report-1743228997.json)*

```
Benchmark     Mean   Stddev  Min    Max    Runs 
concat        369.7  1.5     367.7  372.8  10    
for-loop-add  285.6  7.8     280    301    10    
math          511.9  15.7    500.6  541    10    
read-file     406.3  8.5     401.2  429.9  10    
Times in milliseconds
Apple M2 Pro (10 cores) 16 GB macOS 15.3.2
```
