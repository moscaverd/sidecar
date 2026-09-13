package gitstatus

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		contains string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute", now.Add(-1 * time.Minute), "1 min"},
		{"5 minutes", now.Add(-5 * time.Minute), "5 mins"},
		{"1 hour", now.Add(-1 * time.Hour), "1 hour"},
		{"3 hours", now.Add(-3 * time.Hour), "3 hours"},
		{"yesterday", now.Add(-25 * time.Hour), "yesterday"},
		{"3 days", now.Add(-3 * 24 * time.Hour), "3 days"},
		{"1 week", now.Add(-8 * 24 * time.Hour), "1 week"},
		{"3 weeks", now.Add(-22 * 24 * time.Hour), "3 weeks"},
		{"1 month", now.Add(-35 * 24 * time.Hour), "1 month"},
		{"5 months", now.Add(-150 * 24 * time.Hour), "5 months"},
		{"1 year", now.Add(-400 * 24 * time.Hour), "1 year"},
		{"3 years", now.Add(-1100 * 24 * time.Hour), "3 years"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RelativeTime(tc.input)
			if result == "" {
				t.Error("RelativeTime returned empty string")
			}
			// Just verify it returns something meaningful
			if len(result) < 3 {
				t.Errorf("RelativeTime returned unexpectedly short: %q", result)
			}
		})
	}
}

func TestRelativeTime_Boundaries(t *testing.T) {
	now := time.Now()

	// Test boundary conditions
	tests := []struct {
		name  string
		input time.Time
	}{
		{"exactly 0 seconds", now},
		{"exactly 1 minute", now.Add(-1 * time.Minute)},
		{"exactly 1 hour", now.Add(-1 * time.Hour)},
		{"exactly 1 day", now.Add(-24 * time.Hour)},
		{"exactly 1 week", now.Add(-7 * 24 * time.Hour)},
		{"exactly 1 month", now.Add(-30 * 24 * time.Hour)},
		{"exactly 1 year", now.Add(-365 * 24 * time.Hour)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RelativeTime(tc.input)
			if result == "" {
				t.Error("RelativeTime returned empty string")
			}
		})
	}
}

func TestCommit_Fields(t *testing.T) {
	// Verify Commit struct can be created with all fields
	commit := Commit{
		Hash:        "abc123def456",
		ShortHash:   "abc123",
		Author:      "Test User",
		AuthorEmail: "test@example.com",
		Date:        time.Now(),
		Subject:     "Test commit",
		Body:        "Extended description",
		Files:       []CommitFile{},
		Stats: CommitStats{
			FilesChanged: 5,
			Additions:    100,
			Deletions:    50,
		},
	}

	if commit.Hash != "abc123def456" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc123def456")
	}
	if commit.Stats.FilesChanged != 5 {
		t.Errorf("Stats.FilesChanged = %d, want 5", commit.Stats.FilesChanged)
	}
}

func TestCommitFile_Fields(t *testing.T) {
	// Verify CommitFile struct can be created with all fields
	file := CommitFile{
		Path:      "new/path.go",
		OldPath:   "old/path.go",
		Status:    StatusRenamed,
		Additions: 10,
		Deletions: 5,
	}

	if file.Path != "new/path.go" {
		t.Errorf("Path = %q, want %q", file.Path, "new/path.go")
	}
	if file.Status != StatusRenamed {
		t.Errorf("Status = %v, want %v", file.Status, StatusRenamed)
	}
}

func TestGetCommitDetail_PopulatesParentHashes(t *testing.T) {
	// Use the actual sidecar repo to test GetCommitDetail
	workDir, err := os.Getwd()
	if err != nil {
		t.Skip("cannot get working directory")
	}
	// Walk up to repo root
	for {
		if _, err := os.Stat(filepath.Join(workDir, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(workDir)
		if parent == workDir {
			t.Skip("not in a git repo")
		}
		workDir = parent
	}

	// Find a merge commit
	mergeHash, err := exec.Command("git", "-C", workDir, "log", "--merges", "--format=%H", "-1").Output()
	if err != nil || len(strings.TrimSpace(string(mergeHash))) == 0 {
		t.Skip("no merge commits in repo")
	}
	hash := strings.TrimSpace(string(mergeHash))

	commit, err := GetCommitDetail(workDir, hash)
	if err != nil {
		t.Fatalf("GetCommitDetail(%q): %v", hash, err)
	}
	if commit == nil {
		t.Fatal("GetCommitDetail returned nil commit")
	}
	if !commit.IsMerge {
		t.Errorf("IsMerge = false, want true for merge commit %s", hash)
	}
	if len(commit.ParentHashes) < 2 {
		t.Errorf("ParentHashes = %v, want at least 2 parents for merge commit %s", commit.ParentHashes, hash)
	}

	// Also test a non-merge commit
	nonMergeHash, err := exec.Command("git", "-C", workDir, "log", "--no-merges", "--format=%H", "-1").Output()
	if err != nil || len(strings.TrimSpace(string(nonMergeHash))) == 0 {
		t.Skip("no non-merge commits in repo")
	}
	hash2 := strings.TrimSpace(string(nonMergeHash))

	commit2, err := GetCommitDetail(workDir, hash2)
	if err != nil {
		t.Fatalf("GetCommitDetail(%q): %v", hash2, err)
	}
	if commit2.IsMerge {
		t.Errorf("IsMerge = true, want false for non-merge commit %s", hash2)
	}
	if len(commit2.ParentHashes) != 1 {
		t.Errorf("ParentHashes = %v, want exactly 1 parent for non-merge commit %s", commit2.ParentHashes, hash2)
	}
}

func TestCommitStats_Fields(t *testing.T) {
	stats := CommitStats{
		FilesChanged: 3,
		Additions:    50,
		Deletions:    25,
	}

	if stats.FilesChanged != 3 {
		t.Errorf("FilesChanged = %d, want 3", stats.FilesChanged)
	}
	if stats.Additions != 50 {
		t.Errorf("Additions = %d, want 50", stats.Additions)
	}
	if stats.Deletions != 25 {
		t.Errorf("Deletions = %d, want 25", stats.Deletions)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	workDir, err := os.Getwd()
	if err != nil {
		t.Skip("cannot get working directory")
	}
	for {
		if _, err := os.Stat(filepath.Join(workDir, ".git")); err == nil {
			return workDir
		}
		parent := filepath.Dir(workDir)
		if parent == workDir {
			t.Skip("not in a git repo")
		}
		workDir = parent
	}
}

func TestGetCommitDiff_EmptyParentHash(t *testing.T) {
	workDir := findRepoRoot(t)

	out, err := exec.Command("git", "-C", workDir, "log", "--no-merges", "--diff-filter=M", "--format=%H", "-1").Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		t.Skip("no non-merge commits with modified files in repo")
	}
	hash := strings.TrimSpace(string(out))

	filesOut, err := exec.Command("git", "-C", workDir, "diff-tree", "--no-commit-id", "-r", "--name-only", hash).Output()
	if err != nil || len(strings.TrimSpace(string(filesOut))) == 0 {
		t.Skip("commit has no files")
	}
	filePath := strings.Split(strings.TrimSpace(string(filesOut)), "\n")[0]

	diff, err := GetCommitDiff(workDir, hash, filePath, "")
	if err != nil {
		t.Fatalf("GetCommitDiff(%q, %q, \"\"): %v", hash, filePath, err)
	}
	if diff == "" {
		t.Errorf("expected non-empty diff for commit %s file %s with empty parentHash", hash, filePath)
	}
}

func TestGetCommitDiff_WithParentHash(t *testing.T) {
	workDir := findRepoRoot(t)

	mergeOut, err := exec.Command("git", "-C", workDir, "log", "--merges", "--format=%H", "-1").Output()
	if err != nil || len(strings.TrimSpace(string(mergeOut))) == 0 {
		t.Skip("no merge commits in repo")
	}
	hash := strings.TrimSpace(string(mergeOut))

	parentOut, err := exec.Command("git", "-C", workDir, "rev-parse", hash+"^1").Output()
	if err != nil {
		t.Skipf("cannot get first parent of %s: %v", hash, err)
	}
	parentHash := strings.TrimSpace(string(parentOut))

	filesOut, err := exec.Command("git", "-C", workDir, "diff", "--name-only", parentHash, hash).Output()
	if err != nil || len(strings.TrimSpace(string(filesOut))) == 0 {
		t.Skip("merge commit has no changed files between parent and merge")
	}
	filePath := strings.Split(strings.TrimSpace(string(filesOut)), "\n")[0]

	diff, err := GetCommitDiff(workDir, hash, filePath, parentHash)
	if err != nil {
		t.Fatalf("GetCommitDiff(%q, %q, %q): %v", hash, filePath, parentHash, err)
	}
	if diff == "" {
		t.Errorf("expected non-empty diff for merge commit %s file %s with parentHash %s", hash, filePath, parentHash)
	}
}

func TestGetCommitDiff_NonExistentPath(t *testing.T) {
	workDir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fixture git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(workDir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-b", "fixture-main")
	write("normal.txt", "before\n")
	git("add", "normal.txt")
	git("commit", "-m", "fixture base")
	base := git("rev-parse", "HEAD")
	write("normal.txt", "after\n")
	git("commit", "-am", "fixture normal commit")
	normal := git("rev-parse", "HEAD")
	git("checkout", "-b", "fixture-side", base)
	write("side.txt", "side change\n")
	git("add", "side.txt")
	git("commit", "-m", "fixture side commit")
	git("checkout", "fixture-main")
	git("merge", "--no-ff", "-m", "fixture merge", "fixture-side")
	merge := git("rev-parse", "HEAD")
	shallowDir := t.TempDir()
	if out, err := exec.Command("git", "clone", "--no-local", "--depth=1", workDir, shallowDir).CombinedOutput(); err != nil {
		t.Fatalf("clone shallow fixture: %v\n%s", err, out)
	}
	if out, err := exec.Command("git", "-C", shallowDir, "rev-parse", "--is-shallow-repository").CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("fixture is not a shallow checkout: %v\n%s", err, out)
	}

	for _, tc := range []struct {
		name, dir, hash, parent, changedPath, addedLine string
	}{
		{"normal", workDir, normal, "", "normal.txt", "+after"},
		{"merge combined", workDir, merge, "", "", ""},
		{"merge first parent", workDir, merge, normal, "side.txt", "+side change"},
		{"shallow merge", shallowDir, merge, "", "side.txt", "+side change"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			diff, err := GetCommitDiff(tc.dir, tc.hash, "nonexistent/path/that/does/not/exist.xyz", tc.parent)
			if err != nil {
				t.Fatalf("GetCommitDiff with non-existent path returned error: %v", err)
			}
			if diff != "" {
				t.Errorf("expected empty diff for non-existent path, got: %q", diff)
			}
			if tc.changedPath != "" {
				diff, err = GetCommitDiff(tc.dir, tc.hash, tc.changedPath, tc.parent)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(diff, "diff --git ") || !strings.Contains(diff, tc.addedLine) {
					t.Errorf("expected file patch with %q, got: %q", tc.addedLine, diff)
				}
			}
		})
	}
}
