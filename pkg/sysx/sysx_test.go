package sysx_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/sysx"
)

// ///////////////////////////
// Section: OS detection tests
// ///////////////////////////

func TestSysx_OSDetection(t *testing.T) {
	t.Parallel()
	isLinux := sysx.IsLinux()
	isDarwin := sysx.IsDarwin()
	isWindows := sysx.IsWindows()

	count := 0
	if isLinux {
		count++
	}
	if isDarwin {
		count++
	}
	if isWindows {
		count++
	}
	if count != 1 {
		t.Errorf("expected exactly one OS flag to be true; got IsLinux=%v IsDarwin=%v IsWindows=%v",
			isLinux, isDarwin, isWindows)
	}
}

func TestSysx_OSName(t *testing.T) {
	t.Parallel()
	name := sysx.OSName()
	if name == "" {
		t.Error("OSName() returned empty string")
	}
}

func TestSysx_Arch(t *testing.T) {
	t.Parallel()
	arch := sysx.Arch()
	if arch == "" {
		t.Error("Arch() returned empty string")
	}
}

// ///////////////////////////
// Section: Runtime tests
// ///////////////////////////

func TestSysx_PID(t *testing.T) {
	t.Parallel()
	pid := sysx.PID()
	if pid <= 0 {
		t.Errorf("PID() = %d; want positive integer", pid)
	}
}

func TestSysx_UID(t *testing.T) {
	t.Parallel()
	uid := sysx.UID()
	if uid < 0 && runtime.GOOS != "windows" {
		t.Errorf("UID() = %d; want non-negative integer on non-Windows", uid)
	}
}

func TestSysx_GID(t *testing.T) {
	t.Parallel()
	gid := sysx.GID()
	if gid < 0 && runtime.GOOS != "windows" {
		t.Errorf("GID() = %d; want non-negative integer on non-Windows", gid)
	}
}

func TestSysx_Hostname(t *testing.T) {
	t.Parallel()
	host, err := sysx.Hostname()
	if err != nil {
		t.Fatalf("Hostname() error = %v", err)
	}
	if host == "" {
		t.Error("Hostname() returned empty string")
	}
}

func TestSysx_GoVersion(t *testing.T) {
	t.Parallel()
	v := sysx.GoVersion()
	if !strings.HasPrefix(v, "go") {
		t.Errorf("GoVersion() = %q; want string starting with 'go'", v)
	}
}

func TestSysx_NumCPU(t *testing.T) {
	t.Parallel()
	n := sysx.NumCPU()
	if n < 1 {
		t.Errorf("NumCPU() = %d; want >= 1", n)
	}
}

// ///////////////////////////
// Section: Environment tests
// ///////////////////////////

func TestSysx_GetEnv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		key      string
		setup    func()
		teardown func()
		fallback string
		want     string
	}{
		{
			name:     "unset variable returns fallback",
			key:      "SYSX_TEST_UNSET_XYZ",
			setup:    func() { os.Unsetenv("SYSX_TEST_UNSET_XYZ") },
			teardown: func() {},
			fallback: "default",
			want:     "default",
		},
		{
			name: "set variable returns value",
			key:  "SYSX_TEST_SET_XYZ",
			setup: func() {
				os.Setenv("SYSX_TEST_SET_XYZ", "hello")
			},
			teardown: func() { os.Unsetenv("SYSX_TEST_SET_XYZ") },
			fallback: "default",
			want:     "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()
			got := sysx.Getenv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetEnv(%q, %q) = %q; want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestSysx_HasEnv(t *testing.T) {
	t.Parallel()
	const key = "SYSX_HAS_ENV_TEST"
	os.Unsetenv(key)
	if sysx.Hasenv(key) {
		t.Errorf("HasEnv(%q) = true; want false when not set", key)
	}
	os.Setenv(key, "yes")
	defer os.Unsetenv(key)
	if !sysx.Hasenv(key) {
		t.Errorf("HasEnv(%q) = false; want true when set", key)
	}
}

func TestSysx_GetEnvInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		set      bool
		fallback int
		want     int
	}{
		{"unset returns fallback", "", false, 42, 42},
		{"valid int", "10", true, 42, 10},
		{"invalid value returns fallback", "abc", true, 42, 42},
	}
	const key = "SYSX_INT_TEST"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				os.Setenv(key, tt.value)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}
			got := sysx.GetenvInt(key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetEnvInt(%q, %d) = %d; want %d", key, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestSysx_GetEnvBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value    string
		fallback bool
		want     bool
	}{
		{"1", false, true},
		{"true", false, true},
		{"yes", false, true},
		{"on", false, true},
		{"TRUE", false, true},
		{"YES", false, true},
		{"0", true, false},
		{"false", true, false},
		{"no", true, false},
		{"off", true, false},
		{"garbage", true, true},
	}
	const key = "SYSX_BOOL_TEST"
	for _, tt := range tests {
		t.Run(tt.value+"_fallback_"+func() string {
			if tt.fallback {
				return "true"
			}
			return "false"
		}(), func(t *testing.T) {
			os.Setenv(key, tt.value)
			defer os.Unsetenv(key)
			got := sysx.GetenvBool(key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetEnvBool(%q, %v) with value %q = %v; want %v",
					key, tt.fallback, tt.value, got, tt.want)
			}
		})
	}
}

func TestSysx_GetEnvSlice(t *testing.T) {
	t.Parallel()
	const key = "SYSX_SLICE_TEST"
	os.Setenv(key, "a,b,c")
	defer os.Unsetenv(key)

	got := sysx.GetenvSlice(key, ",")
	if len(got) != 3 {
		t.Fatalf("GetEnvSlice len = %d; want 3", len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("GetEnvSlice = %v; want [a b c]", got)
	}

	os.Unsetenv(key)
	if sysx.GetenvSlice(key, ",") != nil {
		t.Error("GetEnvSlice on unset key should return nil")
	}
}

func TestSysx_EnvMap(t *testing.T) {
	t.Parallel()
	m := sysx.EnvMap()
	if m == nil {
		t.Error("EnvMap() returned nil")
	}
}

// ///////////////////////////
// Section: File system tests
// ///////////////////////////

func TestSysx_FileExists(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_test_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	if !sysx.FileExists(path) {
		t.Errorf("FileExists(%q) = false; want true", path)
	}
	if sysx.FileExists(path + "_nonexistent") {
		t.Errorf("FileExists(%q) = true; want false", path+"_nonexistent")
	}
}

func TestSysx_DirExists(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "sysx_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if !sysx.DirExists(dir) {
		t.Errorf("DirExists(%q) = false; want true", dir)
	}
	if sysx.DirExists(dir + "_nonexistent") {
		t.Errorf("DirExists(%q) = true; want false", dir+"_nonexistent")
	}
}

func TestSysx_IsFile_IsDir(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_isfile_*")
	if err != nil {
		t.Fatal(err)
	}
	fpath := f.Name()
	f.Close()
	defer os.Remove(fpath)

	dir, err := os.MkdirTemp("", "sysx_isdir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if !sysx.IsFile(fpath) {
		t.Errorf("IsFile(%q) = false; want true", fpath)
	}
	if sysx.IsDir(fpath) {
		t.Errorf("IsDir(%q) = true for a file; want false", fpath)
	}
	if sysx.IsFile(dir) {
		t.Errorf("IsFile(%q) = true for a dir; want false", dir)
	}
	if !sysx.IsDir(dir) {
		t.Errorf("IsDir(%q) = false; want true", dir)
	}
}

func TestSysx_FileSize(t *testing.T) {
	t.Parallel()
	content := []byte("hello sysx")
	f, err := os.CreateTemp("", "sysx_size_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Write(content)
	f.Close()
	defer os.Remove(path)

	size, err := sysx.FileSize(path)
	if err != nil {
		t.Fatalf("FileSize error = %v", err)
	}
	if size != int64(len(content)) {
		t.Errorf("FileSize = %d; want %d", size, len(content))
	}

	_, err = sysx.FileSize(path + "_nonexistent")
	if err == nil {
		t.Error("FileSize on non-existent path should return error")
	}
}

func TestSysx_TempDir(t *testing.T) {
	t.Parallel()
	tmp := sysx.TempDir()
	if tmp == "" {
		t.Error("TempDir() returned empty string")
	}
}

func TestSysx_HomeDir(t *testing.T) {
	t.Parallel()
	home, err := sysx.HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() error = %v", err)
	}
	if home == "" {
		t.Error("HomeDir() returned empty string")
	}
}

func TestSysx_WorkingDir(t *testing.T) {
	t.Parallel()
	wd, err := sysx.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error = %v", err)
	}
	if wd == "" {
		t.Error("WorkingDir() returned empty string")
	}
}

// ///////////////////////////
// Section: Process tests
// ///////////////////////////

func TestSysx_ProcessExists(t *testing.T) {
	t.Parallel()
	if !sysx.ProcessExists(os.Getpid()) {
		t.Errorf("ProcessExists(os.Getpid()) = false; want true")
	}
	if sysx.ProcessExists(-1) {
		t.Error("ProcessExists(-1) = true; want false")
	}
}

func TestSysx_CurrentProcessName(t *testing.T) {
	t.Parallel()
	name := sysx.CurrentProcessName()
	if name == "" {
		t.Error("CurrentProcessName() returned empty string")
	}
}

// ///////////////////////////
// Section: Command execution tests
// ///////////////////////////

func TestSysx_ExecCommand(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "hello"}
	} else {
		cmd = "echo"
		args = []string{"hello"}
	}
	if err := sysx.ExecCommand(cmd, args...); err != nil {
		t.Errorf("ExecCommand(%q) error = %v", cmd, err)
	}
}

func TestSysx_ExecOutput(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "hello"}
	} else {
		cmd = "echo"
		args = []string{"hello"}
	}
	out, err := sysx.ExecOutput(cmd, args...)
	if err != nil {
		t.Fatalf("ExecOutput error = %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("ExecOutput = %q; want output containing 'hello'", out)
	}
}

func TestSysx_ExecCommandWithTimeout(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "timeout_test"}
	} else {
		cmd = "echo"
		args = []string{"timeout_test"}
	}

	// Should complete well within 5 seconds.
	if err := sysx.ExecCommandWithTimeout(5*time.Second, cmd, args...); err != nil {
		t.Errorf("ExecCommandWithTimeout error = %v", err)
	}

	// A command that takes longer than the timeout should fail.
	var sleepCmd string
	var sleepArgs []string
	if runtime.GOOS == "windows" {
		sleepCmd = "cmd"
		sleepArgs = []string{"/c", "timeout", "10"}
	} else {
		sleepCmd = "sleep"
		sleepArgs = []string{"10"}
	}
	err := sysx.ExecCommandWithTimeout(100*time.Millisecond, sleepCmd, sleepArgs...)
	if err == nil {
		t.Error("ExecCommandWithTimeout: expected error on timeout, got nil")
	}
}

// ///////////////////////////
// Section: Entry-point tests
// ///////////////////////////

func TestSysx_SystemInfo(t *testing.T) {
	t.Parallel()
	info := sysx.SystemInfo()
	if info == nil {
		t.Fatal("SystemInfo() returned nil")
	}
	required := []string{"os", "arch", "hostname", "pid", "go_version", "executable", "num_cpu"}
	for _, key := range required {
		if _, ok := info[key]; !ok {
			t.Errorf("SystemInfo() missing key %q", key)
		}
	}
	if info["os"] == "" {
		t.Error("SystemInfo()[\"os\"] is empty")
	}
	if info["arch"] == "" {
		t.Error("SystemInfo()[\"arch\"] is empty")
	}
	if info["go_version"] == "" {
		t.Error("SystemInfo()[\"go_version\"] is empty")
	}
}

// ///////////////////////////
// Section: Command builder tests
// ///////////////////////////

func TestSysx_CommandResult_Success(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "ok"}
	} else {
		cmd = "echo"
		args = []string{"ok"}
	}
	res := sysx.NewCommand(cmd).WithArgs(args...).Execute()
	if res == nil {
		t.Fatal("Execute() returned nil")
	}
	if !res.IsSuccess() {
		t.Errorf("Success() = false, want true; err=%v", res.Err())
	}
	if res.ExitCode() != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode())
	}
	if res.Duration() <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestSysx_CommandResult_EmptyName(t *testing.T) {
	t.Parallel()
	res := sysx.NewCommand("").Execute()
	if res.IsSuccess() {
		t.Error("Execute() on empty name should fail")
	}
	if res.ExitCode() != -1 {
		t.Errorf("ExitCode = %d, want -1 for empty name", res.ExitCode())
	}
}

func TestSysx_CommandResult_CombinedOutput(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	res := sysx.NewCommand("bash").WithArgs("-c", "echo stdout; echo stderr >&2").Execute()
	combined := res.Combined()
	if !strings.Contains(combined, "stdout") {
		t.Errorf("Combined() = %q; want to contain 'stdout'", combined)
	}
	if !strings.Contains(combined, "stderr") {
		t.Errorf("Combined() = %q; want to contain 'stderr'", combined)
	}
}

func TestSysx_CommandBuilder_WithDir(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	tmp, err := os.MkdirTemp("", "sysx_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	res := sysx.NewCommand("pwd").WithDir(tmp).Execute()
	if !res.IsSuccess() {
		t.Fatalf("pwd failed: %v", res.Err())
	}
	if !strings.Contains(strings.TrimSpace(res.Stdout()), tmp) {
		t.Errorf("pwd output %q does not contain %q", res.Stdout(), tmp)
	}
}

func TestSysx_CommandBuilder_WithEnv(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	res := sysx.NewCommand("bash").
		WithArgs("-c", "echo $SYSX_CMD_ENV_TEST").
		WithEnv("SYSX_CMD_ENV_TEST=injected").
		Execute()
	if !res.IsSuccess() {
		t.Fatalf("command failed: %v", res.Err())
	}
	if !strings.Contains(res.Stdout(), "injected") {
		t.Errorf("env not injected: stdout=%q", res.Stdout())
	}
}

func TestSysx_CommandBuilder_WithTimeout_Success(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "fast"}
	} else {
		cmd = "echo"
		args = []string{"fast"}
	}
	res := sysx.NewCommand(cmd).WithArgs(args...).WithTimeout(5 * time.Second).Execute()
	if !res.IsSuccess() {
		t.Errorf("WithTimeout fast command failed: %v", res.Err())
	}
}

func TestSysx_CommandBuilder_WithTimeout_Expired(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "timeout", "10"}
	} else {
		cmd = "sleep"
		args = []string{"10"}
	}
	res := sysx.NewCommand(cmd).WithArgs(args...).WithTimeout(100 * time.Millisecond).Execute()
	if res.IsSuccess() {
		t.Error("expected timeout error, got success")
	}
}

func TestSysx_CommandBuilder_WithContext_Cancelled(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	res := sysx.NewCommand("sleep").WithArgs("10").WithContext(ctx).Execute()
	if res.IsSuccess() {
		t.Error("expected context cancellation error, got success")
	}
}

func TestSysx_CommandBuilder_WithStdout(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	var buf bytes.Buffer
	res := sysx.NewCommand("echo").WithArgs("streaming").WithStdout(&buf).Execute()
	if !res.IsSuccess() {
		t.Fatalf("command failed: %v", res.Err())
	}
	if res.Stdout() != "" {
		t.Errorf("Stdout should be empty when custom writer is set, got %q", res.Stdout())
	}
	if !strings.Contains(buf.String(), "streaming") {
		t.Errorf("custom writer did not receive output: %q", buf.String())
	}
}

func TestSysx_CommandBuilder_Run(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "run"}
	} else {
		cmd = "echo"
		args = []string{"run"}
	}
	if err := sysx.NewCommand(cmd).WithArgs(args...).Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
}

func TestSysx_CommandBuilder_Output(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	out, err := sysx.NewCommand("echo").WithArgs("output_test").Output()
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}
	if !strings.Contains(out, "output_test") {
		t.Errorf("Output() = %q; want to contain 'output_test'", out)
	}
}

// ///////////////////////////
// Section: RunCommand tests
// ///////////////////////////

func TestSysx_RunCommand(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "hello"}
	} else {
		cmd = "echo"
		args = []string{"hello"}
	}
	res := sysx.RunCommand(cmd, args...)
	if res == nil {
		t.Fatal("RunCommand returned nil")
	}
	if !res.IsSuccess() {
		t.Errorf("RunCommand(%q) failed: %v", cmd, res.Err())
	}
}

// ///////////////////////////
// Section: ExecCommandContext tests
// ///////////////////////////

func TestSysx_ExecCommandContext_Success(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "ctx"}
	} else {
		cmd = "echo"
		args = []string{"ctx"}
	}
	ctx := context.Background()
	if err := sysx.ExecCommandContext(ctx, cmd, args...); err != nil {
		t.Errorf("ExecCommandContext error = %v", err)
	}
}

func TestSysx_ExecCommandContext_Cancelled(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sysx.ExecCommandContext(ctx, "sleep", "10")
	if err == nil {
		t.Error("expected cancellation error, got nil")
	}
}

func TestSysx_ExecCommandContext_EmptyName(t *testing.T) {
	t.Parallel()
	err := sysx.ExecCommandContext(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

// ///////////////////////////
// Section: ExecOutputLines tests
// ///////////////////////////

func TestSysx_ExecOutputLines(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	lines, err := sysx.ExecOutputLines("bash", "-c", "printf 'a\\nb\\nc\\n'")
	if err != nil {
		t.Fatalf("ExecOutputLines error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("ExecOutputLines len = %d, want 3; lines=%v", len(lines), lines)
	}
	for i, want := range []string{"a", "b", "c"} {
		if lines[i] != want {
			t.Errorf("lines[%d] = %q, want %q", i, lines[i], want)
		}
	}
}

func TestSysx_ExecOutputLines_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := sysx.ExecOutputLines("")
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

// ///////////////////////////
// Section: ExecStreaming tests
// ///////////////////////////

func TestSysx_ExecStreaming(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	var out, errBuf bytes.Buffer
	err := sysx.ExecStreaming(&out, &errBuf, "echo", "streaming_test")
	if err != nil {
		t.Fatalf("ExecStreaming error = %v", err)
	}
	if !strings.Contains(out.String(), "streaming_test") {
		t.Errorf("ExecStreaming stdout = %q; want 'streaming_test'", out.String())
	}
}

func TestSysx_ExecStreaming_NilWriters(t *testing.T) {
	t.Parallel()
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "nil_writers"}
	} else {
		cmd = "echo"
		args = []string{"nil_writers"}
	}
	// nil writers should not panic
	if err := sysx.ExecStreaming(nil, nil, cmd, args...); err != nil {
		t.Errorf("ExecStreaming with nil writers error = %v", err)
	}
}

// ///////////////////////////
// Section: ExecAsync tests
// ///////////////////////////

func TestSysx_ExecAsync(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	cmd, err := sysx.ExecAsync("echo", "async_test")
	if err != nil {
		t.Fatalf("ExecAsync error = %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Errorf("cmd.Wait() error = %v", err)
	}
}

func TestSysx_ExecAsync_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := sysx.ExecAsync("")
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

// ///////////////////////////
// Section: ExecPipeline tests
// ///////////////////////////

func TestSysx_ExecPipeline(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	out, err := sysx.ExecPipeline(
		[]string{"echo", "hello pipeline"},
		[]string{"cat"},
	)
	if err != nil {
		t.Fatalf("ExecPipeline error = %v", err)
	}
	if !strings.Contains(out, "hello pipeline") {
		t.Errorf("ExecPipeline output = %q; want 'hello pipeline'", out)
	}
}

func TestSysx_ExecPipeline_Empty(t *testing.T) {
	t.Parallel()
	_, err := sysx.ExecPipeline()
	if err == nil {
		t.Error("ExecPipeline() with no commands should return error")
	}
}

func TestSysx_ExecPipeline_EmptyCommandName(t *testing.T) {
	t.Parallel()
	_, err := sysx.ExecPipeline([]string{})
	if err == nil {
		t.Error("ExecPipeline with empty command slice should return error")
	}
}

// ///////////////////////////
// Section: File reading tests
// ///////////////////////////

func TestSysx_ReadFile(t *testing.T) {
	t.Parallel()
	content := []byte("sysx read file test\n")
	f, err := os.CreateTemp("", "sysx_read_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Write(content)
	f.Close()
	defer os.Remove(path)

	got, err := sysx.ReadBytes(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("ReadFile = %q, want %q", got, content)
	}

	_, err = sysx.ReadBytes(path + "_nonexistent")
	if err == nil {
		t.Error("ReadFile on non-existent path should return error")
	}
}

func TestSysx_ReadFileString(t *testing.T) {
	t.Parallel()
	content := "hello from ReadFileString\n"
	f, err := os.CreateTemp("", "sysx_readstr_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString(content)
	f.Close()
	defer os.Remove(path)

	got, err := sysx.ReadString(path)
	if err != nil {
		t.Fatalf("ReadFileString error = %v", err)
	}
	if got != content {
		t.Errorf("ReadFileString = %q, want %q", got, content)
	}
}

func TestSysx_ReadLines(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_readlines_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("line1\nline2\nline3\n")
	f.Close()
	defer os.Remove(path)

	lines, err := sysx.ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("ReadLines len = %d, want 3", len(lines))
	}
	for i, want := range []string{"line1", "line2", "line3"} {
		if lines[i] != want {
			t.Errorf("lines[%d] = %q, want %q", i, lines[i], want)
		}
	}
}

func TestSysx_ReadLines_EmptyFile(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_readlines_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	lines, err := sysx.ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines empty file error = %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("ReadLines empty file len = %d, want 0", len(lines))
	}
}

func TestSysx_StreamLines(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_stream_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("alpha\nbeta\ngamma\n")
	f.Close()
	defer os.Remove(path)

	var collected []string
	err = sysx.StreamLines(path, func(line string) error {
		collected = append(collected, line)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamLines error = %v", err)
	}
	if len(collected) != 3 {
		t.Fatalf("StreamLines collected %d lines, want 3", len(collected))
	}
	for i, want := range []string{"alpha", "beta", "gamma"} {
		if collected[i] != want {
			t.Errorf("collected[%d] = %q, want %q", i, collected[i], want)
		}
	}
}

func TestSysx_StreamLines_HandlerError(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_stream_err_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("line1\nline2\nstop\nline4\n")
	f.Close()
	defer os.Remove(path)

	count := 0
	handlerErr := errors.New("stop here")
	err = sysx.StreamLines(path, func(line string) error {
		count++
		if line == "stop" {
			return handlerErr
		}
		return nil
	})
	if err != handlerErr {
		t.Errorf("StreamLines returned %v, want %v", err, handlerErr)
	}
	if count != 3 {
		t.Errorf("handler called %d times, want 3", count)
	}
}

// ///////////////////////////
// Section: File writing tests
// ///////////////////////////

func TestSysx_WriteFile(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "write")
	defer os.Remove(path)

	if err := sysx.WriteBytes(path, []byte("hello")); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "hello" {
		t.Errorf("WriteFile content = %q, want %q", got, "hello")
	}

	// Overwrite
	if err := sysx.WriteBytes(path, []byte("world")); err != nil {
		t.Fatalf("WriteFile overwrite error = %v", err)
	}
	got, _ = os.ReadFile(path)
	if string(got) != "world" {
		t.Errorf("WriteFile overwrite content = %q, want %q", got, "world")
	}
}

func TestSysx_WriteFileString(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "writestr")
	defer os.Remove(path)

	if err := sysx.WriteString(path, "string content"); err != nil {
		t.Fatalf("WriteFileString error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "string content" {
		t.Errorf("WriteFileString content = %q, want %q", got, "string content")
	}
}

func TestSysx_AppendFile(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "append")
	defer os.Remove(path)

	if err := sysx.AppendBytes(path, []byte("first")); err != nil {
		t.Fatalf("AppendFile (create) error = %v", err)
	}
	if err := sysx.AppendBytes(path, []byte("second")); err != nil {
		t.Fatalf("AppendFile (append) error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "firstsecond" {
		t.Errorf("AppendFile content = %q, want %q", got, "firstsecond")
	}
}

func TestSysx_AppendString(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "appendstr")
	defer os.Remove(path)

	sysx.AppendString(path, "A")
	sysx.AppendString(path, "B")
	sysx.AppendString(path, "C")
	got, _ := os.ReadFile(path)
	if string(got) != "ABC" {
		t.Errorf("AppendString content = %q, want %q", got, "ABC")
	}
}

func TestSysx_WriteLines(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "writelines")
	defer os.Remove(path)

	lines := []string{"first", "second", "third"}
	if err := sysx.WriteLines(path, lines); err != nil {
		t.Fatalf("WriteLines error = %v", err)
	}
	got, err := sysx.ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines after WriteLines error = %v", err)
	}
	if len(got) != len(lines) {
		t.Fatalf("ReadLines len = %d, want %d", len(got), len(lines))
	}
	for i, want := range lines {
		if got[i] != want {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want)
		}
	}
}

// ///////////////////////////
// Section: AtomicWriteFile tests
// ///////////////////////////

func TestSysx_AtomicWriteFile(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "atomic")
	defer os.Remove(path)

	data := []byte("atomic content")
	if err := sysx.AtomicWriteBytes(path, data); err != nil {
		t.Fatalf("AtomicWriteFile error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, data) {
		t.Errorf("AtomicWriteFile content = %q, want %q", got, data)
	}
}

func TestSysx_AtomicWriteFile_Overwrite(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "atomic_overwrite")
	defer os.Remove(path)

	sysx.WriteBytes(path, []byte("original"))
	if err := sysx.AtomicWriteBytes(path, []byte("replaced")); err != nil {
		t.Fatalf("AtomicWriteFile overwrite error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "replaced" {
		t.Errorf("AtomicWriteFile overwrite = %q, want %q", got, "replaced")
	}
}

func TestSysx_AtomicWriteFile_ConcurrentWrites(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "atomic_concurrent")
	defer os.Remove(path)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			data := []byte(strings.Repeat("x", n+1))
			sysx.AtomicWriteBytes(path, data)
		}(i)
	}
	wg.Wait()
	// File must be readable and non-empty after concurrent writes.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after concurrent AtomicWriteFile: %v", err)
	}
	if len(got) == 0 {
		t.Error("file is empty after concurrent AtomicWriteFile")
	}
}

// ///////////////////////////
// Section: SafeFileWriter tests
// ///////////////////////////

func TestSysx_SafeFileWriter_Write(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "safe_write")
	defer os.Remove(path)

	w := sysx.NewSafeFileWriter(path)
	if err := w.WriteBytes([]byte("line1\n")); err != nil {
		t.Fatalf("SafeFileWriter.Write error = %v", err)
	}
	if err := w.WriteString("line2\n"); err != nil {
		t.Fatalf("SafeFileWriter.WriteString error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "line1\nline2\n" {
		t.Errorf("SafeFileWriter content = %q, want %q", got, "line1\nline2\n")
	}
}

func TestSysx_SafeFileWriter_Overwrite(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "safe_overwrite")
	defer os.Remove(path)

	w := sysx.NewSafeFileWriter(path)
	w.WriteBytes([]byte("old"))
	if err := w.Overwrite([]byte("new")); err != nil {
		t.Fatalf("SafeFileWriter.Overwrite error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "new" {
		t.Errorf("SafeFileWriter.Overwrite content = %q, want %q", got, "new")
	}
}

func TestSysx_SafeFileWriter_Concurrent(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "safe_concurrent")
	defer os.Remove(path)

	w := sysx.NewSafeFileWriter(path)
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			w.WriteString("entry\n")
		}()
	}
	wg.Wait()

	lines, err := sysx.ReadLines(path)
	if err != nil {
		t.Fatalf("ReadLines after concurrent SafeFileWriter: %v", err)
	}
	if len(lines) != goroutines {
		t.Errorf("expected %d lines, got %d", goroutines, len(lines))
	}
}

// ///////////////////////////
// Section: WriteFileLocked tests
// ///////////////////////////

func TestSysx_WriteFileLocked(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "locked")
	defer os.Remove(path)

	if err := sysx.WriteBytesWithLocked(path, []byte("locked content")); err != nil {
		t.Fatalf("WriteFileLocked error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "locked content" {
		t.Errorf("WriteFileLocked content = %q, want %q", got, "locked content")
	}
}

func TestSysx_WriteFileLocked_Concurrent(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "locked_concurrent")
	defer os.Remove(path)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			sysx.WriteBytesWithLocked(path, []byte(strings.Repeat("y", n+1)))
		}(i)
	}
	wg.Wait()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after concurrent WriteFileLocked: %v", err)
	}
	if len(got) == 0 {
		t.Error("file is empty after concurrent WriteFileLocked")
	}
}

// ///////////////////////////
// Section: Test helpers
// ///////////////////////////

// tmpPath returns a temporary file path for the given test name suffix.
// The caller is responsible for removing the file.
func tmpPath(t *testing.T, suffix string) string {
	t.Helper()
	f, err := os.CreateTemp("", "sysx_"+suffix+"_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path) // remove so tests that create the file start fresh
	return path
}

// ///////////////////////////
// Section: Command accessor tests
// ///////////////////////////

func TestSysx_Command_Accessors(t *testing.T) {
	t.Parallel()
	cmd := sysx.NewCommand("go").
		WithArgs("build", "./...").
		WithDir("/tmp").
		WithEnv("GOOS=linux").
		WithTimeout(10 * time.Second)

	if cmd.Name() != "go" {
		t.Errorf("Name() = %q, want %q", cmd.Name(), "go")
	}
	if len(cmd.Args()) != 2 || cmd.Args()[0] != "build" {
		t.Errorf("Args() = %v, want [build ./...]", cmd.Args())
	}
	if cmd.Dir() != "/tmp" {
		t.Errorf("Dir() = %q, want %q", cmd.Dir(), "/tmp")
	}
	if len(cmd.Env()) != 1 || cmd.Env()[0] != "GOOS=linux" {
		t.Errorf("Env() = %v, want [GOOS=linux]", cmd.Env())
	}
	if cmd.Timeout() != 10*time.Second {
		t.Errorf("Timeout() = %v, want 10s", cmd.Timeout())
	}
}

// ///////////////////////////
// Section: CommandResult accessor tests
// ///////////////////////////

func TestSysx_CommandResult_Accessors(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	res := sysx.NewCommand("bash").WithArgs("-c", "echo stdout; echo stderr >&2").Execute()

	if res.Stdout() == "" {
		t.Error("Stdout() should not be empty")
	}
	if res.Stderr() == "" {
		t.Error("Stderr() should not be empty")
	}
	if res.ExitCode() != 0 {
		t.Errorf("ExitCode() = %d, want 0", res.ExitCode())
	}
	if res.Duration() <= 0 {
		t.Error("Duration() should be positive")
	}
	if res.Err() != nil {
		t.Errorf("Err() = %v, want nil", res.Err())
	}
	if !res.IsSuccess() {
		t.Error("Success() should be true")
	}
	combined := res.Combined()
	if !strings.Contains(combined, "stdout") || !strings.Contains(combined, "stderr") {
		t.Errorf("Combined() = %q; should contain both 'stdout' and 'stderr'", combined)
	}
}

func TestSysx_CommandResult_ErrorAccessor(t *testing.T) {
	t.Parallel()
	res := sysx.NewCommand("").Execute()
	if res.Err() == nil {
		t.Error("Err() should not be nil for empty command name")
	}
	if res.ExitCode() != -1 {
		t.Errorf("ExitCode() = %d, want -1 for empty command", res.ExitCode())
	}
	if res.IsSuccess() {
		t.Error("Success() should be false when Err() is non-nil")
	}
}

// ///////////////////////////
// Section: SafeFileWriter accessor tests
// ///////////////////////////

func TestSysx_SafeFileWriter_Accessors(t *testing.T) {
	t.Parallel()
	path := tmpPath(t, "sfw_accessor")
	defer os.Remove(path)

	w := sysx.NewSafeFileWriter(path).WithPerm(0o600)
	if w.Path() != path {
		t.Errorf("Path() = %q, want %q", w.Path(), path)
	}
	if w.Perm() != 0o600 {
		t.Errorf("Perm() = %o, want %o", w.Perm(), 0o600)
	}
}

// ///////////////////////////
// Section: Network utility tests
// ///////////////////////////

func TestSysx_IsIPv4(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"::1", false},
		{"2001:db8::1", false},
		{"not-an-ip", false},
		{"", false},
		{"300.1.1.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := sysx.IsIPv4(tt.ip)
			if got != tt.want {
				t.Errorf("IsIPv4(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestSysx_IsIPv6(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip   string
		want bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1", true},
		{"192.168.1.1", false},
		{"not-an-ip", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := sysx.IsIPv6(tt.ip)
			if got != tt.want {
				t.Errorf("IsIPv6(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestSysx_IsLocalIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"169.254.1.1", true},
		{"fc00::1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"not-an-ip", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := sysx.IsLocalIP(tt.ip)
			if got != tt.want {
				t.Errorf("IsLocalIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// func TestSysx_IsPortAvailable(t *testing.T) {
// 	t.Parallel()
// 	// Bind a port then verify IsPortAvailable reports it as taken.
// 	ln, err := net.Listen("tcp", "127.0.0.1:0")
// 	if err != nil {
// 		t.Skip("could not bind local port:", err)
// 	}
// 	defer ln.Close()
// 	port := ln.Addr().(*net.TCPAddr).Port
// 	if sysx.IsPortAvailable(port) {
// 		t.Errorf("IsPortAvailable(%d) = true for an already-bound port; want false", port)
// 	}
// }

func TestSysx_IsPortAvailable_FreePort(t *testing.T) {
	t.Parallel()
	// Find a free port by binding then releasing.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("could not bind local port:", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	// After closing, the port should be free.
	if !sysx.IsPortAvailable(port) {
		t.Logf("IsPortAvailable(%d) = false after close; OS may be holding TIME_WAIT state", port)
	}
}

func TestSysx_IsPortOpen_Localhost(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("could not bind local port:", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	if !sysx.IsPortOpen("127.0.0.1", port) {
		t.Errorf("IsPortOpen(127.0.0.1, %d) = false; want true for a bound port", port)
	}
}

func TestSysx_GetLocalIP(t *testing.T) {
	t.Parallel()
	ip, err := sysx.GetLocalIP()
	if err != nil {
		// Some CI environments have no non-loopback interfaces.
		t.Skipf("GetLocalIP returned error (may be expected in CI): %v", err)
	}
	if !sysx.IsIPv4(ip) {
		t.Errorf("GetLocalIP() = %q; want a valid IPv4 address", ip)
	}
}

func TestSysx_GetInterfaceIPs(t *testing.T) {
	t.Parallel()
	ips, err := sysx.GetInterfaceIPs()
	if err != nil {
		t.Fatalf("GetInterfaceIPs() error = %v", err)
	}
	_ = ips // may be empty in minimal containers
}

func TestSysx_ParseHostPort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		addr     string
		wantHost string
		wantPort int
		wantErr  bool
	}{
		{"localhost:8080", "localhost", 8080, false},
		{"192.168.1.1:443", "192.168.1.1", 443, false},
		{"[::1]:80", "::1", 80, false},
		{"example.com:0", "example.com", 0, false},
		{"no-port", "", 0, true},
		{"host:notanumber", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			host, port, err := sysx.ParseHostPort(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHostPort(%q) err = %v, wantErr %v", tt.addr, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if host != tt.wantHost {
					t.Errorf("ParseHostPort(%q) host = %q, want %q", tt.addr, host, tt.wantHost)
				}
				if port != tt.wantPort {
					t.Errorf("ParseHostPort(%q) port = %d, want %d", tt.addr, port, tt.wantPort)
				}
			}
		})
	}
}

func TestSysx_IsValidHost(t *testing.T) {
	t.Parallel()
	if !sysx.IsValidHost("127.0.0.1") {
		t.Error("IsValidHost(127.0.0.1) = false; want true")
	}
	if !sysx.IsValidHost("::1") {
		t.Error("IsValidHost(::1) = false; want true")
	}
	if !sysx.IsValidHost("localhost") {
		t.Skip("localhost DNS resolution failed; skipping")
	}
	if sysx.IsValidHost("this.host.definitely.does.not.exist.invalid") {
		t.Error("IsValidHost for nonexistent host = true; want false")
	}
}

func TestSysx_IsValidURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"http://localhost:8080/path", true},
		{"ftp://files.example.com", true},
		{"not-a-url", false},
		{"//missing-scheme.com", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := sysx.IsValidURL(tt.url)
			if got != tt.want {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestSysx_CheckTCPConnection_InvalidPort(t *testing.T) {
	t.Parallel()
	if err := sysx.CheckTCPConn("localhost", 0, time.Second); err == nil {
		t.Error("CheckTCPConnection with port 0 should return error")
	}
	if err := sysx.CheckTCPConn("localhost", 65536, time.Second); err == nil {
		t.Error("CheckTCPConnection with port 65536 should return error")
	}
}

func TestSysx_CheckTCPConnection_Success(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("could not bind local port:", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	if err := sysx.CheckTCPConn("127.0.0.1", port, 3*time.Second); err != nil {
		t.Errorf("CheckTCPConnection(127.0.0.1, %d) error = %v; want nil", port, err)
	}
}

func TestSysx_CheckTCPConnection_Failure(t *testing.T) {
	t.Parallel()
	err := sysx.CheckTCPConn("127.0.0.1", 1, 500*time.Millisecond)
	if err == nil {
		t.Skip("port 1 appears open (unexpected); skipping failure test")
	}
}

func TestSysx_PingHost_DoesNotPanic(t *testing.T) {
	t.Parallel()
	// Just verify the function doesn't panic.
	_ = sysx.PingHost("127.0.0.1")
}

// ///////////////////////////
// Section: Directory utility tests
// ///////////////////////////

func TestSysx_CreateDir(t *testing.T) {
	t.Parallel()
	base, err := os.MkdirTemp("", "sysx_mkdir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(base)

	nested := base + "/a/b/c"
	if err := sysx.CreateDir(nested); err != nil {
		t.Fatalf("CreateDir(%q) error = %v", nested, err)
	}
	if !sysx.DirExists(nested) {
		t.Errorf("CreateDir(%q): directory does not exist after creation", nested)
	}
	// Idempotent: calling again should be a no-op.
	if err := sysx.CreateDir(nested); err != nil {
		t.Errorf("CreateDir(%q) second call error = %v", nested, err)
	}
}

func TestSysx_RemoveDir(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "sysx_rmdir_*")
	if err != nil {
		t.Fatal(err)
	}
	// Create a nested structure to verify recursive removal.
	nested := dir + "/sub/dir"
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := sysx.RemoveDir(dir); err != nil {
		t.Fatalf("RemoveDir(%q) error = %v", dir, err)
	}
	if sysx.DirExists(dir) {
		t.Errorf("RemoveDir(%q): directory still exists after removal", dir)
	}
	// Calling on a non-existent path should not error.
	if err := sysx.RemoveDir(dir + "_nonexistent"); err != nil {
		t.Errorf("RemoveDir non-existent path: got error %v, want nil", err)
	}
}

func TestSysx_ListDir(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "sysx_listdir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create some entries.
	names := []string{"alpha.txt", "beta.txt", "gamma"}
	for _, name := range names {
		path := dir + "/" + name
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := sysx.ListDir(dir)
	if err != nil {
		t.Fatalf("ListDir(%q) error = %v", dir, err)
	}
	if len(got) != len(names) {
		t.Errorf("ListDir len = %d; want %d", len(got), len(names))
	}
}

func TestSysx_ListDir_Error(t *testing.T) {
	t.Parallel()
	_, err := sysx.ListDir("/nonexistent_sysx_dir_xyz")
	if err == nil {
		t.Error("ListDir on non-existent path should return error")
	}
}

func TestSysx_ListDirFiles(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "sysx_listfiles_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a file and a subdirectory.
	if err := os.WriteFile(dir+"/file.txt", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir+"/subdir", 0o755); err != nil {
		t.Fatal(err)
	}

	files, err := sysx.ListDirFiles(dir)
	if err != nil {
		t.Fatalf("ListDirFiles(%q) error = %v", dir, err)
	}
	if len(files) != 1 || files[0] != "file.txt" {
		t.Errorf("ListDirFiles = %v; want [file.txt]", files)
	}
}

func TestSysx_ListDirDirs(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "sysx_listdirs_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a file and a subdirectory.
	if err := os.WriteFile(dir+"/file.txt", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir+"/subdir", 0o755); err != nil {
		t.Fatal(err)
	}

	dirs, err := sysx.ListDirDirs(dir)
	if err != nil {
		t.Fatalf("ListDirDirs(%q) error = %v", dir, err)
	}
	if len(dirs) != 1 || dirs[0] != "subdir" {
		t.Errorf("ListDirDirs = %v; want [subdir]", dirs)
	}
}

// ///////////////////////////
// Section: Path helper tests
// ///////////////////////////

func TestSysx_BaseName(t *testing.T) {
	t.Parallel()
	// Relative-path cases work identically on all platforms.
	cases := []struct {
		path string
		want string
	}{
		{"file.txt", "file.txt"},
		{"", "."},
	}
	// Unix-style absolute paths are only valid on non-Windows systems;
	// filepath.Base on Windows uses '\' as the separator and would not
	// split on '/', returning the full path instead of the base name.
	if runtime.GOOS != "windows" {
		cases = append(cases, []struct{ path, want string }{
			{"/etc/hosts", "hosts"},
			{"/usr/bin/", "bin"},
		}...)
	}
	for _, tc := range cases {
		got := sysx.BaseName(tc.path)
		if got != tc.want {
			t.Errorf("BaseName(%q) = %q; want %q", tc.path, got, tc.want)
		}
	}
}

func TestSysx_DirName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want string
	}{
		{"/etc/hosts", "/etc"},
		{"/usr/bin/git", "/usr/bin"},
		{"file.txt", "."},
		{"", "."},
	}
	for _, tc := range cases {
		// filepath.Clean (called internally by filepath.Dir) converts '/' to
		// the OS separator on Windows, so normalise to forward slashes before
		// comparing.
		got := filepath.ToSlash(sysx.DirName(tc.path))
		if got != tc.want {
			t.Errorf("DirName(%q) = %q; want %q", tc.path, got, tc.want)
		}
	}
}

func TestSysx_Ext(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want string
	}{
		{"archive.tar.gz", ".gz"},
		{"/etc/hosts", ""},
		{"README.md", ".md"},
		{"noext", ""},
		{"", ""},
	}
	for _, tc := range cases {
		got := sysx.Ext(tc.path)
		if got != tc.want {
			t.Errorf("Ext(%q) = %q; want %q", tc.path, got, tc.want)
		}
	}
}

func TestSysx_AbsPath(t *testing.T) {
	t.Parallel()
	// A relative path must produce a platform-absolute result.
	rel, err := sysx.AbsPath(".")
	if err != nil {
		t.Fatalf("AbsPath(.) error = %v", err)
	}
	if !filepath.IsAbs(rel) {
		t.Errorf("AbsPath(.) = %q; want absolute path", rel)
	}
	// An already-absolute path (the OS temp dir) should round-trip as absolute.
	// We use os.TempDir() because it is guaranteed to exist and be absolute on
	// all platforms (unlike hardcoded Unix paths such as /etc/hosts).
	tmp := os.TempDir()
	abs, err := sysx.AbsPath(tmp)
	if err != nil {
		t.Fatalf("AbsPath(%q) error = %v", tmp, err)
	}
	if !filepath.IsAbs(abs) {
		t.Errorf("AbsPath(%q) = %q; want absolute path", tmp, abs)
	}
}

func TestSysx_JoinPath(t *testing.T) {
	t.Parallel()
	// filepath.Join uses the OS path separator; normalise to forward slashes
	// before comparing so the test passes on both Unix and Windows.
	got := sysx.JoinPath("/usr", "local", "bin")
	if filepath.ToSlash(got) != "/usr/local/bin" {
		t.Errorf("JoinPath = %q; want %q", got, "/usr/local/bin")
	}
	got = sysx.JoinPath("a", "b", "c")
	if filepath.ToSlash(got) != "a/b/c" {
		t.Errorf("JoinPath(a,b,c) = %q; want %q", got, "a/b/c")
	}
}

func TestSysx_CleanPath(t *testing.T) {
	t.Parallel()
	// filepath.Clean replaces '/' with the OS separator on Windows; normalise
	// before comparing.
	got := sysx.CleanPath("/usr//local/./bin/../lib")
	if filepath.ToSlash(got) != "/usr/local/lib" {
		t.Errorf("CleanPath = %q; want %q", got, "/usr/local/lib")
	}
	if sysx.CleanPath("") != "." {
		t.Errorf("CleanPath(\"\") = %q; want %q", sysx.CleanPath(""), ".")
	}
}

func TestSysx_SplitPath(t *testing.T) {
	t.Parallel()
	dir, file := sysx.SplitPath("/usr/local/bin/git")
	// filepath.Split preserves the original separators in the dir component,
	// but normalise with ToSlash for consistency across platforms.
	if filepath.ToSlash(dir) != "/usr/local/bin/" {
		t.Errorf("SplitPath dir = %q; want %q", dir, "/usr/local/bin/")
	}
	if file != "git" {
		t.Errorf("SplitPath file = %q; want %q", file, "git")
	}
}

// ///////////////////////////
// Section: File info tests
// ///////////////////////////

func TestSysx_FileMode(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("file mode bits are approximations on Windows")
	}
	f, err := os.CreateTemp("", "sysx_mode_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	// Set an explicit mode.
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	mode, err := sysx.FileMode(path)
	if err != nil {
		t.Fatalf("FileMode error = %v", err)
	}
	if mode != 0o644 {
		t.Errorf("FileMode = %o; want %o", mode, 0o644)
	}

	_, err = sysx.FileMode(path + "_nonexistent")
	if err == nil {
		t.Error("FileMode on non-existent path should return error")
	}
}

func TestSysx_FileModTime(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_modtime_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	before := time.Now().Add(-time.Second)
	mt, err := sysx.FileModTime(path)
	if err != nil {
		t.Fatalf("FileModTime error = %v", err)
	}
	if mt.Before(before) {
		t.Errorf("FileModTime = %v; want after %v", mt, before)
	}

	_, err = sysx.FileModTime(path + "_nonexistent")
	if err == nil {
		t.Error("FileModTime on non-existent path should return error")
	}
}

// ///////////////////////////
// Section: I/O utility tests
// ///////////////////////////

func TestSysx_CountLines(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_countlines_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("line1\nline2\nline3\n")
	f.Close()
	defer os.Remove(path)

	n, err := sysx.CountLines(path)
	if err != nil {
		t.Fatalf("CountLines error = %v", err)
	}
	if n != 3 {
		t.Errorf("CountLines = %d; want 3", n)
	}
}

func TestSysx_CountLines_Empty(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_countlines_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	n, err := sysx.CountLines(path)
	if err != nil {
		t.Fatalf("CountLines empty file error = %v", err)
	}
	if n != 0 {
		t.Errorf("CountLines empty = %d; want 0", n)
	}
}

func TestSysx_CountLines_Error(t *testing.T) {
	t.Parallel()
	_, err := sysx.CountLines("/nonexistent_sysx_file_xyz")
	if err == nil {
		t.Error("CountLines on non-existent file should return error")
	}
}

func TestSysx_Head(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_head_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("line1\nline2\nline3\nline4\nline5\n")
	f.Close()
	defer os.Remove(path)

	lines, err := sysx.Head(path, 3)
	if err != nil {
		t.Fatalf("Head error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("Head len = %d; want 3", len(lines))
	}
	for i, want := range []string{"line1", "line2", "line3"} {
		if lines[i] != want {
			t.Errorf("Head[%d] = %q; want %q", i, lines[i], want)
		}
	}
}

func TestSysx_Head_FewerLinesThanN(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_head_short_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("only\ntwo\n")
	f.Close()
	defer os.Remove(path)

	lines, err := sysx.Head(path, 100)
	if err != nil {
		t.Fatalf("Head error = %v", err)
	}
	if len(lines) != 2 {
		t.Errorf("Head len = %d; want 2", len(lines))
	}
}

func TestSysx_Head_Zero(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_head_zero_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("line1\nline2\n")
	f.Close()
	defer os.Remove(path)

	lines, err := sysx.Head(path, 0)
	if err != nil {
		t.Fatalf("Head(0) error = %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("Head(0) len = %d; want 0", len(lines))
	}
}

func TestSysx_CopyFile(t *testing.T) {
	t.Parallel()
	src, err := os.CreateTemp("", "sysx_copy_src_*")
	if err != nil {
		t.Fatal(err)
	}
	srcPath := src.Name()
	content := []byte("copy file content")
	src.Write(content)
	src.Close()
	defer os.Remove(srcPath)

	dstPath := srcPath + "_dst"
	defer os.Remove(dstPath)

	if err := sysx.CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile error = %v", err)
	}
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("ReadFile after CopyFile error = %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("CopyFile content = %q; want %q", got, content)
	}
}

func TestSysx_CopyFile_SrcNotFound(t *testing.T) {
	t.Parallel()
	err := sysx.CopyFile("/nonexistent_sysx_src_xyz", "/tmp/sysx_dst_xyz")
	if err == nil {
		t.Error("CopyFile with non-existent src should return error")
	}
}

func TestSysx_TruncateFile(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_trunc_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("some content here")
	f.Close()
	defer os.Remove(path)

	if err := sysx.TruncateFile(path, 4); err != nil {
		t.Fatalf("TruncateFile error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "some" {
		t.Errorf("TruncateFile content = %q; want %q", got, "some")
	}
}

func TestSysx_TruncateFile_ToZero(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "sysx_trunc_zero_*")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.WriteString("data")
	f.Close()
	defer os.Remove(path)

	if err := sysx.TruncateFile(path, 0); err != nil {
		t.Fatalf("TruncateFile(0) error = %v", err)
	}
	got, _ := os.ReadFile(path)
	if len(got) != 0 {
		t.Errorf("TruncateFile(0) content = %q; want empty", got)
	}
}

func TestSysx_Move(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_move_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	content := "move me"

	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := sysx.Move(src, dst); err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	if sysx.FileExists(src) {
		t.Error("source file still exists after Move")
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("could not read destination file: %v", err)
	}
	if string(got) != content {
		t.Errorf("destination content = %q; want %q", string(got), content)
	}
}

func TestSysx_Touch(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_touch_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "new_file.txt")

	// Test creation
	if err := sysx.Touch(path); err != nil {
		t.Fatalf("Touch creation failed: %v", err)
	}
	if !sysx.FileExists(path) {
		t.Error("Touch did not create file")
	}

	// Test update
	oldFi, _ := os.Stat(path)
	time.Sleep(10 * time.Millisecond) // Ensure time difference
	if err := sysx.Touch(path); err != nil {
		t.Fatalf("Touch update failed: %v", err)
	}
	newFi, _ := os.Stat(path)

	if !newFi.ModTime().After(oldFi.ModTime()) {
		t.Errorf("Touch did not update mod time: %v -> %v", oldFi.ModTime(), newFi.ModTime())
	}
}

func TestSysx_IsBinary(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_binary_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	txtPath := filepath.Join(dir, "text.txt")
	binPath := filepath.Join(dir, "bin.dat")

	if err := os.WriteFile(txtPath, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte{0x48, 0x00, 0x4c, 0x4c, 0x4f}, 0644); err != nil {
		t.Fatal(err)
	}

	isBin, err := sysx.IsBinary(txtPath)
	if err != nil {
		t.Fatal(err)
	}
	if isBin {
		t.Errorf("IsBinary(%q) = true; want false", txtPath)
	}

	isBin, err = sysx.IsBinary(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if !isBin {
		t.Errorf("IsBinary(%q) = false; want true", binPath)
	}
}

func TestSysx_Hashing(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_hash_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "data.txt")
	content := "replify"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// MD5 of "replify" is 5f7be9021c5932d1e511859f201c86d8
	wantMD5 := "5f7be9021c5932d1e511859f201c86d8"
	gotMD5, err := sysx.FileMD5(path)
	if err != nil {
		t.Fatal(err)
	}
	if gotMD5 != wantMD5 {
		t.Errorf("MD5 = %q; want %q", gotMD5, wantMD5)
	}

	// SHA256 of "replify" is 89ac3e4273d8ce6ccab52871f2bbdffcdee4399617b0eff0f2b94aadfafe4f0e
	wantSHA256 := "89ac3e4273d8ce6ccab52871f2bbdffcdee4399617b0eff0f2b94aadfafe4f0e"
	gotSHA256, err := sysx.FileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	if gotSHA256 != wantSHA256 {
		t.Errorf("SHA256 = %q; want %q", gotSHA256, wantSHA256)
	}
}

func TestSysx_CopyDir(t *testing.T) {
	parent, err := os.MkdirTemp("", "sysx_copydir_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(parent)

	src := filepath.Join(parent, "src")
	dst := filepath.Join(parent, "dst")

	// Create structure:
	// src/
	//   file1.txt
	//   subdir/
	//     file2.txt
	if err := os.MkdirAll(filepath.Join(src, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(src, "file1.txt"), []byte("one"), 0644)
	os.WriteFile(filepath.Join(src, "subdir", "file2.txt"), []byte("two"), 0644)

	// Test symlink replication (on non-Windows)
	if runtime.GOOS != "windows" {
		os.Symlink("file1.txt", filepath.Join(src, "link.txt"))
	}

	if err := sysx.CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// Verify
	files := []string{
		filepath.Join(dst, "file1.txt"),
		filepath.Join(dst, "subdir", "file2.txt"),
	}
	for _, f := range files {
		if !sysx.FileExists(f) {
			t.Errorf("copied file missing: %s", f)
		}
	}

	if runtime.GOOS != "windows" {
		link := filepath.Join(dst, "link.txt")
		if !sysx.IsSymlink(link) {
			t.Errorf("symlink was not replicated: %s", link)
		}
	}
}

func TestSysx_ClearDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_cleardir_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "f1.txt"), []byte("data"), 0644)
	os.Mkdir(filepath.Join(dir, "d1"), 0755)

	if err := sysx.ClearDir(dir); err != nil {
		t.Fatalf("ClearDir failed: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("ClearDir did not empty directory; remaining entries: %d", len(entries))
	}
	if !sysx.DirExists(dir) {
		t.Error("ClearDir removed the directory itself")
	}
}

func TestSysx_IsEmpty(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_isempty_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	empty, _ := sysx.IsDirEmpty(dir)
	if !empty {
		t.Errorf("IsEmpty(%q) = false; want true", dir)
	}

	os.WriteFile(filepath.Join(dir, "not_empty.txt"), []byte("x"), 0644)
	empty, _ = sysx.IsDirEmpty(dir)
	if empty {
		t.Errorf("IsEmpty(%q) = true; want false", dir)
	}
}

func TestSysx_ReadTail(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_tail_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "lines.txt")
	lines := []string{"line1", "line2", "line3", "line4", "line5"}
	if err := sysx.WriteLines(path, lines); err != nil {
		t.Fatal(err)
	}

	// Test tail 2
	got, err := sysx.Tail(path, 2)
	if err != nil {
		t.Fatalf("ReadTail failed: %v", err)
	}
	want := []string{"line4", "line5"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadTail(2) = %v; want %v", got, want)
	}

	// Test tail more than exists
	got, err = sysx.Tail(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, lines) {
		t.Errorf("ReadTail(10) = %v; want %v", got, lines)
	}
}

func TestSysx_CountLines_Optimized(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_count_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "many_lines.txt")
	numLines := 1000
	lines := make([]string, numLines)
	for i := 0; i < numLines; i++ {
		lines[i] = "this is a line"
	}
	if err := sysx.WriteLines(path, lines); err != nil {
		t.Fatal(err)
	}

	got, err := sysx.CountLines(path)
	if err != nil {
		t.Fatalf("CountLines failed: %v", err)
	}
	if got != numLines {
		t.Errorf("CountLines = %d; want %d", got, numLines)
	}
}

func TestSysx_FileOpenFlags(t *testing.T) {
	// Test String method
	if sysx.CWA.String() != "CWA (CREATE|WRONLY|APPEND)" {
		t.Errorf("CWA.String() = %q", sysx.CWA.String())
	}
	if sysx.RO.String() != "RO (RDONLY)" {
		t.Errorf("RO.String() = %q", sysx.RO.String())
	}
	if sysx.FileOpenFlags(999).String() != "Unknown (Custom Flags)" {
		t.Errorf("Unknown string error")
	}

	// Test IsValid method
	if !sysx.CWA.IsValid() {
		t.Error("CWA should be valid")
	}
	if sysx.FileOpenFlags(999).IsValid() {
		t.Error("999 should be invalid")
	}
}

func TestSysx_OpenFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_openfile_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "test.log")

	// Test CWA (Create, WriteOnly, Append)
	f, err := sysx.OpenFile(path, sysx.CWA, 0644)
	if err != nil {
		t.Fatalf("OpenFile CWA failed: %v", err)
	}
	f.WriteString("hello")
	f.Close()

	// Test RO (ReadOnly)
	f, err = sysx.OpenFile(path, sysx.RO, 0)
	if err != nil {
		t.Fatalf("OpenFile RO failed: %v", err)
	}
	buf := make([]byte, 5)
	n, _ := f.Read(buf)
	f.Close()
	if string(buf[:n]) != "hello" {
		t.Errorf("expected 'hello', got %q", string(buf[:n]))
	}
}

func TestSysx_AtomicOpenFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "sysx_atomicopen_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "locked.txt")

	// Test exclusive write lock (CWA)
	f, err := sysx.AtomicOpenFile(path, sysx.CWA, 0644)
	if err != nil {
		t.Fatalf("AtomicOpenFile CWA failed: %v", err)
	}
	_, _ = f.WriteString("atomic content")
	f.Close() // releases lock

	// Test shared read lock (RO)
	f, err = sysx.AtomicOpenFile(path, sysx.RO, 0)
	if err != nil {
		t.Fatalf("AtomicOpenFile RO failed: %v", err)
	}
	buf := make([]byte, 14)
	n, _ := f.Read(buf)
	f.Close()
	if string(buf[:n]) != "atomic content" {
		t.Errorf("expected 'atomic content', got %q", string(buf[:n]))
	}

	// Test concurrency: multiple goroutines writing sequentially via CRWT
	const goroutines = 5
	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			f, err := sysx.AtomicOpenFile(path, sysx.CRWT, 0644)
			if err != nil {
				return
			}
			f.WriteString(strings.Repeat("z", n+1))
			f.Close()
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
