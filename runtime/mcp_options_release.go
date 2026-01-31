//go:build release

package runtime

import "fmt"

func (o *MCPOptions) validateTestFlags() error {
	if o == nil {
		return nil
	}
	if o.TestBypassTextGating || o.TestBypassClipboardGating {
		return fmt.Errorf("test bypass flags cannot be used in release builds")
	}
	return nil
}
