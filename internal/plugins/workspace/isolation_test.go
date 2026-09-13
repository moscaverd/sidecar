package workspace

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/marcus/sidecar/internal/testutil"
	"github.com/marcus/sidecar/internal/tmuxcmd"
)

var errFixtureTmux = errors.New("fixture tmux session not found")

func failedTmuxCommand(args ...string) *exec.Cmd {
	// Cmd.Err causes Start/Run/Output to fail before starting any process.
	return &exec.Cmd{Args: append([]string{"tmux"}, args...), Err: errFixtureTmux}
}

func TestMain(m *testing.M) {
	tmuxcmd.Command = failedTmuxCommand
	os.Exit(testutil.Run(m))
}

func captureTmuxCommands(t *testing.T) *[][]string {
	t.Helper()
	previous := tmuxcmd.Command
	calls := [][]string{}
	tmuxcmd.Command = func(args ...string) *exec.Cmd {
		calls = append(calls, append([]string(nil), args...))
		return failedTmuxCommand(args...)
	}
	t.Cleanup(func() { tmuxcmd.Command = previous })
	return &calls
}
