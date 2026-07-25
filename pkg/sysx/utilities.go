package sysx

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/strutil"
)

// setenv sets the environment variable named by key to value.
//
// It delegates directly to os.Setenv and propagates any error.
//
// Parameters:
//   - `key`:   the name of the environment variable to set.
//   - `value`: the value to assign.
//
// Returns:
//
//	An error if the variable could not be set, or nil on success.
//
// Example:
//
//	if err := sysx.setenv("LOG_LEVEL", "debug"); err != nil {
//	    log.Fatal(err)
//	}
func setenv(key, value string) error {
	if strutil.IsEmpty(key) {
		return errors.New("sysx: key must not be empty")
	}
	return os.Setenv(key, value)
}

// linuxOSVersion retrieves the Linux distribution version from the standard
// /etc/os-release file.
//
// The function specifically looks for the "PRETTY_NAME" key, which provides a
// human-readable name for the distribution. If the file is inaccessible, or
// if the key is not found, it returns runtime.GOOS as a fallback.
//
// Returns:
//
//	A string identifying the Linux distribution and version, or "linux".
func linuxOSVersion() string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return runtime.GOOS
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "PRETTY_NAME="); ok {
			after = strings.Trim(after, `"`)
			if after != "" {
				return after
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return runtime.GOOS
	}
	return runtime.GOOS
}

// darwinOSVersion retrieves the macOS product version using the system's
// sw_vers utility.
//
// Security fix (CWE-426): The binary is invoked via its absolute path
// /usr/bin/sw_vers to prevent PATH-hijacking attacks. No shell is used;
// exec.Command passes arguments directly to the kernel without shell
// interpretation.
//
// Returns:
//
//	A string identifying the macOS product version (e.g. "14.2.1"), or "darwin".
func darwinOSVersion() string {
	// Use absolute path to prevent PATH-based search hijacking (CWE-426).
	out, err := exec.Command("/usr/bin/sw_vers", "-productVersion").Output()
	if err != nil {
		return runtime.GOOS
	}
	v := strings.TrimSpace(string(out))
	if strutil.IsEmpty(v) {
		return runtime.GOOS
	}
	return v
}

// windowsOSVersion retrieves the Windows product version using the system's
// cmd.exe and ver utility.
//
// Security fix (CWE-426): cmd.exe is referenced by its canonical absolute
// path %SystemRoot%\System32\cmd.exe (resolved via the SYSTEMROOT environment
// variable with a safe fallback) to prevent PATH-hijacking attacks.
//
// Returns:
//
//	A string identifying the Windows version, or "windows" on fallback.
func windowsOSVersion() string {
	// Security fix (CWE-426): Resolve cmd.exe from %SYSTEMROOT% (not PATH)
	// to prevent substitution of a malicious binary via PATH manipulation.
	systemRoot := os.Getenv("SYSTEMROOT")
	if strutil.IsEmpty(systemRoot) {
		systemRoot = `C:\Windows`
	}
	cmdExe := filepath.Join(systemRoot, "System32", "cmd.exe")
	out, err := exec.Command(cmdExe, "/c", "ver").Output()
	if err != nil {
		return runtime.GOOS
	}
	v := strings.TrimSpace(string(out))
	if strutil.IsEmpty(v) {
		return runtime.GOOS
	}
	// example output: Microsoft Windows [Version 10.0.19045.2965]
	return v
}

// getFileMutex retrieves the sync.Mutex associated with the given file path.
//
// It performs an atomic load-or-store operation on a global map of mutexes.
// If no mutex exists for the path, a new one is created and returned.
//
// Returns:
//
//	A non-nil *sync.Mutex unique to the provided path.
func getFileMutex(path string) *sync.Mutex {
	v, _ := fileMutexes.LoadOrStore(path, &sync.Mutex{})
	mu, ok := v.(*sync.Mutex)
	if !ok {
		// Should never happen: fileMutexes only stores *sync.Mutex values.
		mu = &sync.Mutex{}
		fileMutexes.Store(path, mu)
	}
	return mu
}

// parseBool parses a lowercase, trimmed string as a boolean.
//
// Recognized true  values: "1", "true", "yes", "on"
// Recognized false values: "0", "false", "no", "off"
//
// Returns (value, true) when the string is recognized, or (false, false) when
// it is not.
func parseBool(s string) (bool, bool) {
	if strutil.IsEmpty(s) {
		return false, false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	v, err := conv.Bool(s)
	if err != nil {
		return false, false
	}
	return v, true
}

// splitLines splits s into individual lines by "\n", stripping any trailing
// "\r" (to handle "\r\n" line endings). A single trailing newline is consumed
// so that "a\nb\n" returns ["a","b"] rather than ["a","b",""]. An empty input
// returns nil.
func splitLines(s string) []string {
	if strutil.IsEmpty(s) {
		return nil
	}
	s = strings.TrimRight(s, "\n")
	parts := strings.Split(s, "\n")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimRight(p, "\r")
	}
	return result
}
