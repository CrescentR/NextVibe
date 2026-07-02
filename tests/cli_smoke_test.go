package tests

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nextvibe/nextvibe/internal/cli"
)

func TestInstallAllJSONSupportsFlagAfterTarget(t *testing.T) {
	withTempCWD(t, func(root string) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"install", "all", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("install returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			Target string `json:"target"`
			Files  []struct {
				Path   string `json:"path"`
				Action string `json:"action"`
			} `json:"files"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("install output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.Target != "all" {
			t.Fatalf("target = %q, want all", result.Target)
		}
		if len(result.Files) != 6 {
			t.Fatalf("files len = %d, want 6", len(result.Files))
		}

		for _, rel := range []string{
			"AGENTS.md",
			"CLAUDE.md",
			filepath.Join(".claude", "skills", "nextvibe", "SKILL.md"),
			filepath.Join(".claude", "commands", "nv-suggest.md"),
			filepath.Join(".claude", "commands", "nv-check.md"),
			filepath.Join(".cursor", "rules", "nextvibe.mdc"),
		} {
			if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
				t.Fatalf("expected generated file %s: %v", rel, err)
			}
		}
	})
}

func TestTaskJSONCreatesAPIContractTaskForFrontendPrototype(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		writeFile(t, root, "package.json", `{"name":"example","dependencies":{"react":"^18.0.0"}}`)
		writeFile(t, root, filepath.Join("mock", "items.json"), `[{"id":"item-1"}]`)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"task", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("task returned %d, stderr: %s", code, stderr.String())
		}

		var task struct {
			TaskID       string   `json:"taskId"`
			Title        string   `json:"title"`
			AllowedFiles []string `json:"allowedFiles"`
			TaskFile     string   `json:"taskFile"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &task); err != nil {
			t.Fatalf("task output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if task.TaskID != "001" {
			t.Fatalf("task id = %q, want 001", task.TaskID)
		}
		if task.Title != "Design the first API contract" {
			t.Fatalf("title = %q", task.Title)
		}
		if !contains(task.AllowedFiles, "docs/api/openapi.yaml") {
			t.Fatalf("allowed files missing docs/api/openapi.yaml: %#v", task.AllowedFiles)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(task.TaskFile))); err != nil {
			t.Fatalf("expected generated task file %s: %v", task.TaskFile, err)
		}
	})
}

func TestScanTextSupportsChineseLanguage(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		writeFile(t, root, "go.mod", "module example\n\ngo 1.22\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--lang", "zh"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
		}
		output := stdout.String()
		for _, want := range []string{"NextVibe 扫描", "项目：", "阶段：", "技术栈：", "测试命令：", "go test ./..."} {
			if !strings.Contains(output, want) {
				t.Fatalf("scan output missing %q:\n%s", want, output)
			}
		}
	})
}

func TestScanJSONDetectsTestCommands(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		writeFile(t, root, "go.mod", "module example\n\ngo 1.22\n")
		writeFile(t, root, "package.json", `{"name":"example","scripts":{"test":"vitest run"}}`)
		writeFile(t, root, "pom.xml", "<project></project>\n")
		writeFile(t, root, "build.gradle", "plugins { id 'java' }\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			TestCommands []string `json:"testCommands"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("scan output is not valid JSON: %v\n%s", err, stdout.String())
		}
		for _, want := range []string{"go test ./...", "npm test", "mvn test", "gradle test"} {
			if !contains(result.TestCommands, want) {
				t.Fatalf("test commands missing %q: %#v", want, result.TestCommands)
			}
		}
	})
}

func TestScanJSONSuppressesAPIAndDatabaseRisksForCLIOnlyGoProject(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		writeFile(t, root, "go.mod", "module example\n\ngo 1.22\n")
		writeFile(t, root, filepath.Join("cmd", "example", "main.go"), "package main\n\nfunc main() {}\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			Risks []string `json:"risks"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("scan output is not valid JSON: %v\n%s", err, stdout.String())
		}
		for _, risk := range []string{"No API contract found", "No database schema found"} {
			if contains(result.Risks, risk) {
				t.Fatalf("CLI-only project reported %q: %#v", risk, result.Risks)
			}
		}
	})
}

func TestScanJSONKeepsAPIAndDatabaseRisksForProductSurfaces(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, root string)
	}{
		{
			name: "frontend mock prototype",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "package.json", `{"name":"example","dependencies":{"react":"^18.0.0"}}`)
				writeFile(t, root, filepath.Join("mock", "items.json"), `[{"id":"item-1"}]`)
			},
		},
		{
			name: "backend service directory",
			setup: func(t *testing.T, root string) {
				writeFile(t, root, "go.mod", "module example\n\ngo 1.22\n")
				writeFile(t, root, filepath.Join("server", "main.go"), "package server\n")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withTempCWD(t, func(root string) {
				writeFile(t, root, "README.md", "# Example\n")
				tc.setup(t, root)

				var stdout bytes.Buffer
				var stderr bytes.Buffer

				code := cli.Run([]string{"scan", "--json"}, &stdout, &stderr)
				if code != 0 {
					t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
				}

				var result struct {
					Risks []string `json:"risks"`
				}
				if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
					t.Fatalf("scan output is not valid JSON: %v\n%s", err, stdout.String())
				}
				for _, risk := range []string{"No API contract found", "No database schema found"} {
					if !contains(result.Risks, risk) {
						t.Fatalf("risks missing %q: %#v", risk, result.Risks)
					}
				}
			})
		})
	}
}

func TestSuggestJSONGuidesReadyProjectPlanning(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"suggest", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("suggest returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			Recommendation struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Reason string `json:"reason"`
			} `json:"recommendation"`
			AcceptanceCriteria []string `json:"acceptanceCriteria"`
			AgentInstructions  struct {
				AllowedActions   []string `json:"allowedActions"`
				ForbiddenActions []string `json:"forbiddenActions"`
			} `json:"agentInstructions"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("suggest output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.Recommendation.ID != "plan-next-development-slice" {
			t.Fatalf("recommendation id = %q", result.Recommendation.ID)
		}
		if result.Recommendation.Title != "Research and write the next development plan" {
			t.Fatalf("recommendation title = %q", result.Recommendation.Title)
		}
		if !strings.Contains(result.Recommendation.Reason, "inspect local context") {
			t.Fatalf("recommendation reason does not guide local research: %q", result.Recommendation.Reason)
		}
		for _, want := range []string{
			"The plan explains why this slice follows from the scan, roadmap, vision, and current task state",
			"Verification commands are listed before implementation starts",
		} {
			if !contains(result.AcceptanceCriteria, want) {
				t.Fatalf("acceptance criteria missing %q: %#v", want, result.AcceptanceCriteria)
			}
		}
		for _, want := range []string{
			"read nextvibe scan --json output",
			"read docs/vision.md if present",
			"read docs/roadmap.md if present",
		} {
			if !contains(result.AgentInstructions.AllowedActions, want) {
				t.Fatalf("allowed actions missing %q: %#v", want, result.AgentInstructions.AllowedActions)
			}
		}
		if !contains(result.AgentInstructions.ForbiddenActions, "do not implement the planned slice before writing the plan") {
			t.Fatalf("forbidden actions do not block premature implementation: %#v", result.AgentInstructions.ForbiddenActions)
		}
	})
}

func TestTaskJSONCreatesPlanningTaskForReadyProject(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"task", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("task returned %d, stderr: %s", code, stderr.String())
		}

		var task struct {
			TaskID string `json:"taskId"`
			Title  string `json:"title"`
			Goal   string `json:"goal"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &task); err != nil {
			t.Fatalf("task output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if task.TaskID != "001" {
			t.Fatalf("task id = %q, want 001", task.TaskID)
		}
		if task.Title != "Research and write the next development plan" {
			t.Fatalf("task title = %q", task.Title)
		}
		if task.Goal != "Research the local project state and write the next bounded development plan before implementation." {
			t.Fatalf("task goal = %q", task.Goal)
		}
	})
}

func TestTaskJSONCreatesNextTaskAfterCompletedCurrentTask(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)
		previousTaskFile := writeCurrentTaskFixture(t, root, "001", "complete", "Done task")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"task", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("task returned %d, stderr: %s", code, stderr.String())
		}

		var task struct {
			TaskID   string `json:"taskId"`
			Title    string `json:"title"`
			Status   string `json:"status"`
			TaskFile string `json:"taskFile"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &task); err != nil {
			t.Fatalf("task output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if task.TaskID != "002" {
			t.Fatalf("task id = %q, want 002", task.TaskID)
		}
		if task.Status != "active" {
			t.Fatalf("task status = %q, want active", task.Status)
		}
		if task.Title != "Research and write the next development plan" {
			t.Fatalf("task title = %q", task.Title)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(previousTaskFile))); err != nil {
			t.Fatalf("previous task file was not preserved: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(task.TaskFile))); err != nil {
			t.Fatalf("expected generated task file %s: %v", task.TaskFile, err)
		}
	})
}

func TestTaskJSONReusesActiveCurrentTask(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)
		writeCurrentTaskFixture(t, root, "001", "active", "Keep working")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"task", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("task returned %d, stderr: %s", code, stderr.String())
		}

		var task struct {
			TaskID string `json:"taskId"`
			Title  string `json:"title"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &task); err != nil {
			t.Fatalf("task output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if task.TaskID != "001" {
			t.Fatalf("task id = %q, want 001", task.TaskID)
		}
		if task.Title != "Keep working" {
			t.Fatalf("task title = %q", task.Title)
		}
		unexpectedTask := filepath.Join(root, ".nextvibe", "tasks", "002-plan-next-development-slice.md")
		if _, err := os.Stat(unexpectedTask); !os.IsNotExist(err) {
			t.Fatalf("active task should not create %s", unexpectedTask)
		}
	})
}

func TestLanguageAliasSupportsChineseLanguage(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--language", "zh"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "项目：") {
			t.Fatalf("scan output did not use Chinese labels:\n%s", stdout.String())
		}
	})
}

func TestInvalidLanguageFails(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--lang", "fr"}, &stdout, &stderr)
		if code == 0 {
			t.Fatalf("scan returned success, stdout: %s", stdout.String())
		}
		if !strings.Contains(stderr.String(), "unsupported language") {
			t.Fatalf("stderr missing unsupported language error:\n%s", stderr.String())
		}
	})
}

func TestJSONOutputRemainsStableWithLanguageFlag(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--json", "--lang", "zh"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			ProjectName string `json:"projectName"`
			Stage       struct {
				ID    string `json:"id"`
				Label string `json:"label"`
			} `json:"stage"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("scan output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.ProjectName == "" {
			t.Fatal("projectName is empty")
		}
		if result.Stage.ID == "" || result.Stage.Label == "" {
			t.Fatalf("stage missing stable JSON fields: %#v", result.Stage)
		}
	})
}

func withTempCWD(t *testing.T, fn func(root string)) {
	t.Helper()
	root := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previous)
	})
	fn(root)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeReadyCLIProject(t *testing.T, root string) {
	t.Helper()
	writeFile(t, root, "README.md", "# Example\n")
	writeFile(t, root, "go.mod", "module example\n\ngo 1.22\n")
	writeFile(t, root, "Dockerfile", "FROM scratch\n")
	writeFile(t, root, "AGENTS.md", "# Agent instructions\n")
	writeFile(t, root, filepath.Join("cmd", "example", "main.go"), "package main\n\nfunc main() {}\n")
	writeFile(t, root, filepath.Join("cmd", "example", "main_test.go"), "package main\n\nimport \"testing\"\n\nfunc TestMainPackage(t *testing.T) {}\n")
}

func writeCurrentTaskFixture(t *testing.T, root, id, status, title string) string {
	t.Helper()
	taskFile := filepath.ToSlash(filepath.Join(".nextvibe", "tasks", id+"-existing-task.md"))
	writeFile(t, root, taskFile, "# "+id+" - "+title+"\n\n"+
		"Task ID: "+id+"\n"+
		"Status: "+status+"\n"+
		"Title: "+title+"\n\n"+
		"## Background\n\nExisting task.\n\n"+
		"## Goal\n\nKeep the current workflow stable.\n\n"+
		"## Allowed Files\n\n- README.md\n\n"+
		"## Forbidden Changes\n\n- do not expand scope\n\n"+
		"## Acceptance Criteria\n\n- Existing criterion\n")
	writeFile(t, root, filepath.Join(".nextvibe", "current-task.md"), "# Current Task\n\n"+
		"Task ID: "+id+"\n"+
		"Status: "+status+"\n"+
		"Title: "+title+"\n"+
		"Task file: "+taskFile+"\n\n"+
		"Run:\n\n```bash\nnextvibe task --json\nnextvibe check --json\n```\n")
	return taskFile
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
