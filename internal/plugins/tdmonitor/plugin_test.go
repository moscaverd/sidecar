package tdmonitor

import (
	"database/sql"
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcus/sidecar/internal/plugin"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil plugin")
	}
}

func TestPluginID(t *testing.T) {
	p := New()
	if id := p.ID(); id != "td-monitor" {
		t.Errorf("expected ID 'td-monitor', got %q", id)
	}
}

func TestPluginName(t *testing.T) {
	p := New()
	if name := p.Name(); name != "td" {
		t.Errorf("expected Name 'td', got %q", name)
	}
}

func TestPluginIcon(t *testing.T) {
	p := New()
	if icon := p.Icon(); icon != "T" {
		t.Errorf("expected Icon 'T', got %q", icon)
	}
}

func TestFocusContext(t *testing.T) {
	p := New()

	// Without model, should return default
	if ctx := p.FocusContext(); ctx != "td-monitor" {
		t.Errorf("expected context 'td-monitor', got %q", ctx)
	}
}

func TestDiagnosticsNoDatabase(t *testing.T) {
	p := New()
	diags := p.Diagnostics()

	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	if diags[0].Status != "disabled" {
		t.Errorf("expected status 'disabled', got %q", diags[0].Status)
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{1, "1 issue"},
		{5, "5 issues"},
		{10, "10 issues"},
		{100, "100 issues"},
	}

	for _, tt := range tests {
		result := formatCount(tt.count, "issue", "issues")
		if result != tt.expected {
			t.Errorf("formatCount(%d) = %q, expected %q",
				tt.count, result, tt.expected)
		}
	}
}

func TestInitWithNonExistentDatabase(t *testing.T) {
	root := t.TempDir()
	// Existing local directory stops td from resolving global project associations.
	if err := os.Mkdir(filepath.Join(root, ".todos"), 0700); err != nil {
		t.Fatal(err)
	}
	p := New()
	ctx := &plugin.Context{
		WorkDir: root,
		Logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	// Init should NOT return an error even if database doesn't exist
	// This is silent degradation - plugin loads but shows "no database"
	err := p.Init(ctx)
	if err != nil {
		t.Errorf("Init should not return error for missing database, got: %v", err)
	}

	// Plugin should still be usable but model should be nil
	if p.ctx == nil {
		t.Error("context should be set")
	}
	if p.model != nil {
		t.Error("model should be nil when database not found")
	}
}

func TestInitWithValidDatabase(t *testing.T) {
	projectRoot := monitorFixture(t)

	p := New()
	ctx := &plugin.Context{
		WorkDir: projectRoot,
		Logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	err := p.Init(ctx)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}

	// Check if model was created
	if p.model == nil {
		t.Error("model should be created when database exists")
	}

	// Cleanup
	p.Stop()
}

func TestDiagnosticsWithDatabase(t *testing.T) {
	projectRoot := monitorFixture(t)

	p := New()
	ctx := &plugin.Context{
		WorkDir: projectRoot,
		Logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	if err := p.Init(ctx); err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	diags := p.Diagnostics()
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}

	// With database, status should be "ok"
	if diags[0].Status != "ok" {
		t.Errorf("expected status 'ok' with database, got %q", diags[0].Status)
	}
}

func TestNotInstalledModel(t *testing.T) {
	m := NewNotInstalledModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}

	// Test View renders content
	result := m.View(80, 24)
	if result == "" {
		t.Error("expected non-empty view")
	}

	// Check it contains expected content
	if !strings.Contains(result, "External memory") {
		t.Error("expected view to contain pitch text")
	}
}

func TestCommands(t *testing.T) {
	p := New()

	// Without model, should return nil
	cmds := p.Commands()
	if cmds != nil {
		t.Errorf("expected nil commands without model, got %d", len(cmds))
	}
}

func TestStartWithoutModel(t *testing.T) {
	p := New()

	// Start without model should return nil
	cmd := p.Start()
	if cmd != nil {
		t.Error("expected nil command without model")
	}
}

func TestViewWithoutModel(t *testing.T) {
	p := New()

	// View without model should show "no database" message
	view := p.View(80, 24)
	if view == "" {
		t.Error("expected non-empty view")
	}
}

// The pinned base schema intentionally runs td's real migrations in NewEmbedded.
//
//go:embed testdata/td-v0.37-base-schema.sql
var monitorSchema string

func monitorFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// td migrates process-cwd session files and reads the current Git branch.
	t.Chdir(root)
	t.Setenv("TD_SESSION_ID", "sidecar-monitor-fixture")
	if err := os.Mkdir(filepath.Join(root, ".todos"), 0700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(root, ".todos", "issues.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(monitorSchema); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return root
}
