package checker

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/taskgen"
)

type Check struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details,omitempty"`
}

type Result struct {
	TaskID     string  `json:"taskId"`
	Status     string  `json:"status"`
	Passed     bool    `json:"passed"`
	Checks     []Check `json:"checks"`
	NextAction string  `json:"nextAction"`
}

func CheckCurrent(root string) Result {
	task, ok := taskgen.LoadCurrent(root)
	if !ok {
		return Result{
			TaskID: "",
			Status: "no-active-task",
			Passed: false,
			Checks: []Check{
				{Name: "active task exists", Passed: false},
			},
			NextAction: "Run " + brand.CommandName + " task before checking task completion.",
		}
	}

	checks := []Check{}
	for _, rel := range task.AllowedFiles {
		if strings.HasSuffix(rel, "/") {
			checks = append(checks, Check{
				Name:   rel + " directory exists",
				Passed: dirExists(root, rel),
			})
			continue
		}
		checks = append(checks, Check{
			Name:   rel + " exists",
			Passed: fileExists(root, rel),
		})
	}

	for _, check := range changedOutsideAllowed(root, task.AllowedFiles) {
		checks = append(checks, check)
	}

	passed := len(checks) > 0
	for _, check := range checks {
		if !check.Passed {
			passed = false
			break
		}
	}

	status := "complete"
	nextAction := "All basic checks passed. Review the task acceptance criteria before marking it complete."
	if !passed {
		status = "incomplete"
		nextAction = nextActionFor(checks)
	}

	return Result{
		TaskID:     task.TaskID,
		Status:     status,
		Passed:     passed,
		Checks:     checks,
		NextAction: nextAction,
	}
}

func fileExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(strings.TrimSuffix(rel, "/"))))
	return err == nil && !info.IsDir()
}

func dirExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(strings.TrimSuffix(rel, "/"))))
	return err == nil && info.IsDir()
}

func changedOutsideAllowed(root string, allowed []string) []Check {
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		return nil
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	allowedPrefixes := map[string]bool{
		brand.WorkspaceDir + "/": true,
	}
	for _, rel := range allowed {
		clean := filepath.ToSlash(strings.TrimSpace(rel))
		if clean == "" {
			continue
		}
		if strings.HasSuffix(clean, "/") {
			allowedPrefixes[clean] = true
			continue
		}
		allowedPrefixes[clean] = true
	}

	unexpected := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		path = strings.Trim(path, `"`)
		path = filepath.ToSlash(path)
		if path == "" || isLocalNoise(path) {
			continue
		}
		if isAllowed(path, allowedPrefixes) {
			continue
		}
		unexpected = append(unexpected, path)
	}

	if len(unexpected) == 0 {
		return nil
	}
	return []Check{{
		Name:    "changed files stay inside task boundaries",
		Passed:  false,
		Details: "Unexpected changed paths: " + strings.Join(unexpected, ", "),
	}}
}

func isLocalNoise(path string) bool {
	return path == ".gitignore" ||
		strings.HasPrefix(path, ".idea/") ||
		strings.HasPrefix(path, ".vscode/") ||
		path == ".DS_Store" ||
		strings.HasSuffix(path, "/.DS_Store")
}

func isAllowed(path string, allowed map[string]bool) bool {
	for prefix := range allowed {
		if strings.HasSuffix(prefix, "/") {
			if strings.HasPrefix(path, prefix) {
				return true
			}
			continue
		}
		if path == prefix {
			return true
		}
	}
	return false
}

func nextActionFor(checks []Check) string {
	for _, check := range checks {
		if !check.Passed {
			if check.Details != "" {
				return check.Details
			}
			return "Complete: " + check.Name + "."
		}
	}
	return "Review acceptance criteria."
}
