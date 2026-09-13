package version

import (
	"net/http"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/marcus/sidecar/internal/hostexec"
	"github.com/marcus/sidecar/internal/testutil"
)

func TestUpdateCommand(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		method   InstallMethod
		contains []string
	}{
		{
			name:     "go install",
			version:  "v1.0.0",
			method:   InstallMethodGo,
			contains: []string{"go install", "v1.0.0", "github.com/marcus/sidecar"},
		},
		{
			name:     "go install with ldflags",
			version:  "v2.1.3",
			method:   InstallMethodGo,
			contains: []string{"-ldflags", "v2.1.3"},
		},
		{
			name:     "homebrew",
			version:  "v1.0.0",
			method:   InstallMethodHomebrew,
			contains: []string{"brew upgrade sidecar"},
		},
		{
			name:     "binary download",
			version:  "v1.0.0",
			method:   InstallMethodBinary,
			contains: []string{"https://github.com/marcus/sidecar/releases/tag/v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := updateCommand(tt.version, tt.method)
			for _, want := range tt.contains {
				if !strings.Contains(cmd, want) {
					t.Errorf("updateCommand(%q, %q) = %q, want to contain %q", tt.version, tt.method, cmd, want)
				}
			}
		})
	}
}

func TestCheck_DevelopmentVersion(t *testing.T) {
	// Development versions should return empty result without making HTTP calls
	devVersions := []string{"", "unknown", "devel", "devel+abc123"}

	for _, v := range devVersions {
		t.Run(v, func(t *testing.T) {
			result := Check(v)
			if result.HasUpdate {
				t.Errorf("Check(%q) should not have update for dev version", v)
			}
			if result.Error != nil {
				t.Errorf("Check(%q) should not error for dev version: %v", v, result.Error)
			}
		})
	}
}

func TestCheck_APIErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{
			name:       "404 not found",
			statusCode: http.StatusNotFound,
			body:       `{"message": "Not Found"}`,
			wantErr:    true,
		},
		{
			name:       "429 rate limited",
			statusCode: http.StatusTooManyRequests,
			body:       `{"message": "rate limit exceeded"}`,
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			body:       `{"message": "Internal Server Error"}`,
			wantErr:    true,
		},
		{
			name:       "200 success",
			statusCode: http.StatusOK,
			body:       `{"tag_name": "v1.0.0", "html_url": "https://github.com/marcus/sidecar/releases/tag/v1.0.0"}`,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := releaseResponse(t, tt.statusCode, tt.body)
			result := Check("v0.9.0")
			if *calls != 1 || (result.Error != nil) != tt.wantErr {
				t.Fatalf("calls=%d result=%+v", *calls, result)
			}
			if !tt.wantErr && (!result.HasUpdate || result.LatestVersion != "v1.0.0" || result.UpdateURL != "https://github.com/marcus/sidecar/releases/tag/v1.0.0") {
				t.Fatalf("release response lost: %+v", result)
			}
		})
	}
}

func TestCheck_InvalidJSON(t *testing.T) {
	calls := releaseResponse(t, http.StatusOK, `{invalid json`)
	result := Check("v0.9.0")
	if *calls != 1 || result.Error == nil {
		t.Fatalf("calls=%d result=%+v", *calls, result)
	}
}

func TestCheckAsync_CacheHit(t *testing.T) {
	testutil.Home(t, t.TempDir())
	calls := releaseResponse(t, http.StatusInternalServerError, "unexpected cache miss")
	for _, hasUpdate := range []bool{true, false} {
		if err := SaveCache(&CacheEntry{CurrentVersion: "v1.0.0", LatestVersion: "v1.1.0", HasUpdate: hasUpdate, CheckedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		msg := CheckAsync("v1.0.0")()
		if hasUpdate {
			update, ok := msg.(UpdateAvailableMsg)
			if !ok || update.CurrentVersion != "v1.0.0" || update.LatestVersion != "v1.1.0" {
				t.Fatalf("cache update lost: %+v", msg)
			}
		} else if msg != nil {
			t.Fatalf("up-to-date cache returned %+v", msg)
		}
	}
	if *calls != 0 {
		t.Fatalf("cache hit made %d requests", *calls)
	}
}

func TestUpdateAvailableMsg(t *testing.T) {
	// Verify UpdateAvailableMsg structure
	msg := UpdateAvailableMsg{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.1.0",
		UpdateCommand:  "go install ...",
	}

	if msg.CurrentVersion != "v1.0.0" {
		t.Error("CurrentVersion mismatch")
	}
	if msg.LatestVersion != "v1.1.0" {
		t.Error("LatestVersion mismatch")
	}
}

func TestCheckResult(t *testing.T) {
	// Verify CheckResult structure and fields
	result := CheckResult{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.2.0",
		UpdateURL:      "https://github.com/marcus/sidecar/releases/tag/v1.2.0",
		HasUpdate:      true,
		Error:          nil,
	}

	if !result.HasUpdate {
		t.Error("Expected HasUpdate to be true")
	}
	if result.Error != nil {
		t.Error("Expected no error")
	}
}

func TestRelease(t *testing.T) {
	// Verify Release struct for JSON unmarshaling
	r := Release{
		TagName:     "v1.0.0",
		PublishedAt: time.Now(),
		HTMLURL:     "https://github.com/marcus/sidecar/releases/tag/v1.0.0",
	}

	if r.TagName != "v1.0.0" {
		t.Error("TagName mismatch")
	}
}

func TestTdUpdateCommand(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		method   InstallMethod
		contains []string
	}{
		{
			name:     "go install",
			version:  "v0.4.12",
			method:   InstallMethodGo,
			contains: []string{"go install", "v0.4.12", "github.com/marcus/td"},
		},
		{
			name:     "go install v1",
			version:  "v1.0.0",
			method:   InstallMethodGo,
			contains: []string{"go install", "v1.0.0", "marcus/td"},
		},
		{
			name:     "homebrew",
			version:  "v0.4.12",
			method:   InstallMethodHomebrew,
			contains: []string{"brew upgrade td"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tdUpdateCommand(tt.version, tt.method)
			for _, want := range tt.contains {
				if !strings.Contains(cmd, want) {
					t.Errorf("tdUpdateCommand(%q, %q) = %q, want to contain %q", tt.version, tt.method, cmd, want)
				}
			}
		})
	}
}

func TestTdVersionMsg(t *testing.T) {
	// Test TdVersionMsg struct - installed with update
	msgWithUpdate := TdVersionMsg{
		Installed:      true,
		CurrentVersion: "v0.4.12",
		LatestVersion:  "v0.4.13",
		HasUpdate:      true,
	}

	if !msgWithUpdate.Installed {
		t.Error("Expected Installed to be true")
	}
	if !msgWithUpdate.HasUpdate {
		t.Error("Expected HasUpdate to be true")
	}
	if msgWithUpdate.CurrentVersion != "v0.4.12" {
		t.Errorf("CurrentVersion = %q, want v0.4.12", msgWithUpdate.CurrentVersion)
	}

	// Test TdVersionMsg - not installed
	msgNotInstalled := TdVersionMsg{
		Installed: false,
	}

	if msgNotInstalled.Installed {
		t.Error("Expected Installed to be false")
	}
	if msgNotInstalled.HasUpdate {
		t.Error("Expected HasUpdate to be false when not installed")
	}

	// Test TdVersionMsg - installed, up to date
	msgUpToDate := TdVersionMsg{
		Installed:      true,
		CurrentVersion: "v0.4.13",
		LatestVersion:  "v0.4.13",
		HasUpdate:      false,
	}

	if msgUpToDate.HasUpdate {
		t.Error("Expected HasUpdate to be false when up to date")
	}
}

func TestGetTdVersion(t *testing.T) {
	previous := hostexec.Command
	var got []string
	hostexec.Command = func(name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return &exec.Cmd{Args: got, Err: exec.ErrNotFound}
	}
	t.Cleanup(func() { hostexec.Command = previous })
	if version := GetTdVersion(); version != "" {
		t.Fatalf("failed command returned version %q", version)
	}
	if !reflect.DeepEqual(got, []string{"td", "version", "--short"}) {
		t.Fatalf("unexpected td lookup: %v", got)
	}
}

func TestCheckTdAsync(t *testing.T) {
	// CheckTdAsync returns a tea.Cmd that checks td version
	// Behavior depends on:
	// 1. Whether td is installed (GetTdVersion returns non-empty)
	// 2. Cache validity
	// 3. GitHub API response

	cmd := CheckTdAsync()
	if cmd == nil {
		t.Error("CheckTdAsync should return a non-nil command")
	}

	// The command is a closure that returns TdVersionMsg
	// We can't easily test the full flow without mocking, but we verify
	// the command is callable and returns a message type
	msg := cmd()

	// Should return TdVersionMsg (either installed or not)
	switch m := msg.(type) {
	case TdVersionMsg:
		// Expected - verify fields are reasonable
		if m.Installed && m.CurrentVersion == "" {
			t.Error("Installed td should have CurrentVersion")
		}
	case nil:
		// Also acceptable if td is not installed
	default:
		t.Errorf("CheckTdAsync returned unexpected type: %T", msg)
	}
}

func TestCheckTd(t *testing.T) {
	// CheckTd is the synchronous version check for td
	// Similar to Check but for the td repo

	// Development version should skip check
	result := CheckTd("")
	if result.HasUpdate {
		t.Error("Empty version should not have update")
	}

	result = CheckTd("devel")
	if result.HasUpdate {
		t.Error("devel version should not have update")
	}
}
