package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nextvibe/nextvibe/internal/cli"
	"github.com/nextvibe/nextvibe/internal/mcp"
	"github.com/nextvibe/nextvibe/internal/protocol"
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

func TestInstallVerifyJSONChecksGeneratedIntegrations(t *testing.T) {
	withTempCWD(t, func(root string) {
		var installStdout bytes.Buffer
		var installStderr bytes.Buffer
		if code := cli.Run([]string{"install", "all", "--json"}, &installStdout, &installStderr); code != 0 {
			t.Fatalf("install returned %d, stderr: %s", code, installStderr.String())
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := cli.Run([]string{"install", "verify", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("install verify returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			ProtocolVersion string `json:"protocolVersion"`
			Target          string `json:"target"`
			Passed          bool   `json:"passed"`
			Checks          []struct {
				Name   string `json:"name"`
				Path   string `json:"path"`
				Passed bool   `json:"passed"`
			} `json:"checks"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("verify output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.ProtocolVersion != protocol.Version {
			t.Fatalf("protocol version = %q, want %q", result.ProtocolVersion, protocol.Version)
		}
		if result.Target != "all" || !result.Passed {
			t.Fatalf("verify result = %#v", result)
		}
		for _, want := range []struct {
			path string
			name string
		}{
			{"AGENTS.md", "contains nextvibe task --json"},
			{filepath.ToSlash(filepath.Join(".claude", "skills", "nextvibe", "SKILL.md")), "contains Workflow:"},
			{filepath.ToSlash(filepath.Join(".claude", "commands", "nv-check.md")), "contains nextvibe check --json"},
			{filepath.ToSlash(filepath.Join(".cursor", "rules", "nextvibe.mdc")), "contains alwaysApply: true"},
		} {
			if !hasVerificationCheck(result.Checks, want.path, want.name, true) {
				t.Fatalf("verify checks missing %s %s: %#v", want.path, want.name, result.Checks)
			}
		}
	})
}

func TestInstallVerifyJSONFailsWhenGeneratedCommandIsBroken(t *testing.T) {
	withTempCWD(t, func(root string) {
		var installStdout bytes.Buffer
		var installStderr bytes.Buffer
		if code := cli.Run([]string{"install", "all", "--json"}, &installStdout, &installStderr); code != 0 {
			t.Fatalf("install returned %d, stderr: %s", code, installStderr.String())
		}

		brokenCommand := filepath.Join(".claude", "commands", "nv-check.md")
		writeFile(t, root, brokenCommand, "<!-- NEXTVIBE:START -->\n# nv-check\n<!-- NEXTVIBE:END -->\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := cli.Run([]string{"install", "verify", "claude", "--json"}, &stdout, &stderr)
		if code == 0 {
			t.Fatalf("install verify returned success, stdout: %s", stdout.String())
		}

		var result struct {
			Target string `json:"target"`
			Passed bool   `json:"passed"`
			Checks []struct {
				Name   string `json:"name"`
				Path   string `json:"path"`
				Passed bool   `json:"passed"`
			} `json:"checks"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("verify output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.Target != "claude" || result.Passed {
			t.Fatalf("verify result = %#v", result)
		}
		if !hasVerificationCheck(result.Checks, filepath.ToSlash(brokenCommand), "contains nextvibe check --json", false) {
			t.Fatalf("verify did not report missing claude check command: %#v", result.Checks)
		}
	})
}

func TestInstallVerifyJSONFailsBeforeInstall(t *testing.T) {
	withTempCWD(t, func(root string) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := cli.Run([]string{"install", "verify", "cursor", "--json"}, &stdout, &stderr)
		if code == 0 {
			t.Fatalf("install verify returned success, stdout: %s", stdout.String())
		}

		var result struct {
			Target string `json:"target"`
			Passed bool   `json:"passed"`
			Checks []struct {
				Name   string `json:"name"`
				Path   string `json:"path"`
				Passed bool   `json:"passed"`
			} `json:"checks"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("verify output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.Target != "cursor" || result.Passed {
			t.Fatalf("verify result = %#v", result)
		}
		if !hasVerificationCheck(result.Checks, filepath.ToSlash(filepath.Join(".cursor", "rules", "nextvibe.mdc")), "file exists", false) {
			t.Fatalf("verify did not report missing cursor rule: %#v", result.Checks)
		}
	})
}

func TestInitJSONCreatesConfigAndRuleTemplate(t *testing.T) {
	withTempCWD(t, func(root string) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := cli.Run([]string{"init", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("init returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			Created []string `json:"created"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("init output is not valid JSON: %v\n%s", err, stdout.String())
		}
		for _, want := range []string{
			filepath.ToSlash(filepath.Join(".nextvibe", "config.yaml")),
			filepath.ToSlash(filepath.Join(".nextvibe", "rules", "default-task.md")),
		} {
			if !contains(result.Created, want) {
				t.Fatalf("init created missing %q: %#v", want, result.Created)
			}
		}

		assertFileContains(t, root, filepath.Join(".nextvibe", "config.yaml"), "version: "+protocol.Version)
		assertFileContains(t, root, filepath.Join(".nextvibe", "config.yaml"), "requiredCommandsSection: Required Commands")
		assertFileContains(t, root, filepath.Join(".nextvibe", "rules", "default-task.md"), "## Required Commands")
		assertFileContains(t, root, filepath.Join(".nextvibe", "rules", "default-task.md"), "## Evidence Files")
	})
}

func TestCheckJSONFailsWhenDeclaredEvidenceFileIsMissing(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		writeCurrentTaskFixtureWithSections(t, root, "001", "active", "Need evidence", "", "## Evidence Files\n\n- docs/done.md\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := cli.Run([]string{"check", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("check returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			Status string `json:"status"`
			Passed bool   `json:"passed"`
			Checks []struct {
				Name   string `json:"name"`
				Passed bool   `json:"passed"`
			} `json:"checks"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("check output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.Status != "incomplete" || result.Passed {
			t.Fatalf("check result = %#v", result)
		}
		if !hasCheck(result.Checks, "evidence file docs/done.md exists", false) {
			t.Fatalf("check did not report missing evidence file: %#v", result.Checks)
		}
	})
}

func TestCheckJSONPassesDeclaredEvidenceAndRequiredCommand(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		writeFile(t, root, "go.mod", "module example\n\ngo 1.22\n")
		writeFile(t, root, "main_test.go", "package example\n\nimport \"testing\"\n\nfunc TestExample(t *testing.T) {}\n")
		writeFile(t, root, filepath.Join("docs", "done.md"), "done\n")
		writeCurrentTaskFixtureWithSections(t, root, "001", "active", "Evidence complete",
			"## Required Commands\n\n- go test ./...\n",
			"## Evidence Files\n\n- docs/done.md\n",
		)

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := cli.Run([]string{"check", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("check returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			Status string `json:"status"`
			Passed bool   `json:"passed"`
			Checks []struct {
				Name   string `json:"name"`
				Passed bool   `json:"passed"`
			} `json:"checks"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("check output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.Status != "complete" || !result.Passed {
			t.Fatalf("check result = %#v", result)
		}
		if !hasCheck(result.Checks, "evidence file docs/done.md exists", true) {
			t.Fatalf("check did not report evidence file success: %#v", result.Checks)
		}
		if !hasCheck(result.Checks, "required command go test ./...", true) {
			t.Fatalf("check did not report required command success: %#v", result.Checks)
		}
	})
}

func TestMCPServerListsAndCallsScanTool(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)

		input := strings.Join([]string{
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0.0.0"}}}`,
			`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
			`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"nextvibe_scan","arguments":{}}}`,
			"",
		}, "\n")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if err := mcp.Serve(root, strings.NewReader(input), &stdout, &stderr); err != nil {
			t.Fatalf("mcp serve returned error: %v, stderr: %s", err, stderr.String())
		}

		responses := decodeJSONRPCResponses(t, stdout.String())
		if len(responses) != 3 {
			t.Fatalf("response len = %d, want 3\n%s", len(responses), stdout.String())
		}
		if valueAt(responses[0], "result", "serverInfo", "name") != "nextvibe" {
			t.Fatalf("initialize response = %#v", responses[0])
		}
		tools, ok := valueAt(responses[1], "result", "tools").([]any)
		if !ok || !hasMCPTool(tools, "nextvibe_scan") || !hasMCPTool(tools, "nextvibe_check") {
			t.Fatalf("tools/list response = %#v", responses[1])
		}
		if valueAt(responses[2], "result", "structuredContent", "protocolVersion") != protocol.Version {
			t.Fatalf("tools/call scan response = %#v", responses[2])
		}
		if valueAt(responses[2], "result", "structuredContent", "projectName") == "" {
			t.Fatalf("scan response missing projectName: %#v", responses[2])
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

func TestScanJSONIncludesProtocolVersionAndWritesState(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"scan", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("scan returned %d, stderr: %s", code, stderr.String())
		}

		var result struct {
			ProtocolVersion string `json:"protocolVersion"`
			ProjectName     string `json:"projectName"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("scan output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if result.ProtocolVersion != protocol.Version {
			t.Fatalf("protocol version = %q, want %q", result.ProtocolVersion, protocol.Version)
		}

		state := readStateFile(t, root)
		if state.ProtocolVersion != protocol.Version {
			t.Fatalf("state protocol version = %q, want %q", state.ProtocolVersion, protocol.Version)
		}
		if state.ProjectName != result.ProjectName {
			t.Fatalf("state project name = %q, want %q", state.ProjectName, result.ProjectName)
		}
		if state.Scan == nil || state.Scan.ProtocolVersion != protocol.Version {
			t.Fatalf("state scan missing protocol version: %#v", state.Scan)
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
			ProtocolVersion string `json:"protocolVersion"`
			Recommendation  struct {
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
		if result.ProtocolVersion != protocol.Version {
			t.Fatalf("protocol version = %q, want %q", result.ProtocolVersion, protocol.Version)
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
			ProtocolVersion string `json:"protocolVersion"`
			TaskID          string `json:"taskId"`
			Title           string `json:"title"`
			Goal            string `json:"goal"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &task); err != nil {
			t.Fatalf("task output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if task.ProtocolVersion != protocol.Version {
			t.Fatalf("protocol version = %q, want %q", task.ProtocolVersion, protocol.Version)
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

func TestTaskAndCheckUpdateStateJSON(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeReadyCLIProject(t, root)

		var taskStdout bytes.Buffer
		var taskStderr bytes.Buffer
		if code := cli.Run([]string{"task", "--json"}, &taskStdout, &taskStderr); code != 0 {
			t.Fatalf("task returned %d, stderr: %s", code, taskStderr.String())
		}

		stateAfterTask := readStateFile(t, root)
		if stateAfterTask.Suggestion == nil || stateAfterTask.Suggestion.ProtocolVersion != protocol.Version {
			t.Fatalf("state suggestion missing protocol version: %#v", stateAfterTask.Suggestion)
		}
		if stateAfterTask.CurrentTask == nil || stateAfterTask.CurrentTask.TaskID != "001" {
			t.Fatalf("state current task = %#v", stateAfterTask.CurrentTask)
		}

		var checkStdout bytes.Buffer
		var checkStderr bytes.Buffer
		if code := cli.Run([]string{"check", "--json"}, &checkStdout, &checkStderr); code != 0 {
			t.Fatalf("check returned %d, stderr: %s", code, checkStderr.String())
		}

		var checkResult struct {
			ProtocolVersion string `json:"protocolVersion"`
			TaskID          string `json:"taskId"`
		}
		if err := json.Unmarshal(checkStdout.Bytes(), &checkResult); err != nil {
			t.Fatalf("check output is not valid JSON: %v\n%s", err, checkStdout.String())
		}
		if checkResult.ProtocolVersion != protocol.Version {
			t.Fatalf("check protocol version = %q, want %q", checkResult.ProtocolVersion, protocol.Version)
		}

		stateAfterCheck := readStateFile(t, root)
		if stateAfterCheck.Check == nil || stateAfterCheck.Check.TaskID != checkResult.TaskID {
			t.Fatalf("state check = %#v, check result = %#v", stateAfterCheck.Check, checkResult)
		}
		if stateAfterCheck.CurrentTask == nil || stateAfterCheck.CurrentTask.Status != stateAfterCheck.Check.Status {
			t.Fatalf("state current task status did not follow check: task=%#v check=%#v", stateAfterCheck.CurrentTask, stateAfterCheck.Check)
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

func TestTaskCompleteFailsWhenChecksDoNotPass(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeCurrentTaskFixture(t, root, "001", "active", "Blocked task")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"task", "complete", "--json"}, &stdout, &stderr)
		if code == 0 {
			t.Fatalf("task complete returned success, stdout: %s", stdout.String())
		}
		if !strings.Contains(stderr.String(), "current task checks did not pass") {
			t.Fatalf("stderr missing failed check reason:\n%s", stderr.String())
		}
		if _, err := os.Stat(filepath.Join(root, ".nextvibe", "task-history.json")); !os.IsNotExist(err) {
			t.Fatalf("task-history.json should not exist after failed completion")
		}
	})
}

func TestTaskCompleteMarksTaskAndWritesHistory(t *testing.T) {
	withTempCWD(t, func(root string) {
		writeFile(t, root, "README.md", "# Example\n")
		taskFile := writeCurrentTaskFixture(t, root, "001", "active", "Finish me")

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		code := cli.Run([]string{"task", "complete", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("task complete returned %d, stderr: %s", code, stderr.String())
		}

		var task struct {
			ProtocolVersion string `json:"protocolVersion"`
			TaskID          string `json:"taskId"`
			Status          string `json:"status"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &task); err != nil {
			t.Fatalf("task complete output is not valid JSON: %v\n%s", err, stdout.String())
		}
		if task.ProtocolVersion != protocol.Version {
			t.Fatalf("protocol version = %q, want %q", task.ProtocolVersion, protocol.Version)
		}
		if task.TaskID != "001" || task.Status != "complete" {
			t.Fatalf("completed task = %#v", task)
		}

		assertFileContains(t, root, filepath.Join(".nextvibe", "current-task.md"), "Status: complete")
		assertFileContains(t, root, filepath.FromSlash(taskFile), "Status: complete")

		var historyStdout bytes.Buffer
		var historyStderr bytes.Buffer
		code = cli.Run([]string{"task", "history", "--json"}, &historyStdout, &historyStderr)
		if code != 0 {
			t.Fatalf("task history returned %d, stderr: %s", code, historyStderr.String())
		}

		var history struct {
			ProtocolVersion string `json:"protocolVersion"`
			Tasks           []struct {
				TaskID string `json:"taskId"`
				Status string `json:"status"`
			} `json:"tasks"`
		}
		if err := json.Unmarshal(historyStdout.Bytes(), &history); err != nil {
			t.Fatalf("task history output is not valid JSON: %v\n%s", err, historyStdout.String())
		}
		if history.ProtocolVersion != protocol.Version {
			t.Fatalf("history protocol version = %q, want %q", history.ProtocolVersion, protocol.Version)
		}
		if len(history.Tasks) != 1 || history.Tasks[0].TaskID != "001" || history.Tasks[0].Status != "complete" {
			t.Fatalf("history tasks = %#v", history.Tasks)
		}

		state := readStateFile(t, root)
		if state.TaskHistory == nil || len(state.TaskHistory.Tasks) != 1 {
			t.Fatalf("state task history = %#v", state.TaskHistory)
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

func assertFileContains(t *testing.T, root, rel, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s missing %q:\n%s", rel, want, string(data))
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
	return writeCurrentTaskFixtureWithSections(t, root, id, status, title, "", "")
}

func writeCurrentTaskFixtureWithSections(t *testing.T, root, id, status, title, requiredCommands, evidenceFiles string) string {
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
		"## Acceptance Criteria\n\n- Existing criterion\n\n"+
		requiredCommands+
		evidenceFiles)
	writeFile(t, root, filepath.Join(".nextvibe", "current-task.md"), "# Current Task\n\n"+
		"Task ID: "+id+"\n"+
		"Status: "+status+"\n"+
		"Title: "+title+"\n"+
		"Task file: "+taskFile+"\n\n"+
		"Run:\n\n```bash\nnextvibe task --json\nnextvibe check --json\n```\n")
	return taskFile
}

type stateFile struct {
	ProtocolVersion string `json:"protocolVersion"`
	ProjectName     string `json:"projectName"`
	Scan            *struct {
		ProtocolVersion string `json:"protocolVersion"`
	} `json:"scan"`
	Suggestion *struct {
		ProtocolVersion string `json:"protocolVersion"`
	} `json:"suggestion"`
	CurrentTask *struct {
		ProtocolVersion string `json:"protocolVersion"`
		TaskID          string `json:"taskId"`
		Status          string `json:"status"`
	} `json:"currentTask"`
	TaskHistory *struct {
		ProtocolVersion string `json:"protocolVersion"`
		Tasks           []struct {
			TaskID string `json:"taskId"`
			Status string `json:"status"`
		} `json:"tasks"`
	} `json:"taskHistory"`
	Check *struct {
		ProtocolVersion string `json:"protocolVersion"`
		TaskID          string `json:"taskId"`
		Status          string `json:"status"`
	} `json:"check"`
}

func readStateFile(t *testing.T, root string) stateFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".nextvibe", "state.json"))
	if err != nil {
		t.Fatalf("expected state.json: %v", err)
	}
	var state stateFile
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("state.json is not valid JSON: %v\n%s", err, string(data))
	}
	return state
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func hasVerificationCheck(checks []struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Passed bool   `json:"passed"`
}, path, name string, passed bool) bool {
	for _, check := range checks {
		if check.Path == path && check.Name == name && check.Passed == passed {
			return true
		}
	}
	return false
}

func hasCheck(checks []struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}, name string, passed bool) bool {
	for _, check := range checks {
		if check.Name == name && check.Passed == passed {
			return true
		}
	}
	return false
}

func decodeJSONRPCResponses(t *testing.T, output string) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(output))
	responses := []map[string]any{}
	for {
		var response map[string]any
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("invalid JSON-RPC response stream: %v\n%s", err, output)
		}
		responses = append(responses, response)
	}
	return responses
}

func valueAt(value any, path ...string) any {
	current := value
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[key]
	}
	return current
}

func hasMCPTool(tools []any, name string) bool {
	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if ok && tool["name"] == name {
			return true
		}
	}
	return false
}
