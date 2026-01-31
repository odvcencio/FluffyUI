#!/usr/bin/env bash
set -euo pipefail

# Basic allocation audit for core render paths.

go test ./widgets -run ^$ -bench Render -benchmem

go test ./runtime -run ^$ -bench Buffer -benchmem
