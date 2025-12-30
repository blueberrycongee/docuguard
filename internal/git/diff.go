package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/blueberrycongee/docuguard/pkg/types"
)

var (
	// diffHeaderRegex matches diff file headers: diff --git a/path b/path
	diffHeaderRegex = regexp.MustCompile(`^diff --git a/(.+) b/(.+)$`)
	// hunkHeaderRegex matches hunk headers: @@ -old_start,old_count +new_start,new_count @@
	hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	// newFileRegex matches new file mode lines.
	newFileRegex = regexp.MustCompile(`^new file mode`)
	// deletedFileRegex matches deleted file mode lines.
	deletedFileRegex = regexp.MustCompile(`^deleted file mode`)
)

// GetDiff returns the diff between the base branch and HEAD.
// It first tries the three-dot notation (baseBranch...HEAD), then falls back
// to two-dot notation if that fails.
func GetDiff(baseBranch string) (string, error) {
	cmd := exec.Command("git", "diff", baseBranch+"...HEAD")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "diff", baseBranch)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get git diff: %w", err)
		}
	}
	return string(output), nil
}

// GetDiffUncommitted returns the diff of uncommitted changes.
// This includes both staged and unstaged changes.
func GetDiffUncommitted() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	staged, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	cmd = exec.Command("git", "diff")
	unstaged, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get unstaged diff: %w", err)
	}

	return string(staged) + string(unstaged), nil
}

// ParseDiff parses diff content and returns file-level change information.
func ParseDiff(diffContent string) ([]types.FileDiff, error) {
	var fileDiffs []types.FileDiff
	var currentDiff *types.FileDiff

	scanner := bufio.NewScanner(strings.NewReader(diffContent))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := diffHeaderRegex.FindStringSubmatch(line); matches != nil {
			if currentDiff != nil {
				fileDiffs = append(fileDiffs, *currentDiff)
			}
			currentDiff = &types.FileDiff{
				OldPath:    matches[1],
				NewPath:    matches[2],
				ChangeType: types.ChangeModified,
			}
			continue
		}

		if currentDiff == nil {
			continue
		}

		if newFileRegex.MatchString(line) {
			currentDiff.ChangeType = types.ChangeAdded
			continue
		}

		if deletedFileRegex.MatchString(line) {
			currentDiff.ChangeType = types.ChangeDeleted
			continue
		}

		if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
			lc := types.LineChange{}

			lc.OldStart, _ = strconv.Atoi(matches[1])
			if matches[2] != "" {
				lc.OldCount, _ = strconv.Atoi(matches[2])
			} else {
				lc.OldCount = 1
			}

			lc.NewStart, _ = strconv.Atoi(matches[3])
			if matches[4] != "" {
				lc.NewCount, _ = strconv.Atoi(matches[4])
			} else {
				lc.NewCount = 1
			}

			currentDiff.ChangedLines = append(currentDiff.ChangedLines, lc)
		}
	}

	if currentDiff != nil {
		fileDiffs = append(fileDiffs, *currentDiff)
	}

	return fileDiffs, scanner.Err()
}

// FilterGoFiles filters the diff list to include only Go source files.
// Test files (*_test.go) are excluded.
func FilterGoFiles(diffs []types.FileDiff) []types.FileDiff {
	var goFiles []types.FileDiff
	for _, d := range diffs {
		if strings.HasSuffix(d.NewPath, ".go") || strings.HasSuffix(d.OldPath, ".go") {
			if !strings.HasSuffix(d.NewPath, "_test.go") {
				goFiles = append(goFiles, d)
			}
		}
	}
	return goFiles
}

// IsInGitRepo checks if the current directory is inside a git repository.
func IsInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetCurrentBranch returns the name of the current branch.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
