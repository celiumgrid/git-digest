# Version And Base Config Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `git-digest version` show an install-time module version when ldflags are absent, and rename runtime/user-facing config semantics from "default configuration" to "base configuration" while keeping global config auto-loaded as wizard/CLI defaults.

**Architecture:** Keep the existing config precedence model, but rename the user-facing concept to "base configuration". Add a small runtime version resolver that prefers GoReleaser ldflags and falls back to `runtime/debug.ReadBuildInfo()` for `go install module@version` installs.

**Tech Stack:** Go, Cobra, runtime/debug, existing config merge and i18n layers

---
