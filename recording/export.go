package recording

import (
	"fmt"
	"os/exec"
	"strconv"
)

// AggOptions configures asciinema-agg exports.
type AggOptions struct {
	Theme    string
	FontSize int
	Speed    float64
	FPS      int
}

// ExportWithAgg uses the external "agg" tool to render asciicast output.
// The output format is inferred from the output file extension.
func ExportWithAgg(inputPath, outputPath string, options AggOptions) error {
	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("input and output paths are required")
	}
	agg, err := exec.LookPath("agg")
	if err != nil {
		return fmt.Errorf("agg not found: %w", err)
	}
	args := []string{inputPath, "-o", outputPath}
	if options.Theme != "" {
		args = append(args, "--theme", options.Theme)
	}
	if options.FontSize > 0 {
		args = append(args, "--font-size", strconv.Itoa(options.FontSize))
	}
	if options.Speed > 0 {
		args = append(args, "--speed", fmt.Sprintf("%.2f", options.Speed))
	}
	if options.FPS > 0 {
		args = append(args, "--fps", strconv.Itoa(options.FPS))
	}
	cmd := exec.Command(agg, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("agg failed: %w (%s)", err, string(output))
	}
	return nil
}
