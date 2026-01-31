#!/usr/bin/env bash
set -euo pipefail

exclude_pattern="${EXCLUDE_PATTERN:-/(third_party|examples|docs|scripts|tools|cmd)(/|$)}"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

merged="$tmpdir/overall.cover"
first=true

packages=$(go list ./... | rg -v "$exclude_pattern")

printf "Computing overall coverage (excluding %s)\n" "$exclude_pattern"

for pkg in $packages; do
	printf "Testing %s\n" "$pkg"
	outfile="$tmpdir/$(echo "$pkg" | tr '/.' '__').cover"
	logfile="$tmpdir/$(echo "$pkg" | tr '/.' '__').log"
	if ! go test "$pkg" -timeout 60s -coverprofile "$outfile" > "$logfile" 2>&1; then
		cat "$logfile" >&2
		exit 1
	fi
	if [ ! -s "$outfile" ]; then
		continue
	fi
	if $first; then
		cp "$outfile" "$merged"
		first=false
	else
		tail -n +2 "$outfile" >> "$merged"
	fi

done

if $first; then
	echo "No coverage data produced." >&2
	exit 1
fi

overall=$(go tool cover -func "$merged" | tail -n 1)
echo "$overall"
