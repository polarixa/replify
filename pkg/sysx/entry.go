package sysx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// IsTTY reports whether w is connected to a terminal (character device).
//
// Parameters:
//   - `w`: the writer to test
//
// Returns:
//
// true when w is an *os.File whose device mode includes os.ModeCharDevice.
func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// UserInfo returns a formatted string containing the numeric user and group
// identifiers of the current process.
//
// The string has the form "uid=X gid=Y", where X and Y are the values
// returned by UID() and GID() respectively.
//
// Returns:
//
//	A non-empty string of the form "uid=<uid> gid=<gid>".
//
// Example:
//
//	fmt.Println(sysx.UserInfo()) // "uid=1000 gid=1000"
func UserInfo() string {
	return fmt.Sprintf("uid=%d gid=%d", UID(), GID())
}

// SystemInfo returns a map of key/value strings summarising the most
// commonly needed runtime and operating system attributes of the current
// process.
//
// The map always contains the following keys:
//   - "os"         – operating system name (runtime.GOOS)
//   - "arch"       – CPU architecture (runtime.GOARCH)
//   - "hostname"   – machine hostname; empty string on lookup failure
//   - "pid"        – current process identifier as a decimal string
//   - "go_version" – Go runtime version (e.g. "go1.24.0")
//   - "executable" – path of the current executable; empty on failure
//   - "num_cpu"    – number of logical CPUs as a decimal string
//
// Returns:
//
//	A non-nil map[string]string containing the system information entries.
//
// Example:
//
//	info := sysx.SystemInfo()
//	for k, v := range info {
//	    fmt.Printf("%s = %s\n", k, v)
//	}
func SystemInfo() map[string]string {
	host, _ := Hostname()
	exe, _ := ExecutablePath()
	return map[string]string{
		"os":         OSName(),
		"arch":       Arch(),
		"hostname":   host,
		"pid":        strconv.Itoa(PID()),
		"go_version": GoVersion(),
		"executable": exe,
		"num_cpu":    fmt.Sprintf("%d", NumCPU()),
	}
}

// IsPrivileged reports whether the current process is running with root
// (super-user) privileges.
//
// On Unix-like systems this is determined by checking whether the effective
// user identifier is 0. On Windows, UID() returns -1 and this function
// always returns false; use the Windows API for accurate privilege checks.
//
// Returns:
//
//	A boolean value:
//	 - true  when UID() == 0 (running as root on Unix);
//	 - false otherwise.
//
// Example:
//
//	if sysx.IsPrivileged() {
//	    fmt.Println("running as root")
//	}
func IsPrivileged() bool {
	return UID() == 0
}

// RunCommand creates a Command for name, runs it with the provided args, and
// returns the structured CommandResult.
//
// It is the single-call equivalent of:
//
// NewCommand(name).WithArgs(args...).Execute()
//
// Parameters:
//   - `name`: the program name or path.
//   - `args`: optional arguments.
//
// Returns:
//
// A non-nil *CommandResult.
//
// Example:
//
// res := sysx.RunCommand("git", "status")
//
//	if !res.IsSuccess() {
//	   log.Printf("exit %d: %s", res.ExitCode(), res.Stderr())
//	}
func RunCommand(name string, args ...string) *CommandResult {
	return NewCommand(name).WithArgs(args...).Execute()
}

// ExecCommand runs the named program with the provided arguments and waits
// for it to complete. Both stdout and stderr are discarded.
//
// The function does not use shell interpolation; name must be a program name
// or absolute path. An error is returned if name is empty, if the program
// cannot be found, or if the program exits with a non-zero status.
//
// Parameters:
//   - `name`: the program name or path to execute.
//   - `args`: optional arguments to pass to the program.
//
// Returns:
//
// An error if the command could not be started or exited non-zero, or nil on success.
//
// Example:
//
//	if err := sysx.ExecCommand("go", "build", "./..."); err != nil {
//	   log.Fatal(err)
//	}
func ExecCommand(name string, args ...string) error {
	if strutil.IsEmpty(name) {
		return errors.New("sysx: command name must not be empty")
	}
	return NewCommand(name).WithArgs(args...).Run()
}

// ExecCommandContext runs the named program under the provided context.
// The command is cancelled when ctx is cancelled or its deadline expires.
//
// Parameters:
//   - `ctx`:  the context controlling cancellation and deadline.
//   - `name`: the program name or path.
//   - `args`: optional arguments.
//
// Returns:
//
// An error if the command failed, was cancelled, or timed out; nil on success.
//
// Example:
//
// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// defer cancel()
//
//	if err := sysx.ExecCommandContext(ctx, "go", "test", "./..."); err != nil {
//	   log.Fatal(err)
//	}
func ExecCommandContext(ctx context.Context, name string, args ...string) error {
	if strutil.IsEmpty(name) {
		return errors.New("sysx: command name must not be empty")
	}
	return NewCommand(name).WithArgs(args...).WithContext(ctx).Run()
}

// ExecCommandWithTimeout runs the named program with the provided arguments
// and a deadline of timeout. If the command does not finish within the
// deadline the process is killed and a context deadline-exceeded error is
// returned.
//
// Parameters:
//   - `timeout`: maximum duration to wait for the command to complete.
//   - `name`:    the program name or path to execute.
//   - `args`:    optional arguments to pass to the program.
//
// Returns:
//
// An error if the command timed out, could not be started, or exited non-zero, or nil on success.
//
// Example:
//
// err := sysx.ExecCommandWithTimeout(5*time.Second, "ping", "-c", "1", "localhost")
func ExecCommandWithTimeout(timeout time.Duration, name string, args ...string) error {
	if strutil.IsEmpty(name) {
		return errors.New("sysx: command name must not be empty")
	}
	return NewCommand(name).WithArgs(args...).WithTimeout(timeout).Run()
}

// ExecOutput runs the named program with the provided arguments, waits for it
// to complete, and returns the combined stdout and stderr output as a string.
//
// The function does not use shell interpolation. An error is returned if name
// is empty, if the program cannot be found, or if the program exits with a
// non-zero status.
//
// Parameters:
//   - `name`: the program name or path to execute.
//   - `args`: optional arguments to pass to the program.
//
// Returns:
//
// (string, error): the combined output and nil on success, or the partial
// output and a non-nil error on failure.
//
// Example:
//
// out, err := sysx.ExecOutput("git", "rev-parse", "HEAD")
//
//	if err != nil {
//	   log.Fatal(err)
//	}
//
// fmt.Println(strings.TrimSpace(out))
func ExecOutput(name string, args ...string) (string, error) {
	if strutil.IsEmpty(name) {
		return "", errors.New("sysx: command name must not be empty")
	}
	r := NewCommand(name).WithArgs(args...).Execute()
	return r.Combined(), r.Err()
}

// ExecOutputLines runs the named program and returns its stdout split into
// individual lines. Line endings are stripped; empty lines are preserved.
// Stderr is captured but discarded on success. On failure, the error from the
// command is returned.
//
// Parameters:
//   - `name`: the program name or path.
//   - `args`: optional arguments.
//
// Returns:
//
// ([]string, error): lines of stdout and nil on success, or nil and a
// non-nil error on failure.
//
// Example:
//
// lines, err := sysx.ExecOutputLines("git", "log", "--oneline", "-5")
//
//	for _, l := range lines {
//	   fmt.Println(l)
//	}
func ExecOutputLines(name string, args ...string) ([]string, error) {
	if strutil.IsEmpty(name) {
		return nil, errors.New("sysx: command name must not be empty")
	}
	run := NewCommand(name).WithArgs(args...).Execute()
	if run.Err() != nil {
		return nil, run.Err()
	}
	return splitLines(run.Stdout()), nil
}

// ExecOutputWithTimeout runs the named program with the provided arguments
// and a deadline of timeout, returning the combined stdout and stderr.
//
// If the command does not finish within the deadline the process is killed and
// a context deadline-exceeded error is returned along with any output produced
// before the timeout.
//
// Parameters:
//   - `timeout`: maximum duration to wait for the command to complete.
//   - `name`:    the program name or path to execute.
//   - `args`:    optional arguments to pass to the program.
//
// Returns:
//
// (string, error): the combined output and nil on success, or partial output
// and a non-nil error on failure or timeout.
//
// Example:
//
// out, err := sysx.ExecOutputWithTimeout(3*time.Second, "curl", "-s", "http://localhost")
func ExecOutputWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	if strutil.IsEmpty(name) {
		return "", errors.New("sysx: command name must not be empty")
	}
	cmd := NewCommand(name).WithArgs(args...).WithTimeout(timeout).Execute()
	return cmd.Combined(), cmd.Err()
}

// ExecStreaming runs the named program, forwarding its stdout and stderr to
// the provided writers in real time as the command executes. Either writer
// may be nil to discard the corresponding stream.
//
// Parameters:
//   - `stdout`: writer receiving standard output; nil to discard.
//   - `stderr`: writer receiving standard error; nil to discard.
//   - `name`:   the program name or path.
//   - `args`:   optional arguments.
//
// Returns:
//
// An error if the command failed or could not be started; nil on success.
//
// Example:
//
// err := sysx.ExecStreaming(os.Stdout, os.Stderr, "go", "build", "./...")
func ExecStreaming(stdout, stderr io.Writer, name string, args ...string) error {
	if strutil.IsEmpty(name) {
		return errors.New("sysx: command name must not be empty")
	}
	c := NewCommand(name).WithArgs(args...)
	if stdout != nil {
		c = c.WithStdout(stdout)
	}
	if stderr != nil {
		c = c.WithStderr(stderr)
	}
	return c.Run()
}

// ExecAsync starts the named program asynchronously and returns the underlying
// *exec.Cmd without waiting for it to finish. The caller is responsible for
// calling cmd.Wait() to release associated resources and obtain the exit status.
//
// Parameters:
//   - `name`: the program name or path.
//   - `args`: optional arguments.
//
// Returns:
//
// (*exec.Cmd, error): the started command handle and nil on success, or nil
// and a non-nil error if the command could not be started.
//
// Example:
//
// cmd, err := sysx.ExecAsync("long-running-server", "--port", "8080")
//
//	if err != nil {
//	   log.Fatal(err)
//	}
//
// // ... do other work ...
// cmd.Wait()
func ExecAsync(name string, args ...string) (*exec.Cmd, error) {
	if strutil.IsEmpty(name) {
		return nil, errors.New("sysx: command name must not be empty")
	}
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// ExecPipeline executes a sequence of commands as a shell pipeline: the
// standard output of each command is connected to the standard input of the
// next. The stdout of the final command is returned as a string.
//
// Each element of commands is a []string where the first element is the
// program name and the remaining elements are its arguments. An error is
// returned if any command in the pipeline fails to start or exits non-zero.
//
// Parameters:
//   - `commands`: one or more [name, arg...] command descriptors.
//
// Returns:
//
// (string, error): the stdout of the last command and nil on success, or
// partial output and a non-nil error on failure.
//
// Example:
//
// out, err := sysx.ExecPipeline(
//
//	[]string{"cat", "/etc/passwd"},
//	[]string{"grep", "root"},
//	[]string{"cut", "-d:", "-f1"},
//
// )
func ExecPipeline(commands ...[]string) (string, error) {
	if len(commands) == 0 {
		return "", errors.New("sysx: pipeline requires at least one command")
	}
	cmds := make([]*exec.Cmd, len(commands))
	for i, args := range commands {
		if len(args) == 0 {
			return "", fmt.Errorf("sysx: pipeline command at index %d has no name", i)
		}
		cmds[i] = exec.Command(args[0], args[1:]...)
	}
	// Wire stdout of each command to stdin of the next.
	for i := 0; i < len(cmds)-1; i++ {
		pipe, err := cmds[i].StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("sysx: pipeline pipe error at index %d: %w", i, err)
		}
		cmds[i+1].Stdin = pipe
	}
	var outBuf commandBuffer
	cmds[len(cmds)-1].Stdout = &outBuf

	// Start all commands in order.
	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return "", fmt.Errorf("sysx: pipeline start error at index %d: %w", i, err)
		}
	}
	// Wait for all commands in order so pipes drain correctly.
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return outBuf.String(), fmt.Errorf("sysx: pipeline wait error at index %d: %w", i, err)
		}
	}
	return outBuf.String(), nil
}

// ExecCommandInDir runs the named program with the provided arguments from
// the specified working directory and waits for it to complete.
//
// stdout and stderr are discarded. The function does not use shell
// interpolation.
//
// Parameters:
//   - `dir`:  the working directory in which to run the command.
//   - `name`: the program name or path to execute.
//   - `args`: optional arguments to pass to the program.
//
// Returns:
//
// An error if the command could not be started or exited non-zero, or nil on success.
//
// Example:
//
// err := sysx.ExecCommandInDir("/tmp/myproject", "go", "test", "./...")
func ExecCommandInDir(dir, name string, args ...string) error {
	if strutil.IsEmpty(name) {
		return errors.New("sysx: command name must not be empty")
	}
	return NewCommand(name).WithArgs(args...).WithDir(dir).Run()
}

// ExecOutputInDir runs the named program with the provided arguments from
// the specified working directory, waits for it to complete, and returns the
// combined stdout and stderr output.
//
// Parameters:
//   - `dir`:  the working directory in which to run the command.
//   - `name`: the program name or path to execute.
//   - `args`: optional arguments to pass to the program.
//
// Returns:
//
// (string, error): the combined output and nil on success, or the partial
// output and a non-nil error on failure.
//
// Example:
//
// out, err := sysx.ExecOutputInDir("/tmp/myproject", "git", "status")
func ExecOutputInDir(dir, name string, args ...string) (string, error) {
	if strutil.IsEmpty(name) {
		return "", errors.New("sysx: command name must not be empty")
	}
	r := NewCommand(name).WithArgs(args...).WithDir(dir).Execute()
	return r.Combined(), r.Err()
}
