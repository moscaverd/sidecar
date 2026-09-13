// Package hostexec provides the boundary for installed tools and host actions.
package hostexec

import "os/exec"

// Production uses the standard process functions. Tests substitute inert commands
// before callbacks can inspect or start installed tools, editors, or browsers.
var Command = exec.Command
var CommandContext = exec.CommandContext
var LookPath = exec.LookPath
