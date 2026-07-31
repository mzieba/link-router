//go:build !windows

package browsers

// detectPlatform has no platform-specific sources outside Windows; the false
// result tells Detect to fall back to the PATH lookup, which is how browsers on
// Linux and macOS are found.
func detectPlatform() ([]Browser, bool) { return nil, false }
