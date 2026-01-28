#!/bin/bash
# regenerate-demos.sh - Regenerate all FluffyUI demos

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEMOS_DIR="$PROJECT_ROOT/docs/demos"
DURATION="${1:-6}"

echo "========================================="
echo "FluffyUI Demo Regeneration"
echo "========================================="
echo "Duration: ${DURATION}s per demo"
echo "Output: $DEMOS_DIR"
echo ""

# Check dependencies
if ! command -v agg &> /dev/null; then
    echo "Warning: agg not found. GIFs will not be generated."
    echo "Install with: cargo install --git https://github.com/asciinema/agg"
    echo ""
fi

# Generate cast files
echo "Step 1: Generating cast files..."
cd "$PROJECT_ROOT"
go run ./examples/generate-demos --out "$DEMOS_DIR" --duration "$DURATION"
echo ""

# Convert to GIF if agg is available
if command -v agg &> /dev/null; then
    echo "Step 2: Converting to GIFs..."
    cd "$DEMOS_DIR"
    
    for cast in *.cast; do
        if [[ "$cast" == "candy-wars.cast" ]]; then
            # Skip candy-wars.cast as it may be corrupted
            echo "Skipping: $cast (may need agent recording)"
            continue
        fi
        
        gif="${cast%.cast}.gif"
        if [[ -f "$gif" ]]; then
            echo "Skipping: $gif (already exists)"
        else
            echo "Converting: $cast -> $gif"
            agg --theme monokai --font-size 16 --fps-cap 30 \
                --last-frame-duration 0.001 "$cast" "$gif" 2>&1 || echo "Failed: $cast"
        fi
    done
else
    echo "Step 2: Skipped (agg not installed)"
fi

echo ""
echo "========================================="
echo "Demo Regeneration Complete!"
echo "========================================="
echo ""
echo "Generated files:"
ls -lh "$DEMOS_DIR"/*.cast 2>/dev/null | wc -l | xargs echo "  Cast files:"
ls -lh "$DEMOS_DIR"/*.gif 2>/dev/null | wc -l | xargs echo "  GIF files:"
echo ""
echo "Total size:"
du -sh "$DEMOS_DIR/"
echo ""
echo "To view a demo:"
echo "  asciinema play $DEMOS_DIR/hero.cast"
echo ""
echo "To view a GIF:"
echo "  open $DEMOS_DIR/hero.gif"
