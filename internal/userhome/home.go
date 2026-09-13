// Package userhome provides the home lookup shared by configuration and adapters.
package userhome

import "os"

// Dir preserves os.UserHomeDir's production behavior. Tests replace this lookup
// before constructing adapters so fixtures never need to change process HOME.
var Dir = os.UserHomeDir

// Getenv preserves environment overrides in production. Tests provide explicit
// data-directory values without changing the host environment.
var Getenv = os.Getenv
