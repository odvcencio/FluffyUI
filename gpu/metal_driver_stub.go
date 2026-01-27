//go:build !darwin

package gpu

func newMetalDriver() (Driver, error) {
	return nil, ErrUnsupported
}
