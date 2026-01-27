//go:build !linux && !windows

package gpu

import "errors"

type osmesaContext struct {
	width  int
	height int
}

func newOSMesaContext(_, _ int) (*osmesaContext, error) {
	return nil, errors.New("osmesa not supported")
}

func (c *osmesaContext) Resize(width, height int) error {
	return errors.New("osmesa not supported")
}

func (c *osmesaContext) Buffer() []byte {
	return nil
}

func (c *osmesaContext) Destroy() {}
