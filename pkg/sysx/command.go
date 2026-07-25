package sysx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/strutil"
)

// NewCommand creates a new Command for the program identified by name.
//
// name may be an absolute path, a relative path, or a bare program name that
// is resolved through the process $PATH. No validation is performed until
// Execute, Run, or Output is called.
//
// Parameters:
//   - `name`: the program name or path to execute.
//
// Returns:
//
// A pointer to a new Command ready for configuration.
//
// Example:
//
// result := sysx.NewCommand("git").
//
//	WithArgs("rev-parse", "HEAD").
//	WithDir("/path/to/repo").
//	Execute()
//
//	if result.IsSuccess() {
//	   fmt.Println(strings.TrimSpace(result.Stdout()))
//	}
func NewCommand(name string) *Command {
	return &Command{name: name}
}

// Name returns the program name or path configured on the Command.
//
// Returns:
//
//	A string containing the program name or path.
//
// Example:
//
//	cmd := sysx.NewCommand("git")
//	fmt.Println(cmd.Name()) // "git"
func (c *Command) Name() string {
	return c.name
}

// Args returns the positional arguments configured on the Command.
//
// Returns:
//
//	A slice of strings containing the command-line arguments.
//
// Example:
//
//	cmd := sysx.NewCommand("git").WithArgs("log", "--oneline")
//	fmt.Println(cmd.Args()) // ["log", "--oneline"]
func (c *Command) Args() []string {
	return c.args
}

// Dir returns the working directory configured on the Command.
// An empty string means the child process inherits the caller's directory.
//
// Returns:
//
//	A string containing the working directory path, or empty if not set.
func (c *Command) Dir() string {
	return c.dir
}

// Env returns the extra environment variable bindings ("KEY=VALUE") that
// will be merged on top of the calling process environment when the command
// is executed.
//
// Returns:
//
//	A slice of "KEY=VALUE" strings; nil if no extra bindings were added.
func (c *Command) Env() []string {
	return c.env
}

// Timeout returns the maximum execution duration configured on the Command.
// A zero duration means no timeout is applied.
//
// Returns:
//
//	A time.Duration; zero if no timeout was set.
func (c *Command) Timeout() time.Duration {
	return c.timeout
}

// WithArgs sets the positional arguments passed to the program.
// Calling WithArgs replaces any previously set arguments.
//
// Parameters:
//   - `args`: the command-line arguments.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithArgs(args ...string) *Command {
	c.args = args
	return c
}

// WithDir sets the working directory for the command.
// An empty string leaves the working directory unchanged
// (the child inherits the calling process's directory).
//
// Parameters:
//   - `dir`: the working directory path.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithDir(dir string) *Command {
	c.dir = dir
	return c
}

// WithEnv appends one or more environment variable bindings in "KEY=VALUE"
// form to the command's environment. Bindings are merged on top of the
// calling process environment; later values for the same key shadow earlier ones.
// Calling WithEnv multiple times accumulates bindings.
//
// Parameters:
//   - `env`: one or more "KEY=VALUE" environment variable bindings.
//
// Returns:
//
// The receiver, enabling method chaining.
//
// Example:
//
// sysx.NewCommand("go").
//
//	WithArgs("build", "./...").
//	WithEnv("GOOS=linux", "GOARCH=amd64").
//	Execute()
func (c *Command) WithEnv(env ...string) *Command {
	c.env = append(c.env, env...)
	return c
}

// WithEnvCast sets a single environment variable binding.
// This is a convenience method that is equivalent to calling WithEnv(fmt.Sprintf("%s=%s", key, value)).
//
// Parameters:
//   - `key`: the environment variable key.
//   - `value`: the environment variable value.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithEnvCast(key, value any) *Command {
	return c.WithEnv(fmt.Sprintf("%s=%s", key, conv.StringOrEmpty(value)))
}

// WithEnvf appends an environment variable binding created from a format string.
// This is a convenience method that is equivalent to calling WithEnv(fmt.Sprintf(format, args...)).
//
// Parameters:
//   - `format`: the format string.
//   - `args`: the arguments to format.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithEnvf(format string, args ...any) *Command {
	return c.WithEnv(fmt.Sprintf(format, args...))
}

// WithTimeout sets a maximum duration for the command. If the command does not
// finish within the deadline the process is killed and Execute returns a
// CommandResult whose Err() wraps context.DeadlineExceeded.
//
// A zero or negative duration means no timeout is applied.
//
// Parameters:
//   - `d`: the maximum duration to wait.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithTimeout(d time.Duration) *Command {
	c.timeout = d
	return c
}

// WithContext attaches an existing context to the command, enabling external
// cancellation and deadline propagation. If both WithContext and WithTimeout
// are used, the more restrictive deadline wins.
//
// Parameters:
//   - `ctx`: the context to attach.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithContext(ctx context.Context) *Command {
	c.ctx = ctx
	return c
}

// WithStdin sets the reader that supplies the command's standard input.
//
// Parameters:
//   - `r`: the io.Reader to use as stdin.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithStdin(r io.Reader) *Command {
	c.stdin = r
	return c
}

// WithStdout sets the writer to which the command's standard output is
// forwarded in real time. When a custom writer is provided,
// CommandResult.Stdout() will be empty because data is not buffered internally.
//
// Parameters:
//   - `w`: the io.Writer to receive stdout.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithStdout(w io.Writer) *Command {
	c.stdout = w
	return c
}

// WithStderr sets the writer to which the command's standard error is
// forwarded in real time. When a custom writer is provided,
// CommandResult.Stderr() will be empty because data is not buffered internally.
//
// Parameters:
//   - `w`: the io.Writer to receive stderr.
//
// Returns:
//
// The receiver, enabling method chaining.
func (c *Command) WithStderr(w io.Writer) *Command {
	c.stderr = w
	return c
}

// Execute runs the command and returns a CommandResult containing captured
// stdout, stderr, exit code, wall-clock duration, and any error.
//
// If WithStdout or WithStderr were provided, the corresponding fields in
// CommandResult are empty because the data was streamed to those writers.
//
// Execute is safe to call multiple times; each call spawns a new process.
//
// Returns:
//
// A non-nil *CommandResult describing the outcome of the command.
//
// Example:
//
// res := sysx.NewCommand("bash").
//
//	WithArgs("-c", "echo hello").
//	WithTimeout(5 * time.Second).
//	WithEnv("APP_ENV=prod").
//	WithDir("/tmp").
//	Execute()
//
// fmt.Printf("exit=%d stdout=%q duration=%v\n", res.ExitCode(), res.Stdout(), res.Duration())
func (c *Command) Execute() *CommandResult {
	if strutil.IsEmpty(c.name) {
		return &CommandResult{
			err:      errors.New("sysx: command name must not be empty"),
			exitCode: -1,
		}
	}
	cmd, cancel := c.buildCmd()
	defer cancel()

	var outBuf, errBuf commandBuffer
	if c.stdout != nil {
		cmd.Stdout = c.stdout
	} else {
		cmd.Stdout = &outBuf
	}
	if c.stderr != nil {
		cmd.Stderr = c.stderr
	} else {
		cmd.Stderr = &errBuf
	}

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &CommandResult{duration: duration}
	if c.stdout == nil {
		result.stdout = outBuf.String()
	}
	if c.stderr == nil {
		result.stderr = errBuf.String()
	}
	if err != nil {
		result.err = err
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.exitCode = exitErr.ExitCode()
		} else {
			result.exitCode = -1
		}
	}
	return result
}

// Run runs the command, discarding all output, and returns only the error.
// It is a convenience method equivalent to calling Execute and inspecting
// the Err() result.
//
// Returns:
//
// An error if the command could not be started or exited non-zero; nil on success.
func (c *Command) Run() error {
	return c.Execute().Err()
}

// Output runs the command and returns the combined stdout+stderr as a string
// along with any error.
//
// Returns:
//
// (string, error): the combined output and nil on success, or combined
// partial output and a non-nil error on failure.
func (c *Command) Output() (string, error) {
	r := c.Execute()
	return r.Combined(), r.Err()
}

// buildCmd constructs the underlying *exec.Cmd from the Command fields.
// The returned CancelFunc must always be called to release context resources.
func (c *Command) buildCmd() (*exec.Cmd, context.CancelFunc) {
	base := c.ctx
	if base == nil {
		base = context.Background()
	}
	var cancel context.CancelFunc
	if c.timeout > 0 {
		base, cancel = context.WithTimeout(base, c.timeout)
	} else {
		cancel = func() {}
	}
	cmd := exec.CommandContext(base, c.name, c.args...)
	if strutil.IsNotEmpty(c.dir) {
		cmd.Dir = c.dir
	}
	if len(c.env) > 0 {
		cmd.Env = append(os.Environ(), c.env...)
	}
	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}
	return cmd, cancel
}

// Stdout returns the captured standard output of the command.
// Returns an empty string when a custom io.Writer was provided via WithStdout.
//
// Returns:
//
//	A string containing the captured stdout of the command.
func (r *CommandResult) Stdout() string {
	return r.stdout
}

// Stderr returns the captured standard error of the command.
// Returns an empty string when a custom io.Writer was provided via WithStderr.
//
// Returns:
//
//	A string containing the captured stderr of the command.
func (r *CommandResult) Stderr() string {
	return r.stderr
}

// ExitCode returns the process exit code; 0 indicates success.
// -1 indicates that the exit code could not be determined
// (e.g. the process was killed by a signal or a context was cancelled).
//
// Returns:
//
//	An int representing the process exit code.
func (r *CommandResult) ExitCode() int {
	return r.exitCode
}

// Duration returns the wall-clock time spent waiting for the command to complete.
//
// Returns:
//
//	A time.Duration representing the execution time.
func (r *CommandResult) Duration() time.Duration {
	return r.duration
}

// Err returns the error from command execution.
// A non-nil value indicates the command could not be started or exited non-zero.
//
// Returns:
//
//	An error describing the failure, or nil on success.
func (r *CommandResult) Err() error {
	return r.err
}

// Success reports whether the command completed without error.
//
// Returns:
//
//	true when Err() is nil; false otherwise.
func (r *CommandResult) IsSuccess() bool {
	return r.err == nil
}

// Combined returns the concatenation of Stdout followed by Stderr.
//
// Returns:
//
//	A string containing the combined output of the command.
func (r *CommandResult) Combined() string {
	return r.stdout + r.stderr
}

// Write appends p to the buffer.
func (b *commandBuffer) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

// String returns the accumulated content as a string.
func (b *commandBuffer) String() string {
	return b.buf.String()
}
