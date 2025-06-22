package testdata

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// MockCommandExecutor mocks command execution for testing
type MockCommandExecutor struct {
	Commands []MockCommand
	Index    int
}

type MockCommand struct {
	Name     string
	Args     []string
	Stdout   string
	Stderr   string
	ExitCode int
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Commands: []MockCommand{},
		Index:    0,
	}
}

func (m *MockCommandExecutor) AddCommand(name string, args []string, stdout, stderr string, exitCode int) {
	m.Commands = append(m.Commands, MockCommand{
		Name:     name,
		Args:     args,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	})
}

func (m *MockCommandExecutor) Execute(name string, args ...string) error {
	if m.Index >= len(m.Commands) {
		return fmt.Errorf("unexpected command: %s %s", name, strings.Join(args, " "))
	}

	cmd := m.Commands[m.Index]
	m.Index++

	if cmd.Name != name {
		return fmt.Errorf("expected command %s, got %s", cmd.Name, name)
	}

	if cmd.ExitCode != 0 {
		return fmt.Errorf("exit status %d", cmd.ExitCode)
	}

	return nil
}

// MockFileSystem mocks file system operations
type MockFileSystem struct {
	Files map[string]*MockFile
}

type MockFile struct {
	Content string
	Info    os.FileInfo
	Error   error
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string]*MockFile),
	}
}

func (m *MockFileSystem) AddFile(path string, content string, info os.FileInfo) {
	m.Files[path] = &MockFile{
		Content: content,
		Info:    info,
	}
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if file, ok := m.Files[path]; ok {
		if file.Error != nil {
			return nil, file.Error
		}
		return file.Info, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if file, ok := m.Files[path]; ok {
		if file.Error != nil {
			return nil, file.Error
		}
		return []byte(file.Content), nil
	}
	return nil, os.ErrNotExist
}

// MockFileInfo implements os.FileInfo interface
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime string
	isDir   bool
}

func NewMockFileInfo(name string, size int64, isDir bool) *MockFileInfo {
	return &MockFileInfo{
		name:  name,
		size:  size,
		isDir: isDir,
		mode:  0644,
	}
}

func (m *MockFileInfo) Name() string      { return m.name }
func (m *MockFileInfo) Size() int64       { return m.size }
func (m *MockFileInfo) Mode() os.FileMode { return m.mode }
func (m *MockFileInfo) ModTime() string   { return m.modTime }
func (m *MockFileInfo) IsDir() bool       { return m.isDir }
func (m *MockFileInfo) Sys() interface{}  { return nil }

// MockCmd implements a testable version of exec.Cmd
type MockCmd struct {
	name   string
	args   []string
	stdout io.Writer
	stderr io.Writer
	output string
	err    error
}

func (m *MockCmd) Run() error {
	return m.err
}

func (m *MockCmd) Output() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(m.output), nil
}

func (m *MockCmd) StdoutPipe() (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(strings.NewReader(m.output)), nil
}

func (m *MockCmd) StderrPipe() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockCmd) Start() error {
	return m.err
}

func (m *MockCmd) Wait() error {
	return m.err
}

// CommandRunner interface for dependency injection
type CommandRunner interface {
	Command(name string, args ...string) *exec.Cmd
	LookPath(file string) (string, error)
}

// RealCommandRunner implements CommandRunner using real exec package
type RealCommandRunner struct{}

func (r *RealCommandRunner) Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func (r *RealCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// MockCommandRunner implements CommandRunner for testing
type MockCommandRunner struct {
	Commands map[string]*MockCmd
	Paths    map[string]string
}

func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		Commands: make(map[string]*MockCmd),
		Paths:    make(map[string]string),
	}
}

func (m *MockCommandRunner) AddCommand(name string, args []string, output string, err error) {
	key := name + " " + strings.Join(args, " ")
	m.Commands[key] = &MockCmd{
		name:   name,
		args:   args,
		output: output,
		err:    err,
	}
}

func (m *MockCommandRunner) AddPath(file string, path string) {
	m.Paths[file] = path
}

func (m *MockCommandRunner) Command(name string, args ...string) *exec.Cmd {
	// This is simplified - in real tests you'd need to return a properly mocked exec.Cmd
	return exec.Command("echo", "mocked")
}

func (m *MockCommandRunner) LookPath(file string) (string, error) {
	if path, ok := m.Paths[file]; ok {
		return path, nil
	}
	return "", fmt.Errorf("executable file not found in $PATH")
}
