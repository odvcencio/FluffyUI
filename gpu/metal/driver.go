package metal

import (
	"github.com/odvcencio/fluffyui/gpu"
)

// New returns a Metal driver when available.
func New() (gpu.Driver, error) {
	return nil, gpu.ErrUnsupported
}
