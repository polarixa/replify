package sysx

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// FileExists reports whether a file exists at the given path.
//
// The function returns true for any path that exists in the file system,
// including directories and symbolic links. Use IsFile to restrict the check
// to regular files.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when any file system entry exists at path;
//	 - false when the path does not exist or the existence cannot be determined.
//
// Example:
//
//	if sysx.FileExists("/etc/hosts") {
//	    fmt.Println("hosts file found")
//	}
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists reports whether a directory exists at the given path.
//
// The function returns true only when the path exists and refers to a
// directory (not a regular file or other entry).
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when a directory exists at path;
//	 - false otherwise.
//
// Example:
//
//	if sysx.DirExists("/tmp") {
//	    fmt.Println("tmp dir exists")
//	}
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// IsFile reports whether the given path exists and is a regular file.
//
// Symbolic links are followed; if the link target is a regular file this
// function returns true.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when path exists and is a regular file;
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsFile("/etc/passwd") {
//	    fmt.Println("regular file")
//	}
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

// IsDir reports whether the given path exists and is a directory.
//
// Symbolic links are followed; if the link target is a directory this
// function returns true.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when path exists and is a directory;
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsDir("/home") {
//	    fmt.Println("is a directory")
//	}
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// IsSymlink reports whether the given path is a symbolic link.
//
// Unlike IsFile and IsDir, this function does NOT follow symbolic links; it
// inspects the link entry itself using os.Lstat.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when path exists and is a symbolic link;
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsSymlink("/usr/bin/python") {
//	    fmt.Println("python is a symlink")
//	}
func IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// IsExecutable reports whether the file at the given path is executable by
// its owner.
//
// The check is based on file mode bits (0100). On Windows, it verifies
// if the file ends with a known executable extension or matches the PATHEXT
// environment variable.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when the file exists and its owner-execute bit is set (or is an executable on Windows);
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsExecutable("/usr/bin/git") {
//	    fmt.Println("git is executable")
//	}
func IsExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if IsWindows() {
		if !fi.Mode().IsRegular() {
			return false
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".exe", ".bat", ".cmd", ".com", ".ps1":
			return true
		default:
			if ext == "" {
				return false
			}
			pathext := os.Getenv("PATHEXT")
			if pathext != "" {
				for _, pe := range strings.Split(pathext, string(os.PathListSeparator)) {
					if strings.EqualFold(ext, strings.TrimSpace(pe)) {
						return true
					}
				}
			}
			return false
		}
	}
	return fi.Mode()&0100 != 0
}

// IsReadable reports whether the file at the given path is readable by its
// owner.
//
// The check is based on file mode bits (0400). On Windows, mode bits are an
// approximation.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when the file exists and its owner-read bit is set;
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsReadable("/etc/hosts") {
//	    fmt.Println("readable")
//	}
func IsReadable(path string) bool {
	// On Windows, mode bits are unreliable. Probe by attempting to open
	// the file for reading — the same approach IsWritable uses for writes.
	if IsWindows() {
		f, err := os.Open(path)
		if err != nil {
			return false
		}
		defer f.Close()
		return true
	}
	fi, err := os.Stat(path)
	return err == nil && fi.Mode()&0400 != 0
}

// IsWritable reports whether the file at the given path can be opened for
// writing by the current process.
//
// The check attempts to open the file with os.O_WRONLY. This approach
// respects ACLs and is more accurate than mode-bit inspection alone.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	A boolean value:
//	 - true  when the file exists and can be opened for writing;
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsWritable("/tmp/out.log") {
//	    fmt.Println("writable")
//	}
func IsWritable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	defer f.Close()
	return true
}

// IsBinary reports whether the file at the given path appears to be a binary
// file.
//
// Detection is performed using a heuristic: the first 8 KiB of the file are
// searched for a null byte (0x00). If one is found, the file is considered
// binary. An empty file is not considered binary.
//
// Parameters:
//   - `path`: the file system path to check.
//
// Returns:
//
//	(bool, error): true if the file is binary and nil on success, or false
//	and a non-nil error if the file cannot be opened or read.
//
// Example:
//
//	isBin, err := sysx.IsBinary("/usr/bin/ls")
//	if err == nil && isBin {
//	    fmt.Println("binary file detected")
//	}
func IsBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read first 8KB
	buf := make([]byte, 8192)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}

// FileSize returns the size of the file at the given path in bytes.
//
// Parameters:
//   - `path`: the file system path to inspect.
//
// Returns:
//
//	(int64, error): the file size in bytes and nil on success, or 0 and a
//	non-nil error if the file does not exist or cannot be stat'd.
//
// Example:
//
//	size, err := sysx.FileSize("/var/log/syslog")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("%d bytes\n", size)
func FileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// FileMode returns the permission bits (os.FileMode) of the file at the given
// path.
//
// Symbolic links are followed; the mode of the link target is returned.
//
// Parameters:
//   - `path`: the file system path to inspect.
//
// Returns:
//
//	(os.FileMode, error): the file permission bits and nil on success, or 0
//	and a non-nil error if the file does not exist or cannot be stat'd.
//
// Example:
//
//	mode, err := sysx.FileMode("/etc/passwd")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("%o\n", mode) // e.g. "644"
func FileMode(path string) (os.FileMode, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Mode().Perm(), nil
}

// FileModTime returns the modification time of the file at the given path.
//
// Symbolic links are followed; the modification time of the link target is
// returned.
//
// Parameters:
//   - `path`: the file system path to inspect.
//
// Returns:
//
//	(time.Time, error): the modification time and nil on success, or the zero
//	time and a non-nil error if the file does not exist or cannot be stat'd.
//
// Example:
//
//	t, err := sysx.FileModTime("/var/log/syslog")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(t.Format(time.RFC3339))
func FileModTime(path string) (time.Time, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// TempDir returns the default directory to use for temporary files.
//
// It delegates directly to os.TempDir. On Unix, it returns $TMPDIR if set,
// else /tmp. On Windows it returns %TMP%, %TEMP%, or %USERPROFILE%.
//
// Returns:
//
//	A non-empty string containing the temporary directory path.
//
// Example:
//
//	fmt.Println(sysx.TempDir()) // "/tmp"
func TempDir() string {
	return os.TempDir()
}

// HomeDir returns the current user's home directory.
//
// The lookup is performed via os/user.Current(). On most systems this
// respects the HOME environment variable.
//
// Returns:
//
//	(string, error): the home directory path and nil on success, or an empty
//	string and a non-nil error if the current user cannot be determined.
//
// Example:
//
//	home, err := sysx.HomeDir()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(home)
func HomeDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}

// MustHomeDir returns the current user's home directory and panics if the
// lookup fails.
//
// Returns:
//
//	A non-empty string containing the home directory path.
//
// Example:
//
//	fmt.Println(sysx.MustHomeDir())
func MustHomeDir() string {
	h, err := HomeDir()
	if err != nil {
		panic("sysx: cannot retrieve home directory: " + err.Error())
	}
	return h
}

// WorkingDir returns the current working directory of the process.
//
// It delegates directly to os.Getwd.
//
// Returns:
//
//	(string, error): the working directory path and nil on success, or an
//	empty string and a non-nil error on failure.
//
// Example:
//
//	wd, err := sysx.WorkingDir()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(wd)
func WorkingDir() (string, error) {
	return os.Getwd()
}

// MustWorkingDir returns the current working directory of the process and
// panics if the lookup fails.
//
// Returns:
//
//	A non-empty string containing the current working directory path.
//
// Example:
//
//	fmt.Println(sysx.MustWorkingDir())
func MustWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic("sysx: cannot retrieve working directory: " + err.Error())
	}
	return wd
}

// Move renames (moves) src to dst. If dst already exists, it is overwritten.
//
// Move is more robust than os.Rename: it attempts an atomic rename first,
// but if that fails because src and dst are on different logical devices
// (cross-device link error), it falls back to a manual copy-and-delete
// strategy.
//
// Parameters:
//   - `src`: the source file or directory path.
//   - `dst`: the destination path.
//
// Returns:
//
//	An error if the move or fallback copy fails; nil on success.
//
// Example:
//
//	if err := sysx.Move("/tmp/data.txt", "/home/user/data.txt"); err != nil {
//	    log.Fatal(err)
//	}
func Move(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fallback for cross-device rename
	if IsDir(src) {
		if err := CopyDir(src, dst); err != nil {
			return err
		}
	} else {
		if err := CopyFile(src, dst); err != nil {
			return err
		}
	}
	return os.RemoveAll(src)
}

// Touch creates the file at path if it does not exist, or updates its access
// and modification times to the current time if it does.
//
// Parameters:
//   - `path`: the file system path to touch.
//
// Returns:
//
//	An error if the file could not be created or its times updated; nil on success.
//
// Example:
//
//	if err := sysx.Touch("/tmp/marker.lock"); err != nil {
//	    log.Fatal(err)
//	}
func Touch(path string) error {
	if !FileExists(path) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		return f.Close()
	}
	now := time.Now()
	return os.Chtimes(path, now, now)
}

// FileMD5 calculates the MD5 checksum of the file at the given path and
// returns it as a lowercase hexadecimal string.
//
// Parameters:
//   - `path`: the file system path to hash.
//
// Returns:
//
//	(string, error): the MD5 hash and nil on success, or an empty string and
//	a non-nil error if the file cannot be opened or read.
//
// Example:
//
//	hash, err := sysx.FileMD5("/etc/hosts")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(hash)
func FileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// FileSHA256 calculates the SHA256 checksum of the file at the given path and
// returns it as a lowercase hexadecimal string.
//
// Parameters:
//   - `path`: the file system path to hash.
//
// Returns:
//
//	(string, error): the SHA256 hash and nil on success, or an empty string
//	and a non-nil error if the file cannot be opened or read.
//
// Example:
//
//	hash, err := sysx.FileSHA256("/etc/hosts")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(hash)
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CopyFile copies the contents of the file at src to a new or truncated file
// at dst. The destination file is created with permission 0644 if it does not
// exist, or truncated if it does. The copy is performed using io.Copy with a
// buffered writer for efficiency.
//
// Parameters:
//   - `src`: the source file path to read from.
//   - `dst`: the destination file path to write to.
//
// Returns:
//
//	An error if the source cannot be opened, the destination cannot be
//	created, or the copy fails; nil on success.
//
// Example:
//
//	if err := sysx.CopyFile("/etc/hosts", "/tmp/hosts.bak"); err != nil {
//	    log.Fatal(err)
//	}
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	if _, err := io.Copy(w, in); err != nil {
		return err
	}
	return w.Flush()
}

// TruncateFile truncates or extends the file at path to the given size in
// bytes. If size is greater than the current file size, the file is extended
// with zero bytes. If the file does not exist, an error is returned.
//
// Parameters:
//   - `path`: the file system path to truncate.
//   - `size`: the desired file size in bytes; must be non-negative.
//
// Returns:
//
//	An error if the file does not exist or the truncation fails; nil on success.
//
// Example:
//
//	if err := sysx.TruncateFile("/tmp/output.bin", 0); err != nil {
//	    log.Fatal(err)
//	}
func TruncateFile(path string, size int64) error {
	return os.Truncate(path, size)
}

// RemoveFileIfExist removes the file at path if it exists.
//
// Parameters:
//   - `path`: the file path to remove.
//
// Returns:
//
//	An error if the file could not be removed; nil on success or if the
//	file does not exist.
//
// Example:
//
//	if err := sysx.RemoveFileIfExist("/tmp/app/cache/file.txt"); err != nil {
//	    log.Fatal(err)
//	}
func RemoveFileIfExist(path string) error {
	if FileExists(path) {
		return os.Remove(path)
	}
	return nil
}

// HumanizeBytes converts a byte count into a human-readable string using
// binary prefixes (KiB, MiB, GiB, etc.). It rounds to one decimal place.
//
// Parameters:
//   - `size`: the size in bytes to humanize.
//
// Returns:
//
//	A string representing the size in a human-readable format.
//
// Example:
//
//	fmt.Println(sysx.HumanizeBytes(1536)) // "1.5 KiB"
func HumanizeBytes(size int64) string {
	if size < 0 {
		return "0 B"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
