package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcus/sidecar/internal/config"
	"github.com/marcus/sidecar/internal/hostexec"
	"github.com/marcus/sidecar/internal/state"
	"github.com/marcus/sidecar/internal/tmuxcmd"
	"github.com/marcus/sidecar/internal/userhome"
	"github.com/marcus/sidecar/internal/version"
)

func TestMain(m *testing.M) { os.Exit(Run(m)) }

func TestDefaultPersistenceUsesOnlyFixtureProfile(t *testing.T) {
	root, err := userhome.Dir()
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(root, ".config", "sidecar")
	cfg := config.Default()
	cfg.Features.Flags["fixture-only"] = true
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load()
	if err != nil || !loaded.Features.Flags["fixture-only"] {
		t.Fatalf("fixture configuration round trip failed: %v", err)
	}
	if config.ConfigPath() != filepath.Join(expected, "config.json") {
		t.Fatal("configuration escaped fixture profile")
	}
	if err := state.Init(); err != nil {
		t.Fatal(err)
	}
	if err := state.SetGitDiffMode("side-by-side"); err != nil {
		t.Fatal(err)
	}
	entry := &version.CacheEntry{LatestVersion: "v2.0.0", CurrentVersion: "v1.0.0", CheckedAt: time.Now().Truncate(time.Second), HasUpdate: true}
	if err := version.SaveCache(entry); err != nil {
		t.Fatal(err)
	}
	cached, err := version.LoadCache()
	if err != nil || *cached != *entry {
		t.Fatalf("fixture version cache round trip failed: %v", err)
	}
	for _, name := range []string{"config.json", "state.json", "version_cache.json"} {
		info, err := os.Stat(filepath.Join(expected, name))
		if err != nil || !info.Mode().IsRegular() {
			t.Fatalf("fixture %s not persisted: %v", name, err)
		}
	}
}

func TestSyntheticToolFailuresNeverStartProcesses(t *testing.T) {
	for _, cmd := range []*exec.Cmd{tmuxcmd.Command("display-message", "-p", "fixture"), hostexec.Command("td", "version", "--short")} {
		if cmd.Err == nil {
			t.Fatal("fixture command has no pre-start error")
		}
		if err := cmd.Run(); err == nil {
			t.Fatal("fixture command unexpectedly ran")
		}
		if cmd.Process != nil {
			t.Fatal("fixture command started a process")
		}
	}
}
