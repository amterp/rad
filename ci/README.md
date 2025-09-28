# Rad CI System

This directory contains CI/CD scripts that run automated checks on pull requests, including testing and
binary size comparison.

## Overview

Our CI system uses **Rad scripts**, in part as a dog-fooding exercise.

GitHub Actions (see `.github/workflows/pr-checks.yml`) orchestrates the workflow,
which installs Rad from binary releases and runs PR checks such as tests,
binary size comparison, and benchmarking.

## Structure

- `benchmark-scripts/` - Performance benchmark scripts for CI
- `binary-size-compare.rad` - Compares binary sizes between PR and base branch
- `benchmark-compare.rad` - Compares benchmark performance between PR and base branch
- `pr-comment.rad` - Generates comprehensive PR comment with all check results
- `test-runner.rad` - Runs test suite for CI
- `install-rad.sh` - Installs Rad from GitHub releases

## Adding New Benchmarks

To add a new benchmark for CI:

1. Create a new `.rad` script in `benchmark-scripts/`
2. Add the script path to the `benchmarks` array in `benchmark-compare.rad`
3. The benchmark will automatically be included in PR checks
