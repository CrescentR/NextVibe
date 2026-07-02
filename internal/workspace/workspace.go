package workspace

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nextvibe/nextvibe/internal/brand"
)

type InitResult struct {
	WorkspaceDir string   `json:"workspaceDir"`
	Created      []string `json:"created"`
	Existing     []string `json:"existing"`
	Message      string   `json:"message"`
}

type fileTemplate struct {
	Path    string
	Content string
}

func Ensure(root string) (InitResult, error) {
	result := InitResult{
		WorkspaceDir: brand.WorkspaceDir,
		Created:      []string{},
		Existing:     []string{},
	}

	tasksDir := filepath.Join(root, brand.WorkspaceDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		return result, err
	}

	templates := []fileTemplate{
		{Path: filepath.Join(brand.WorkspaceDir, "project.md"), Content: projectTemplate()},
		{Path: filepath.Join(brand.WorkspaceDir, "stage.md"), Content: stageTemplate()},
		{Path: filepath.Join(brand.WorkspaceDir, "roadmap.md"), Content: roadmapTemplate()},
		{Path: filepath.Join(brand.WorkspaceDir, "current-task.md"), Content: currentTaskTemplate()},
		{Path: filepath.Join(brand.WorkspaceDir, "decisions.md"), Content: decisionsTemplate()},
		{Path: filepath.Join(brand.WorkspaceDir, "agent-context.md"), Content: agentContextTemplate()},
	}

	for _, template := range templates {
		absPath := filepath.Join(root, template.Path)
		if _, err := os.Stat(absPath); err == nil {
			result.Existing = append(result.Existing, filepath.ToSlash(template.Path))
			continue
		} else if !os.IsNotExist(err) {
			return result, err
		}
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return result, err
		}
		if err := os.WriteFile(absPath, []byte(template.Content), 0o644); err != nil {
			return result, err
		}
		result.Created = append(result.Created, filepath.ToSlash(template.Path))
	}

	taskRel := filepath.ToSlash(filepath.Join(brand.WorkspaceDir, "tasks"))
	if len(result.Created) == 0 {
		result.Existing = appendIfMissing(result.Existing, taskRel)
	} else {
		result.Created = appendIfMissing(result.Created, taskRel)
	}

	if len(result.Created) == 0 {
		result.Message = brand.ProjectName + " workspace already exists; no files were overwritten."
	} else {
		result.Message = brand.ProjectName + " workspace initialized; existing files were preserved."
	}

	return result, nil
}

func appendIfMissing(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append(values, value)
}

func projectTemplate() string {
	return strings.Join([]string{
		"# Project",
		"",
		"Use this file to capture the durable project goal, audience, and product boundaries.",
		"",
		"- Product name: " + brand.ProjectName,
		"- Primary workflow: coding agents call `" + brand.CommandName + "` to decide what to build next.",
		"- Model API usage: none",
		"",
	}, "\n")
}

func stageTemplate() string {
	return strings.Join([]string{
		"# Stage",
		"",
		"Current stage is detected by `" + brand.CommandName + " scan` and `" + brand.CommandName + " suggest`.",
		"",
		"Keep human notes here when a rule-based stage needs extra context.",
		"",
	}, "\n")
}

func roadmapTemplate() string {
	return strings.Join([]string{
		"# Roadmap",
		"",
		"1. Keep the local CLI useful and predictable.",
		"2. Improve project detection rules.",
		"3. Add richer agent integrations.",
		"4. Expose the same protocol through MCP without adding model calls.",
		"",
	}, "\n")
}

func currentTaskTemplate() string {
	return strings.Join([]string{
		"# Current Task",
		"",
		"No active task yet.",
		"",
		"Run:",
		"",
		"```bash",
		brand.CommandName + " task",
		"```",
		"",
	}, "\n")
}

func decisionsTemplate() string {
	return strings.Join([]string{
		"# Decisions",
		"",
		"Record durable product and engineering decisions here.",
		"",
		"- " + brand.ProjectName + " does not call model APIs.",
		"- " + brand.ProjectName + " is optimized for agent-readable output first.",
		"",
	}, "\n")
}

func agentContextTemplate() string {
	return strings.Join([]string{
		"# Agent Context",
		"",
		"When continuing this project, inspect local state first:",
		"",
		"```bash",
		brand.CommandName + " scan --json",
		brand.CommandName + " suggest --json",
		brand.CommandName + " task --json",
		"```",
		"",
		"Respect the returned task boundaries. After edits, run:",
		"",
		"```bash",
		brand.CommandName + " check --json",
		"```",
		"",
	}, "\n")
}
