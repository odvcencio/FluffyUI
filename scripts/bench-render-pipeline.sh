#!/usr/bin/env bash
set -euo pipefail

go test ./runtime -run '^$' -bench RenderPipelineDepth -benchmem "$@"

