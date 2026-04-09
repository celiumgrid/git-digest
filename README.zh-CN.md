# Git Digest

Git Digest 是一个根据 Git 提交记录生成工作报告的命令行工具。

语言文档：
- 英文版：[README.md](./README.md)

## 功能

- 支持单仓库和多仓库分析
- 通过 OpenAI-compatible Chat Completions 统一接入 AI 平台
- 支持 OpenAI、Gemini、DeepSeek
- 支持预设周期、指定日期和自定义区间
- 支持文本和 Markdown 输出
- 支持自定义提示词模板
- 支持交互式向导
- 支持英文和中文界面

## 安装

当前仅支持 Go 安装，不提供 Homebrew 分发。

```bash
go install github.com/celiumgrid/git-digest/cmd/git-digest@latest
```

## 快速开始

```bash
export OPENAI_API_KEY="your-api-key"
git-digest
```

也可以显式启动向导：

```bash
git-digest wizard
```

## 常用示例

```bash
# 语言
git-digest --language en
git-digest --language zh

# 时间输入
git-digest --period last-7d
git-digest --period last-week
git-digest --period last-month
git-digest --period last-year
git-digest --on 2025-05-25
git-digest --from 2025-05-19 --to 2025-05-26

# 仓库范围
git-digest --repo /path/to/your/repo
git-digest --repos /path/to/projects

# Provider 配置
git-digest --provider openai
git-digest --provider gemini
git-digest --provider deepseek
git-digest --provider openai --model gpt-4.1-mini
git-digest --provider openai --base-url https://your-proxy.example/v1

# 提示词模板
git-digest --prompt basic
git-digest --prompt detailed
git-digest --prompt targeted
git-digest --prompt /path/to/custom.txt

# 保存为全局基础配置
git-digest --repos /path/to/projects --period last-7d --format markdown --save-base-config
```

## 时间输入模型

支持三种时间输入方式：

- `--period <preset>`
- `--on <YYYY-MM-DD>`
- `--from <YYYY-MM-DD> --to <YYYY-MM-DD>`

支持的 `--period` 值：

- `today`
- `yesterday`
- `last-7d`
- `last-30d`
- `this-week`
- `last-week`
- `this-month`
- `last-month`
- `this-year`
- `last-year`

规则：

- `--period` 不能和 `--on` 同时使用
- `--period` 不能和 `--from/--to` 同时使用
- `--on` 不能和 `--from/--to` 同时使用
- `--from` 和 `--to` 必须成对出现
- 不传时间参数时，默认使用 `--period last-7d`

## 语言

通过 `--language en|zh` 选择语言。

- 默认语言：`en`
- 交互式向导会先询问语言
- CLI 帮助、交互提示、运行日志和报告内容都会跟随语言切换

## 配置优先级

配置按以下顺序生效：

1. 内置默认值
2. 全局基础配置（`~/.config/git-digest/config.json`）
3. 命令行参数
4. 交互式向导输入

如果使用 `--save-base-config`，当前配置会写入全局基础配置文件。

## 提示词模板

内置提示词类型：

- `basic`
- `detailed`
- `targeted`

也可以传入自定义提示词文件路径：

```bash
git-digest --prompt /path/to/custom.txt
```

自定义提示词文件中应包含 `{{.CommitMessages}}` 作为提交内容占位符。

## AI Provider 配置

统一配置字段：

- `provider`
- `base_url`
- `api_key`
- `model`

默认 Provider 模板：

- OpenAI: `https://api.openai.com/v1`
- Gemini: `https://generativelanguage.googleapis.com/v1beta/openai`
- DeepSeek: `https://api.deepseek.com/v1`

环境变量回退：

- OpenAI: `OPENAI_API_KEY`
- Gemini: `GEMINI_API_KEY`，其次 `GOOGLE_API_KEY`
- DeepSeek: `DEEPSEEK_API_KEY`，其次 `OPENAI_API_KEY`

## 输出

支持的输出格式：

- `text`
- `markdown`

## 发布说明

本项目通过 Go 安装和 GitHub Releases 分发。
不支持 Homebrew。

## 许可证

MIT
