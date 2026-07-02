package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nextvibe/nextvibe/internal/brand"
)

const (
	sectionStart = "<!-- NEXTVIBE:START -->"
	sectionEnd   = "<!-- NEXTVIBE:END -->"
)

type FileResult struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

type Result struct {
	Target  string       `json:"target"`
	Files   []FileResult `json:"files"`
	Message string       `json:"message"`
}

func Install(root, target string) (Result, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	switch target {
	case "codex":
		return installCodex(root)
	case "claude":
		return installClaude(root)
	case "cursor":
		return installCursor(root)
	case "all":
		return installAll(root)
	default:
		return Result{}, fmt.Errorf("unknown install target %q; expected codex, claude, cursor, or all", target)
	}
}

func installAll(root string) (Result, error) {
	combined := Result{Target: "all", Files: []FileResult{}}
	for _, target := range []string{"codex", "claude", "cursor"} {
		result, err := Install(root, target)
		if err != nil {
			return Result{}, err
		}
		combined.Files = append(combined.Files, result.Files...)
	}
	combined.Message = brand.ProjectName + " integrations installed for Codex, Claude Code, and Cursor."
	return combined, nil
}

func installCodex(root string) (Result, error) {
	file, err := upsertSection(root, "AGENTS.md", codexSection())
	if err != nil {
		return Result{}, err
	}
	return Result{
		Target:  "codex",
		Files:   []FileResult{file},
		Message: "Codex integration installed.",
	}, nil
}

func installClaude(root string) (Result, error) {
	files := []FileResult{}
	for _, item := range []struct {
		path    string
		content string
	}{
		{"CLAUDE.md", claudeRootSection()},
		{filepath.ToSlash(filepath.Join(".claude", "skills", "nextvibe", "SKILL.md")), claudeSkillSection()},
		{filepath.ToSlash(filepath.Join(".claude", "commands", "nv-suggest.md")), claudeSuggestCommand()},
		{filepath.ToSlash(filepath.Join(".claude", "commands", "nv-check.md")), claudeCheckCommand()},
	} {
		file, err := upsertSection(root, item.path, item.content)
		if err != nil {
			return Result{}, err
		}
		files = append(files, file)
	}
	return Result{
		Target:  "claude",
		Files:   files,
		Message: "Claude Code integration installed.",
	}, nil
}

func installCursor(root string) (Result, error) {
	file, err := upsertSection(root, filepath.ToSlash(filepath.Join(".cursor", "rules", "nextvibe.mdc")), cursorRuleSection())
	if err != nil {
		return Result{}, err
	}
	return Result{
		Target:  "cursor",
		Files:   []FileResult{file},
		Message: "Cursor integration installed.",
	}, nil
}

func upsertSection(root, rel, section string) (FileResult, error) {
	absPath := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return FileResult{}, err
	}

	managed := sectionStart + "\n" + strings.TrimSpace(section) + "\n" + sectionEnd + "\n"
	existing, err := os.ReadFile(absPath)
	if os.IsNotExist(err) {
		if err := os.WriteFile(absPath, []byte(managed), 0o644); err != nil {
			return FileResult{}, err
		}
		return FileResult{Path: filepath.ToSlash(rel), Action: "created"}, nil
	}
	if err != nil {
		return FileResult{}, err
	}

	content := string(existing)
	next := ""
	if strings.Contains(content, sectionStart) && strings.Contains(content, sectionEnd) {
		next = replaceManagedSection(content, managed)
	} else {
		next = strings.TrimRight(content, "\r\n") + "\n\n" + managed
	}

	if next == content {
		return FileResult{Path: filepath.ToSlash(rel), Action: "unchanged"}, nil
	}
	if err := os.WriteFile(absPath, []byte(next), 0o644); err != nil {
		return FileResult{}, err
	}
	return FileResult{Path: filepath.ToSlash(rel), Action: "updated"}, nil
}

func replaceManagedSection(content, replacement string) string {
	start := strings.Index(content, sectionStart)
	end := strings.Index(content, sectionEnd)
	if start == -1 || end == -1 || end < start {
		return content
	}
	end += len(sectionEnd)
	next := content[:start] + strings.TrimRight(replacement, "\n") + content[end:]
	if !strings.HasSuffix(next, "\n") {
		next += "\n"
	}
	return next
}

func codexSection() string {
	return `## ` + brand.ProjectName + ` Agent Navigation

When the user asks to continue the project, decide the next step, recover direction, or proceed with development, use ` + brand.ProjectName + ` first.

Run:
- ` + "`" + brand.CommandName + ` scan --json` + "`" + `
- ` + "`" + brand.CommandName + ` suggest --json` + "`" + `
- ` + "`" + brand.CommandName + ` task --json` + "`" + `

Use the returned task boundaries.
Do not expand the task scope.

After editing, run:
- ` + "`" + brand.CommandName + ` check --json` + "`"
}

func claudeRootSection() string {
	return `## ` + brand.ProjectName + ` Project Navigation

Use ` + brand.ProjectName + ` when the user asks:
- continue this project
- what should I do next
- 下一步做什么
- 继续推进项目
- 项目卡住了
- 帮我规划接下来的开发

Run:
- ` + "`" + brand.CommandName + ` scan --json` + "`" + `
- ` + "`" + brand.CommandName + ` suggest --json` + "`" + `
- ` + "`" + brand.CommandName + ` task --json` + "`" + `

Respect the returned task boundaries and acceptance criteria.

After editing, run:
- ` + "`" + brand.CommandName + ` check --json` + "`"
}

func claudeSkillSection() string {
	return `# ` + brand.ProjectName + ` Project Navigation

Use this skill when the user asks to continue the project, recover direction, decide the next development step, or plan what to do after an AI-assisted build stalls.

Workflow:

1. Run ` + "`" + brand.CommandName + ` scan --json` + "`" + ` to inspect project state.
2. Run ` + "`" + brand.CommandName + ` suggest --json` + "`" + ` to get the next recommended task.
3. Run ` + "`" + brand.CommandName + ` task --json` + "`" + ` to create or read the active task.
4. Follow the task boundaries exactly.
5. After edits, run ` + "`" + brand.CommandName + ` check --json` + "`" + `.

Do not ask the user to copy a generated prompt into Claude Code. Call the local CLI directly and use the JSON result.`
}

func claudeSuggestCommand() string {
	return `# nv-suggest

Run the ` + brand.ProjectName + ` project navigation sequence:

` + "```bash" + `
` + brand.CommandName + ` scan --json
` + brand.CommandName + ` suggest --json
` + brand.CommandName + ` task --json
` + "```" + `

Use the returned task boundaries before editing.`
}

func claudeCheckCommand() string {
	return `# nv-check

Run:

` + "```bash" + `
` + brand.CommandName + ` check --json
` + "```" + `

Use the result to decide whether the active task is complete.`
}

func cursorRuleSection() string {
	return `---
description: Use ` + brand.ProjectName + ` before broad project continuation work
globs: "**/*"
alwaysApply: true
---

# ` + brand.ProjectName + `

Before continuing broad project development, inspect ` + brand.ProjectName + ` project state.

Use:
- ` + "`" + brand.CommandName + ` scan --json` + "`" + `
- ` + "`" + brand.CommandName + ` suggest --json` + "`" + `
- ` + "`" + brand.CommandName + ` task --json` + "`" + `

Respect task boundaries and acceptance criteria.

After editing, run:
- ` + "`" + brand.CommandName + ` check --json` + "`"
}
