#!/usr/bin/env bash
set -euo pipefail

widgets_threshold=60
runtime_threshold=75

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

printf "Running widgets coverage...\n"
go test ./widgets -coverprofile "$tmpdir/widgets.cover" > /dev/null
widgets_cov=$(go tool cover -func "$tmpdir/widgets.cover" | awk '/^total:/{print $3}' | sed 's/%//')

printf "Running runtime coverage...\n"
go test ./runtime -coverprofile "$tmpdir/runtime.cover" > /dev/null
runtime_cov=$(go tool cover -func "$tmpdir/runtime.cover" | awk '/^total:/{print $3}' | sed 's/%//')

check_threshold() {
	local name="$1"
	local cov="$2"
	local threshold="$3"
	if awk -v cov="$cov" -v th="$threshold" 'BEGIN { exit (cov+0 < th+0) }'; then
		printf "%s coverage %s%% (threshold %s%%)\n" "$name" "$cov" "$threshold"
	else
		printf "%s coverage %s%% is below threshold %s%%\n" "$name" "$cov" "$threshold" >&2
		exit 1
	fi
}

check_threshold "widgets" "$widgets_cov" "$widgets_threshold"
check_threshold "runtime" "$runtime_cov" "$runtime_threshold"
