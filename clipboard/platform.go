package clipboard

// PlatformClipboard uses platform-native command-line tools to access the
// system clipboard. It auto-detects the available clipboard tool at
// construction time.
//
// Supported tools by platform:
//   - Linux: xclip, xsel, wl-copy/wl-paste (Wayland)
//   - macOS: pbcopy/pbpaste
//   - Windows: PowerShell Set-Clipboard/Get-Clipboard
type PlatformClipboard struct {
	readCmd  []string // command + args for reading clipboard
	writeCmd []string // command + args for writing clipboard
	tool     string   // name of the detected tool for diagnostics
}
