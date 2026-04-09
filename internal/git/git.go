package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/celiumgrid/git-digest/internal/i18n"
)

// Options Git操作的选项
type Options struct {
	RepoPath string // Git仓库路径
	Author   string // 作者名称，用于筛选提交
	Language string
}

// NewGitOptions 创建新的Git选项
func NewGitOptions(repoPath string) *Options {
	return NewGitOptionsWithLanguage(repoPath, i18n.LanguageEnglish)
}

// NewGitOptionsWithLanguage 创建带语言信息的Git选项
func NewGitOptionsWithLanguage(repoPath, language string) *Options {
	// 如果没有指定路径，使用当前目录
	if repoPath == "" {
		repoPath = "."
	}

	// 创建Git选项
	opts := &Options{
		RepoPath: repoPath,
		Language: i18n.NormalizeLanguage(language),
	}

	// 获取当前用户的Git用户名
	author, err := GetGitUserName(repoPath, opts.Language)
	if err == nil && author != "" {
		opts.Author = author
	}

	return opts
}

// CommitInfo 表示一个Git提交的信息
type CommitInfo struct {
	Hash         string
	Author       string
	Date         time.Time
	Message      string
	Branches     []string // 分支信息
	ChangedFiles []string
	RepoPath     string // 仓库路径，标识提交来自哪个仓库
}

// GetCommitsBetween 获取指定时间范围内的所有提交
func GetCommitsBetween(fromDate, toDate time.Time, opts *Options) ([]CommitInfo, error) {
	language := optionLanguage(opts)
	// 格式化日期为git log可接受的格式
	fromStr := fromDate.Format(time.RFC3339)
	toStr := toDate.Format(time.RFC3339)

	// 构建git log命令的参数列表
	args := []string{
		"log",
		"--all",                            // 获取所有分支的提交
		"--pretty=format:%H|%an|%ad|%s|%D", // 添加%D获取分支信息
		"--date=iso",
		"--since=" + fromStr,
		"--until=" + toStr,
	}

	// 如果指定了作者，添加作者筛选条件
	if opts != nil && opts.Author != "" {
		args = append(args, "--author="+opts.Author)
	}

	// 构建git log命令
	cmd := exec.Command("git", args...)

	// 设置工作目录
	if opts != nil {
		cmd.Dir = opts.RepoPath
	}

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T(language, "git.exec_log"), err)
	}

	// 解析输出
	return parseCommits(string(output), language)
}

// GetCommitsThisWeek 获取本周的所有提交
func GetCommitsThisWeek(opts *Options) ([]CommitInfo, error) {
	// 计算本周一和下周一的日期
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // 周日
		weekday = 7
	}

	// 计算本周一的日期
	monday := now.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	// 计算下周一的日期
	nextMonday := monday.AddDate(0, 0, 7)

	return GetCommitsBetween(monday, nextMonday, opts)
}

// GetCommitDetails 获取指定提交的详细信息
func GetCommitDetails(hash string, opts *Options) (*CommitInfo, error) {
	language := optionLanguage(opts)
	// 获取提交的基本信息
	cmd := exec.Command("git", "show",
		"--pretty=format:%H|%an|%ad|%s|%D",
		"--date=iso",
		hash)

	// 设置工作目录
	if opts != nil {
		cmd.Dir = opts.RepoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T(language, "git.commit_details"), err)
	}

	// 解析提交信息
	commits, err := parseCommits(string(output), language)
	if err != nil || len(commits) == 0 {
		return nil, fmt.Errorf(i18n.T(language, "git.parse_commit_details"), err)
	}

	commit := commits[0]

	// 获取变更的文件列表
	cmdFiles := exec.Command("git", "show", "--name-only", "--pretty=format:", hash)

	// 设置工作目录
	if opts != nil {
		cmdFiles.Dir = opts.RepoPath
	}

	outputFiles, err := cmdFiles.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T(language, "git.changed_files"), err)
	}

	// 解析变更文件列表
	files := strings.Split(strings.TrimSpace(string(outputFiles)), "\n")
	commit.ChangedFiles = files

	return &commit, nil
}

// GetGitUserName 获取Git用户名
func GetGitUserName(repoPath, language string) (string, error) {
	// 构建git config命令获取用户名
	cmd := exec.Command("git", "config", "user.name")

	// 设置工作目录
	if repoPath != "" {
		cmd.Dir = repoPath
	}

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		// 如果获取失败，尝试获取全局用户名
		cmdGlobal := exec.Command("git", "config", "--global", "user.name")
		output, err = cmdGlobal.Output()
		if err != nil {
			return "", fmt.Errorf(i18n.T(language, "git.git_user"), err)
		}
	}

	// 去除空白字符
	return strings.TrimSpace(string(output)), nil
}

// parseCommits 解析git log的输出
func parseCommits(output string, language string) ([]CommitInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5) // 增加了分支信息字段
		if len(parts) < 5 {
			continue
		}

		hash := parts[0]
		author := parts[1]
		dateStr := parts[2]
		message := parts[3]
		refNames := parts[4]

		// 解析日期
		date, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			return nil, fmt.Errorf(i18n.T(language, "git.parse_date"), err)
		}

		// 解析分支信息
		var branches []string
		if refNames != "" {
			// 分割引用名称（如HEAD -> main, origin/main）
			refs := strings.Split(refNames, ",")
			for _, ref := range refs {
				ref = strings.TrimSpace(ref)
				// 只保留分支名称，去除tag和HEAD指针
				switch {
				case strings.Contains(ref, "refs/heads/"):
					branch := strings.TrimPrefix(ref, "refs/heads/")
					branches = append(branches, branch)
				case strings.Contains(ref, "HEAD -> "):
					branch := strings.TrimPrefix(ref, "HEAD -> ")
					branches = append(branches, branch)
				case !strings.Contains(ref, "tag:") && !strings.HasPrefix(ref, "HEAD"):
					// 去除远程分支前缀
					if strings.Contains(ref, "/") {
						parts := strings.SplitN(ref, "/", 2)
						if len(parts) > 1 {
							branches = append(branches, parts[1])
						}
					} else {
						branches = append(branches, ref)
					}
				}
			}
		}

		// 去除重复的分支名
		uniqueBranches := make([]string, 0)
		branchMap := make(map[string]bool)
		for _, branch := range branches {
			if !branchMap[branch] {
				branchMap[branch] = true
				uniqueBranches = append(uniqueBranches, branch)
			}
		}

		commits = append(commits, CommitInfo{
			Hash:     hash,
			Author:   author,
			Date:     date,
			Message:  message,
			Branches: uniqueBranches,
		})
	}

	return commits, nil
}

func optionLanguage(opts *Options) string {
	if opts == nil {
		return i18n.LanguageEnglish
	}
	return i18n.NormalizeLanguage(opts.Language)
}

// DiscoverGitRepos 发现指定目录下的所有Git仓库
func DiscoverGitRepos(rootPath, language string) ([]string, error) {
	var repos []string
	language = i18n.NormalizeLanguage(language)

	// 如果根路径为空，使用当前目录
	if rootPath == "" {
		rootPath = "."
	}

	// 展开用户目录并转换为绝对路径
	absRootPath, err := NormalizePath(rootPath)
	if err != nil {
		return nil, fmt.Errorf(i18n.T(language, "git.abs_path"), err)
	}

	// 定义要跳过的目录（常见的非Git仓库目录）
	skipDirs := map[string]bool{
		"node_modules":     true,
		"vendor":           true,
		".vscode":          true,
		".idea":            true,
		"target":           true,
		"build":            true,
		"dist":             true,
		"out":              true,
		"bin":              true,
		"obj":              true,
		".next":            true,
		".nuxt":            true,
		"coverage":         true,
		".nyc_output":      true,
		".pytest_cache":    true,
		"__pycache__":      true,
		".gradle":          true,
		".mvn":             true,
		"bower_components": true,
		"jspm_packages":    true,
		".tmp":             true,
		"tmp":              true,
		"temp":             true,
		".cache":           true,
		"logs":             true,
		"*.log":            true,
	}

	fmt.Printf(i18n.T(language, "git.scan_dir")+"\n", absRootPath)

	// 遍历目录
	err = filepath.Walk(absRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 忽略权限错误等，继续遍历
			return nil
		}

		// 如果是目录
		if info.IsDir() {
			// 检查是否为.git目录
			if info.Name() == ".git" {
				repoPath := filepath.Dir(path)
				repos = append(repos, repoPath)
				fmt.Printf(i18n.T(language, "git.found_repo")+"\n", repoPath)
				// 跳过.git目录的子目录遍历
				return filepath.SkipDir
			}

			// 跳过常见的非Git仓库目录
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}

			// 跳过隐藏目录（除了已知的如.git）
			if strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
				// 允许一些常见的项目目录
				allowedHiddenDirs := map[string]bool{
					".github": true,
					".vscode": false, // 已在skipDirs中
					".idea":   false, // 已在skipDirs中
				}
				if !allowedHiddenDirs[info.Name()] {
					return filepath.SkipDir
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf(i18n.T(language, "git.walk"), err)
	}

	fmt.Printf(i18n.T(language, "git.scan_done")+"\n", len(repos))
	return repos, nil
}
