package metal

import (
	"m31labs.dev/fluffyui/gpu"
)

// New returns a Metal driver when available.
func New() (gpu.Driver, error) {
	return nil, gpu.ErrUnsupported
}
