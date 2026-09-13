// Package testutil contains isolation helpers imported only by test files.
package testutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/marcus/sidecar/internal/hostexec"
	"github.com/marcus/sidecar/internal/tmuxcmd"
	"github.com/marcus/sidecar/internal/userhome"
)

type blockedTransport struct{}

func (blockedTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("fixture network unavailable")
}

// Home selects an explicit fixture lookup without changing the environment.
// Like testing.T.Setenv, it must not be used from parallel tests.
func Home(t *testing.T, root string) {
	t.Helper()
	if !filepath.IsAbs(root) {
		t.Fatal("fixture home must be absolute")
	}
	previous := userhome.Dir
	userhome.Dir = func() (string, error) { return root, nil }
	t.Cleanup(func() { userhome.Dir = previous })
}

// Env selects one explicit data-directory override for a nonparallel fixture.
func Env(t *testing.T, key, value string) {
	t.Helper()
	previous := userhome.Getenv
	userhome.Getenv = func(name string) string {
		if name == key {
			return value
		}
		return previous(name)
	}
	t.Cleanup(func() { userhome.Getenv = previous })
}

// Run gives a test process a disposable profile and inert Git configuration.
// Production code keeps its original home lookup; protected environment values
// must remain exactly unchanged, including unset versus empty values.
func Run(m *testing.M) int {
	protected := []string{"HOME", "USERPROFILE", "CODEX_HOME"}
	type value struct {
		text string
		set  bool
	}
	before := map[string]value{}
	for _, key := range protected {
		text, set := os.LookupEnv(key)
		before[key] = value{text, set}
	}
	root, err := os.MkdirTemp("", "sidecar-test-profile-")
	if err != nil {
		panic(err)
	}
	previous := userhome.Dir
	userhome.Dir = func() (string, error) { return root, nil }
	previousEnv := userhome.Getenv
	userhome.Getenv = func(string) string { return "" }
	previousTransport := http.DefaultTransport
	http.DefaultTransport = blockedTransport{}
	previousHostContext := hostexec.CommandContext
	hostexec.CommandContext = func(_ context.Context, name string, args ...string) *exec.Cmd {
		return &exec.Cmd{Args: append([]string{name}, args...), Err: exec.ErrNotFound}
	}
	previousHostCommand, previousLookPath := hostexec.Command, hostexec.LookPath
	hostexec.Command = func(name string, args ...string) *exec.Cmd {
		return &exec.Cmd{Args: append([]string{name}, args...), Err: exec.ErrNotFound}
	}
	hostexec.LookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	previousContextCommand := tmuxcmd.CommandContext
	tmuxcmd.CommandContext = func(_ context.Context, args ...string) *exec.Cmd {
		return &exec.Cmd{Args: append([]string{"tmux"}, args...), Err: errors.New("fixture tmux session not found")}
	}
	previousCommand := tmuxcmd.Command
	tmuxcmd.Command = func(args ...string) *exec.Cmd {
		return &exec.Cmd{Args: append([]string{"tmux"}, args...), Err: errors.New("fixture tmux session not found")}
	}
	for key, val := range map[string]string{"GIT_CONFIG_GLOBAL": os.DevNull, "GIT_CONFIG_NOSYSTEM": "1", "GIT_TERMINAL_PROMPT": "0", "GIT_CONFIG_COUNT": "0", "TERMIMG_BYPASS_DETECTION": "halfblocks"} {
		if err := os.Setenv(key, val); err != nil {
			panic(err)
		}
	}
	code := m.Run()
	for _, key := range protected {
		text, set := os.LookupEnv(key)
		if before[key] != (value{text, set}) {
			fmt.Fprintln(os.Stderr, "test changed protected environment variable:", key)
			code = 1
		}
	}
	userhome.Dir = previous
	userhome.Getenv = previousEnv
	tmuxcmd.Command = previousCommand
	tmuxcmd.CommandContext = previousContextCommand
	hostexec.Command, hostexec.LookPath = previousHostCommand, previousLookPath
	hostexec.CommandContext = previousHostContext
	http.DefaultTransport = previousTransport
	if err := os.RemoveAll(root); err != nil {
		fmt.Fprintln(os.Stderr, "remove test profile:", err)
		code = 1
	}
	return code
}
