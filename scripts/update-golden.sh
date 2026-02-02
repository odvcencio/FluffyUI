#!/bin/bash
# update-golden.sh - Regenerate golden snapshot files for widget tests
#
# Usage:
#   ./scripts/update-golden.sh              # Update all golden files
#   ./scripts/update-golden.sh TestGolden_Button  # Update specific test
#
# Golden files are stored in widgets/testdata/*.golden
#
# When intentional changes are made to widget rendering:
# 1. Run this script to regenerate golden files
# 2. Review the diff carefully (git diff widgets/testdata/)
# 3. Commit the updated golden files with your changes

set -e

cd "$(dirname "$0")/.."

if [ -n "$1" ]; then
    echo "Updating golden files for: $1"
    FLUFFYUI_UPDATE_SNAPSHOTS=1 go test ./widgets/... -run "$1" -v
else
    echo "Updating all golden files..."
    FLUFFYUI_UPDATE_SNAPSHOTS=1 go test ./widgets/... -run TestGolden -v
fi

echo ""
echo "Golden files updated. Review changes with:"
echo "  git diff widgets/testdata/"
