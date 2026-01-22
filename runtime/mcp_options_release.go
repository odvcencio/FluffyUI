//go:build release

package runtime

func (o *MCPOptions) validateTestFlags() {
	if o == nil {
		return
	}
	if o.TestBypassTextGating || o.TestBypassClipboardGating {
		panic("test bypass flags cannot be used in release builds")
	}
}
