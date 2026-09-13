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
	for _, zone := range []*time.Location{time.FixedZone("fixture UTC", 0), time.FixedZone("fixture offset", -7*60*60)} {
		t.Run(zone.String(), func(t *testing.T) {
			entry := &version.CacheEntry{LatestVersion: "v2.0.0", CurrentVersion: "v1.0.0", CheckedAt: time.Date(2026, time.September, 12, 20, 0, 0, 123456789, zone), HasUpdate: true}
			if err := version.SaveCache(entry); err != nil {
				t.Fatal(err)
			}
			cached, err := version.LoadCache()
			if err != nil || cached == nil {
				t.Fatalf("fixture version cache round trip failed: %v", err)
			}
			// JSON preserves the instant and offset, not time.Location identity.
			if cached.LatestVersion != entry.LatestVersion || cached.CurrentVersion != entry.CurrentVersion || cached.HasUpdate != entry.HasUpdate || !cached.CheckedAt.Equal(entry.CheckedAt) {
				t.Fatalf("fixture version cache round trip changed values: got %+v, want %+v", cached, entry)
			}
		})
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
