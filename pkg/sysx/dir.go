package sysx

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// CreateDir creates the named directory, along with any necessary parent
// directories, using permission 0755.
//
// It is equivalent to os.MkdirAll(path, 0755) and is idempotent: calling it
// on a directory that already exists is a no-op.
//
// Parameters:
//   - `path`: the directory path to create.
//
// Returns:
//
//	An error if the directory could not be created; nil on success.
//
// Example:
//
//	if err := sysx.CreateDir("/tmp/app/logs"); err != nil {
//	    log.Fatal(err)
//	}
func CreateDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// CreateDirs creates multiple directories with the given permission.
//
// It is equivalent to os.MkdirAll for each directory and is idempotent: calling it
// on a directory that already exists is a no-op.
//
// Parameters:
//   - `perm`: the permission to use for creating the directories.
//   - `paths`: the directory paths to create.
//
// Returns:
//
//	An error if any of the directories could not be created; nil on success.
//
// Example:
//
//	if err := sysx.CreateDirs(0o755, "/tmp/app/logs", "/tmp/app/cache"); err != nil {
//	    log.Fatal(err)
//	}
func CreateDirs(perm fs.FileMode, paths ...string) error {
	for _, path := range paths {
		if strutil.IsEmpty(path) {
			continue
		}
		if err := os.MkdirAll(path, perm); err != nil {
			return err
		}
	}
	return nil
}

// CreateSubDirs creates subdirectories under a parent directory with the given permission.
//
// It is equivalent to os.MkdirAll for each subdirectory and is idempotent: calling it
// on a directory that already exists is a no-op.
//
// Parameters:
//   - `perm`: the permission to use for creating the directories.
//   - `parentDir`: the parent directory path.
//   - `subDirs`: the subdirectory names to create.
//
// Returns:
//
//	An error if any of the subdirectories could not be created; nil on success.
//
// Example:
//
//	if err := sysx.CreateSubDirs(0o755, "/tmp/app", "logs", "cache"); err != nil {
//	    log.Fatal(err)
//	}
func CreateSubDirs(perm fs.FileMode, parent string, subdirs ...string) error {
	if strutil.IsEmpty(parent) {
		return errors.New("parent directory is empty")
	}
	for _, dir := range subdirs {
		if strutil.IsEmpty(dir) {
			continue
		}
		fpath := filepath.Join(parent, dir)
		if err := os.MkdirAll(fpath, perm); err != nil {
			return err
		}
	}
	return nil
}

// RemoveDir removes the directory at path together with all of its contents.
//
// It wraps os.RemoveAll. Calling RemoveDir on a path that does not exist is
// not an error; the function returns nil in that case.
//
// Parameters:
//   - `path`: the directory path to remove.
//
// Returns:
//
//	An error if the directory could not be removed; nil on success or if the
//	path does not exist.
//
// Example:
//
//	if err := sysx.RemoveDir("/tmp/app/cache"); err != nil {
//	    log.Fatal(err)
//	}
func RemoveDir(path string) error {
	return os.RemoveAll(path)
}

// ListDir returns the names of all entries (files, directories, symlinks) in
// the directory at path, in the order returned by os.ReadDir (lexicographic
// order). Only entry names are returned; use filepath.Join to build full paths.
//
// Parameters:
//   - `path`: the directory path to read.
//
// Returns:
//
//	([]string, error): the entry names and nil on success, or nil and a
//	non-nil error if the directory cannot be read.
//
// Example:
//
//	names, err := sysx.ListDir("/etc")
//	for _, name := range names {
//	    fmt.Println(name)
//	}
func ListDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names, nil
}

// ListDirFiles returns only the names of regular files in the directory at
// path. Directories, symbolic links, and other non-regular entries are
// excluded. Entries are returned in lexicographic order.
//
// Parameters:
//   - `path`: the directory path to read.
//
// Returns:
//
//	([]string, error): the file names and nil on success, or nil and a
//	non-nil error if the directory cannot be read.
//
// Example:
//
//	files, err := sysx.ListDirFiles("/var/log")
//	for _, f := range files {
//	    fmt.Println(f)
//	}
func ListDirFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Type().IsRegular() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// ListDirDirs returns only the names of subdirectories inside the directory
// at path. Regular files, symbolic links, and other non-directory entries are
// excluded. Entries are returned in lexicographic order.
//
// Parameters:
//   - `path`: the directory path to read.
//
// Returns:
//
//	([]string, error): the subdirectory names and nil on success, or nil and
//	a non-nil error if the directory cannot be read.
//
// Example:
//
//	dirs, err := sysx.ListDirDirs("/home")
//	for _, d := range dirs {
//	    fmt.Println(d)
//	}
func ListDirDirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// CopyDir recursively copies the directory tree from src to dst.
//
// The destination directory is created with the same permissions as src.
// Existing files in the destination are overwritten. Symbolic links within
// the source are replicated at the destination (not followed).
//
// Parameters:
//   - `src`: the source directory path.
//   - `dst`: the destination directory path.
//
// Returns:
//
//	An error if the copy operation fails; nil on success.
//
// Example:
//
//	if err := sysx.CopyDir("/home/user/docs", "/tmp/docs_bak"); err != nil {
//	    log.Fatal(err)
//	}
func CopyDir(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return &os.PathError{Op: "CopyDir", Path: src, Err: os.ErrInvalid}
	}

	if err := os.MkdirAll(dst, fi.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			// On Windows, symlink creation requires elevated privileges or
			// Developer Mode. Fall back to copying the target as a regular
			// file so the operation succeeds without requiring special access.
			if symlinkErr := os.Symlink(link, dstPath); symlinkErr != nil {
				// Resolve the link relative to the source directory.
				target := link
				if !filepath.IsAbs(target) {
					target = filepath.Join(filepath.Dir(srcPath), link)
				}
				// Security fix (CWE-22): Ensure the resolved symlink target
				// stays within the source directory tree. Without this check
				// an attacker-controlled symlink (e.g. ../../../etc/passwd)
				// would cause CopyDir to read arbitrary files on the host.
				cleanTarget := filepath.Clean(target)
				cleanSrc := filepath.Clean(src)
				prefix := cleanSrc + string(filepath.Separator)
				if cleanTarget != cleanSrc &&
					!strings.HasPrefix(cleanTarget, prefix) {
					return fmt.Errorf("sysx: CopyDir: symlink %q target %q escapes source directory %q",
						srcPath, cleanTarget, cleanSrc)
				}
				if err := CopyFile(target, dstPath); err != nil {
					return err
				}
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// ClearDir removes all files and subdirectories from the directory at path,
// but leaves the directory itself intact.
//
// If path does not exist, an error is returned.
//
// Parameters:
//   - `path`: the directory path to clear.
//
// Returns:
//
//	An error if the directory could not be read or its contents removed; nil
//	on success.
//
// Example:
//
//	if err := sysx.ClearDir("/tmp/cache"); err != nil {
//	    log.Fatal(err)
//	}
func ClearDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// IsDirEmpty reports whether the directory at path exists and contains no entries.
//
// Parameters:
//   - `path`: the directory path to check.
//
// Returns:
//
//	(bool, error): true if the directory is empty and nil on success, or false
//	and a non-nil error if the directory does not exist or cannot be read.
//
// Example:
//
//	empty, err := sysx.IsDirEmpty("/tmp/empty_dir")
//	if err == nil && empty {
//	    fmt.Println("directory is empty")
//	}
func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// IsSafeDirEmpty reports whether the directory at path exists and contains no entries.
//
// Parameters:
//   - `path`: the directory path to check.
//
// Returns:
//
//	true if the directory is empty and nil on success, or false if the directory does not exist or cannot be read.
//
// Example:
//
//	empty := sysx.IsSafeDirEmpty("/tmp/empty_dir")
//	if empty {
//	    fmt.Println("directory is empty")
//	}
func IsSafeDirEmpty(path string) bool {
	ok, err := IsDirEmpty(path)
	if err != nil {
		return false
	}
	return ok
}

// RemoveDirIfExist removes the directory at path if it exists.
//
// Parameters:
//   - `path`: the directory path to remove.
//
// Returns:
//
//	An error if the directory could not be removed; nil on success or if the
//	directory does not exist.
//
// Example:
//
//	if err := sysx.RemoveDirIfExist("/tmp/cache"); err != nil {
//	    log.Fatal(err)
//	}
func RemoveDirIfExist(path string) error {
	if DirExists(path) {
		return os.RemoveAll(path)
	}
	return nil
}
