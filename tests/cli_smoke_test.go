package tests

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
