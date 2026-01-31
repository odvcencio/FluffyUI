//go:build !release

package runtime

func (o *MCPOptions) validateTestFlags() error {
	return nil
}
