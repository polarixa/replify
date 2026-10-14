//go:build !windows
// +build !windows

package explorer

// StartedByExplorer returns true if the program was invoked by the user
// double-clicking on the executable from explorer.exe.
// On non-Windows platforms, it always returns false.
func StartedByExplorer() bool {
	return false
}
