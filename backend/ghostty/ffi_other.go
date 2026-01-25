//go:build !linux && !darwin && !windows

package ghostty

func defaultLibName() string {
	return ""
}
