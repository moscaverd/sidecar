// Package tmuxcmd centralizes the process boundary shared by terminal features.
package tmuxcmd

import (
	"context"
	"os/exec"
)

// Command preserves the production tmux invocation. Tests replace this boundary
// before running any terminal behavior so fixture session names stay synthetic.
var Command = func(args ...string) *exec.Cmd { return exec.Command("tmux", args...) }

// CommandContext is the cancellable variant of the same tmux boundary.
var CommandContext = func(ctx context.Context, args ...string) *exec.Cmd { return exec.CommandContext(ctx, "tmux", args...) }
