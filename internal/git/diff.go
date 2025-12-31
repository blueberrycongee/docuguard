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

	// Go symbol patterns for extracting from diff lines
	funcDeclRegex   = regexp.MustCompile(`func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`)
	constVarRegex   = regexp.MustCompile(`^\s*(\w+)\s*=`)
	typeRegex       = regexp.MustCompile(`type\s+(\w+)\s+`)
	constBlockRegex = regexp.MustCompile(`const\s*\(`)
	varBlockRegex   = regexp.MustCompile(`var\s*\(`)
)

// GetDiff returns the diff between the base branch and HEAD.
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

// ParseDiffWithContent parses diff and extracts changed lines content.
func ParseDiffWithContent(diffContent string) ([]types.FileDiff, error) {
	var fileDiffs []types.FileDiff
	var currentDiff *types.FileDiff
	var addedLines, removedLines []string

	scanner := bufio.NewScanner(strings.NewReader(diffContent))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := diffHeaderRegex.FindStringSubmatch(line); matches != nil {
			if currentDiff != nil {
				currentDiff.AddedLines = addedLines
				currentDiff.RemovedLines = removedLines
				fileDiffs = append(fileDiffs, *currentDiff)
			}
			currentDiff = &types.FileDiff{
				OldPath:    matches[1],
				NewPath:    matches[2],
				ChangeType: types.ChangeModified,
			}
			addedLines = nil
			removedLines = nil
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

		// Capture added and removed lines
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, strings.TrimPrefix(line, "+"))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removedLines = append(removedLines, strings.TrimPrefix(line, "-"))
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
		currentDiff.AddedLines = addedLines
		currentDiff.RemovedLines = removedLines
		fileDiffs = append(fileDiffs, *currentDiff)
	}

	return fileDiffs, scanner.Err()
}

// ExtractSymbolsFromDiffLines extracts Go symbol names from diff lines.
func ExtractSymbolsFromDiffLines(lines []string) []string {
	seen := make(map[string]bool)
	var symbols []string

	for _, line := range lines {
		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Extract function names
		if matches := funcDeclRegex.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if !seen[name] {
				seen[name] = true
				symbols = append(symbols, name)
			}
		}

		// Extract type names
		if matches := typeRegex.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if !seen[name] {
				seen[name] = true
				symbols = append(symbols, name)
			}
		}

		// Extract const/var names (simple assignment)
		if matches := constVarRegex.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			// Filter out Go keywords and common patterns
			if !isGoKeyword(name) && !seen[name] {
				seen[name] = true
				symbols = append(symbols, name)
			}
		}
	}

	return symbols
}

func isGoKeyword(s string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true,
		"continue": true, "default": true, "defer": true, "else": true,
		"fallthrough": true, "for": true, "func": true, "go": true,
		"goto": true, "if": true, "import": true, "interface": true,
		"map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true,
		"var": true, "true": true, "false": true, "nil": true,
	}
	return keywords[s]
}

// FilterGoFiles filters the diff list to include only Go source files.
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
