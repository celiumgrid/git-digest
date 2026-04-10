# Git Digest

Git Digest is a CLI for generating work reports from Git commit history.

Language docs:
- Chinese: [README.zh-CN.md](./README.zh-CN.md)

## Features

- Pulls commit history from one repository or many repositories
- Generates reports through an OpenAI-compatible chat completion layer
- Supports OpenAI, Gemini, and DeepSeek providers
- Supports preset periods, a single day, or a custom date range
- Supports text and Markdown output
- Supports custom prompt templates
- Supports interactive wizard mode
- Supports English and Chinese UI

## Installation

Go is the only supported distribution path. The built-in prompt templates are embedded in the binary, so `go install` is self-contained.

```bash
go install github.com/celiumgrid/git-digest/cmd/git-digest@latest
```

## Quick Start

```bash
export OPENAI_API_KEY="your-api-key"
git-digest
```

Or launch the wizard explicitly:

```bash
git-digest wizard
```

## Common Examples

```bash
# Language
git-digest --language en
git-digest --language zh

# Time input
git-digest --period last-7d
git-digest --period last-week
git-digest --period last-month
git-digest --period last-year
git-digest --on 2025-05-25
git-digest --from 2025-05-19 --to 2025-05-26

# Repository scope
git-digest --repo /path/to/your/repo
git-digest --repos /path/to/projects

# Provider configuration
git-digest --provider openai
git-digest --provider gemini
git-digest --provider deepseek
git-digest --provider openai --model gpt-4.1-mini
git-digest --provider openai --base-url https://your-proxy.example/v1

# Prompt templates
git-digest --prompt basic
git-digest --prompt detailed
git-digest --prompt targeted
git-digest --prompt /path/to/custom.txt

# Create the global base config
git-digest config init
```

## Time Input Model

Git Digest accepts three time input modes:

- `--period <preset>`
- `--on <YYYY-MM-DD>`
- `--from <YYYY-MM-DD> --to <YYYY-MM-DD>`

Supported `--period` values:

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

Rules:

- `--period` cannot be combined with `--on`
- `--period` cannot be combined with `--from/--to`
- `--on` cannot be combined with `--from/--to`
- `--from` and `--to` must be used together
- If no time input is provided, `--period last-7d` is used

## Language

Language is selected with `--language en|zh`.

- Default language: `en`
- The interactive wizard asks for language first
- All CLI help, prompts, runtime messages, and report text follow the selected language

## Configuration Priority

Configuration is applied in this order:

1. Built-in defaults
2. Global base config (`~/.config/git-digest/config.json`)
3. CLI flags
4. Interactive wizard input

Use `git-digest config init` to create or overwrite the global base config file.
This command writes a dedicated global template instead of saving the defaults from a single report run.
Fields left blank in the config wizard stay unset.

After saving it once, normal runs load it automatically:

```bash
git-digest
git-digest wizard
```

To override one value for the current run, pass a CLI flag:

```bash
git-digest --period last-month
```

To skip the global base config for one run:

```bash
git-digest --no-base-config
```

## Prompt Templates

Built-in prompt types:

- `basic`
- `detailed`
- `targeted`

These built-in templates are embedded in the binary.

You can also pass a custom prompt file path:

```bash
git-digest --prompt /path/to/custom.txt
```

Custom prompt files should include `{{.CommitMessages}}` as the placeholder for commit content.

## AI Provider Configuration

Unified provider fields:

- `provider`
- `base_url`
- `api_key`
- `model`

Default provider templates:

- OpenAI: `https://api.openai.com/v1`
- Gemini: `https://generativelanguage.googleapis.com/v1beta/openai`
- DeepSeek: `https://api.deepseek.com/v1`

Environment variable fallback:

- OpenAI: `OPENAI_API_KEY`
- Gemini: `GEMINI_API_KEY`, then `GOOGLE_API_KEY`
- DeepSeek: `DEEPSEEK_API_KEY`, then `OPENAI_API_KEY`

## Output

Supported formats:

- `text`
- `markdown`

## Release Notes

This project is distributed through Go installation and GitHub Releases.
Homebrew is not supported.

## License

MIT
