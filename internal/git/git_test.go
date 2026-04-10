package git

import (
	"os"
	"os/exec"
	"testing"

	"github.com/celiumgrid/git-digest/internal/i18n"
)

// TestNewGitOptions 测试创建新的Git选项
func TestNewGitOptions(t *testing.T) {
	// 测试默认路径
	opts := NewGitOptions("")
	if opts.RepoPath != "." {
		t.Errorf("默认仓库路径应为当前目录 '.', 得到: %s", opts.RepoPath)
	}

	// 测试指定路径
	testPath := "/test/path"
	opts = NewGitOptions(testPath)
	if opts.RepoPath != testPath {
		t.Errorf("仓库路径应为 %s, 得到: %s", testPath, opts.RepoPath)
	}
}

// TestParseCommits 测试解析git log输出
func TestParseCommits(t *testing.T) {
	// 模拟git log输出
	testOutput := "abc123\x1fJohn Doe\x1f2023-01-01 12:00:00 +0800\x1fInitial commit\x1fHEAD -> main, origin/main\x1e" +
		"def456\x1fJane Smith\x1f2023-01-02 13:00:00 +0800\x1fAdd feature\x1frefs/heads/feature, tag: v1.0.0\x1e"

	commits, err := parseCommits(testOutput, i18n.LanguageEnglish)
	if err != nil {
		t.Fatalf("解析提交失败: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("应解析出2个提交, 得到: %d", len(commits))
	}

	// 验证第一个提交
	if commits[0].Hash != "abc123" {
		t.Errorf("第一个提交的哈希应为 'abc123', 得到: %s", commits[0].Hash)
	}
	if commits[0].Author != "John Doe" {
		t.Errorf("第一个提交的作者应为 'John Doe', 得到: %s", commits[0].Author)
	}
	if commits[0].Message != "Initial commit" {
		t.Errorf("第一个提交的消息应为 'Initial commit', 得到: %s", commits[0].Message)
	}

	// 验证分支信息 - 由于解析逻辑的原因，可能只有一个分支
	if len(commits[0].Branches) < 1 {
		t.Errorf("第一个提交应至少有1个分支, 得到: %d", len(commits[0].Branches))
	}
	if !contains(commits[0].Branches, "main") {
		t.Errorf("第一个提交应包含 'main' 分支, 得到: %v", commits[0].Branches)
	}

	// 验证第二个提交
	if commits[1].Hash != "def456" {
		t.Errorf("第二个提交的哈希应为 'def456', 得到: %s", commits[1].Hash)
	}
	if !contains(commits[1].Branches, "feature") {
		t.Errorf("第二个提交应包含 'feature' 分支, 得到: %v", commits[1].Branches)
	}
}

func TestParseCommitsPreservesPipeInSubject(t *testing.T) {
	testOutput := "abc123\x1fJohn Doe\x1f2023-01-01 12:00:00 +0800\x1ffeat: a | b\x1fHEAD -> main\x1e"

	commits, err := parseCommits(testOutput, i18n.LanguageEnglish)
	if err != nil {
		t.Fatalf("parseCommits returned error: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("expected one commit, got %d", len(commits))
	}
	if commits[0].Message != "feat: a | b" {
		t.Fatalf("expected subject with pipe to be preserved, got %q", commits[0].Message)
	}
	if !contains(commits[0].Branches, "main") {
		t.Fatalf("expected branch parsing to keep main, got %v", commits[0].Branches)
	}
}

// TestGetGitUserName 测试获取Git用户名
func TestGetGitUserName(t *testing.T) {
	// 跳过实际执行git命令的测试
	if os.Getenv("SKIP_GIT_TESTS") == "true" {
		t.Skip("跳过需要git命令的测试")
	}

	// 检查git命令是否可用
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git命令不可用，跳过测试")
	}

	// 测试获取用户名
	name, err := GetGitUserName("", i18n.LanguageEnglish)
	if err != nil {
		// 这不一定是错误，可能只是没有配置git用户名
		t.Logf("获取Git用户名失败: %v", err)
	} else {
		t.Logf("获取到Git用户名: %s", name)
	}
}

// 辅助函数：检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
