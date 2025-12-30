package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/yourname/docuguard/pkg/types"
)

var (
	// 匹配 diff 文件头: diff --git a/path b/path
	diffHeaderRegex = regexp.MustCompile(`^diff --git a/(.+) b/(.+)$`)
	// 匹配 hunk 头: @@ -old_start,old_count +new_start,new_count @@
	hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	// 匹配新文件
	newFileRegex = regexp.MustCompile(`^new file mode`)
	// 匹配删除文件
	deletedFileRegex = regexp.MustCompile(`^deleted file mode`)
)

// GetDiff 获取指定基准分支到当前 HEAD 的 diff
func GetDiff(baseBranch string) (string, error) {
	// 使用 git diff 获取变更
	cmd := exec.Command("git", "diff", baseBranch+"...HEAD")
	output, err := cmd.Output()
	if err != nil {
		// 尝试不带 ... 的方式
		cmd = exec.Command("git", "diff", baseBranch)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get git diff: %w", err)
		}
	}
	return string(output), nil
}

// GetDiffUncommitted 获取未提交的变更（工作区 + 暂存区）
func GetDiffUncommitted() (string, error) {
	// 获取暂存区变更
	cmd := exec.Command("git", "diff", "--cached")
	staged, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	// 获取工作区变更
	cmd = exec.Command("git", "diff")
	unstaged, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get unstaged diff: %w", err)
	}

	return string(staged) + string(unstaged), nil
}

// ParseDiff 解析 diff 内容，返回文件级别的变更信息
func ParseDiff(diffContent string) ([]types.FileDiff, error) {
	var fileDiffs []types.FileDiff
	var currentDiff *types.FileDiff

	scanner := bufio.NewScanner(strings.NewReader(diffContent))
	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否是新的文件 diff
		if matches := diffHeaderRegex.FindStringSubmatch(line); matches != nil {
			// 保存之前的 diff
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

		// 检查是否是新文件
		if newFileRegex.MatchString(line) {
			currentDiff.ChangeType = types.ChangeAdded
			continue
		}

		// 检查是否是删除文件
		if deletedFileRegex.MatchString(line) {
			currentDiff.ChangeType = types.ChangeDeleted
			continue
		}

		// 解析 hunk 头
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

	// 保存最后一个 diff
	if currentDiff != nil {
		fileDiffs = append(fileDiffs, *currentDiff)
	}

	return fileDiffs, scanner.Err()
}

// FilterGoFiles 过滤出 Go 文件的变更
func FilterGoFiles(diffs []types.FileDiff) []types.FileDiff {
	var goFiles []types.FileDiff
	for _, d := range diffs {
		if strings.HasSuffix(d.NewPath, ".go") || strings.HasSuffix(d.OldPath, ".go") {
			// 排除测试文件（可选）
			if !strings.HasSuffix(d.NewPath, "_test.go") {
				goFiles = append(goFiles, d)
			}
		}
	}
	return goFiles
}
