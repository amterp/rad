# Rad CI System

This directory contains CI/CD scripts that run automated checks on pull requests, including testing and
binary size comparison.

## Overview

Our CI system uses **Rad scripts**, in part as a dog-fooding exercise.

GitHub Actions (see `.github/workflows/pr-checks.yml`) orchestrates the workflow,
which installs Rad from binary releases and runs PR checks such as tests,
binary size comparison, and benchmarking.
