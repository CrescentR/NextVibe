package checker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/protocol"
	"github.com/nextvibe/nextvibe/internal/taskgen"
)

type Check struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details,omitempty"`
}

type Result struct {
	ProtocolVersion string  `json:"protocolVersion"`
	TaskID          string  `json:"taskId"`
	Status          string  `json:"status"`
	Passed          bool    `json:"passed"`
	Checks          []Check `json:"checks"`
	NextAction      string  `json:"nextAction"`
}

func CheckCurrent(root string) Result {
	task, ok := taskgen.LoadCurrent(root)
	if !ok {
		return Result{
			ProtocolVersion: protocol.Version,
			TaskID:          "",
			Status:          "no-active-task",
			Passed:          false,
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

	requirements := loadEvidenceRequirements(root, task.TaskFile)
	for _, rel := range requirements.EvidenceFiles {
		checks = append(checks, evidenceFileCheck(root, rel))
	}
	for _, command := range requirements.RequiredCommands {
		checks = append(checks, requiredCommandCheck(root, command))
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
	nextAction := "All completion evidence checks passed. Review the task acceptance criteria before marking it complete."
	if !passed {
		status = "incomplete"
		nextAction = nextActionFor(checks)
	}

	return Result{
		ProtocolVersion: protocol.Version,
		TaskID:          task.TaskID,
		Status:          status,
		Passed:          passed,
		Checks:          checks,
		NextAction:      nextAction,
	}
}

type evidenceRequirements struct {
	RequiredCommands []string
	EvidenceFiles    []string
}

func loadEvidenceRequirements(root, taskFile string) evidenceRequirements {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(taskFile)))
	if err != nil {
		return evidenceRequirements{}
	}
	content := string(data)
	return evidenceRequirements{
		RequiredCommands: readListSection(content, "Required Commands"),
		EvidenceFiles:    readListSection(content, "Evidence Files"),
	}
}

func evidenceFileCheck(root, rel string) Check {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return Check{Name: "evidence file path is not empty", Passed: false}
	}
	if strings.HasSuffix(rel, "/") {
		return Check{
			Name:   "evidence directory " + rel + " exists",
			Passed: dirExists(root, rel),
		}
	}
	return Check{
		Name:   "evidence file " + rel + " exists",
		Passed: fileExists(root, rel),
	}
}

func requiredCommandCheck(root, command string) Check {
	command = strings.TrimSpace(command)
	if command == "" {
		return Check{Name: "required command is not empty", Passed: false}
	}
	fields, err := splitCommand(command)
	if err != nil {
		return Check{
			Name:    "required command " + command,
			Passed:  false,
			Details: err.Error(),
		}
	}
	if len(fields) == 0 {
		return Check{Name: "required command is not empty", Passed: false}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return Check{
			Name:    "required command " + command,
			Passed:  false,
			Details: "command timed out after 2m",
		}
	}
	if err != nil {
		details := strings.TrimSpace(string(output))
		if details == "" {
			details = err.Error()
		}
		return Check{
			Name:    "required command " + command,
			Passed:  false,
			Details: truncateDetails(details),
		}
	}
	return Check{Name: "required command " + command, Passed: true}
}

func splitCommand(command string) ([]string, error) {
	fields := []string{}
	var builder strings.Builder
	var quote rune
	escaped := false
	for _, r := range command {
		if escaped {
			builder.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			builder.WriteRune(r)
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			if builder.Len() > 0 {
				fields = append(fields, builder.String())
				builder.Reset()
			}
			continue
		}
		builder.WriteRune(r)
	}
	if escaped {
		builder.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote in required command")
	}
	if builder.Len() > 0 {
		fields = append(fields, builder.String())
	}
	return fields, nil
}

func truncateDetails(details string) string {
	details = strings.TrimSpace(details)
	if len(details) <= 500 {
		return details
	}
	return details[:500] + "..."
}

func readListSection(content, section string) []string {
	body := readSection(content, section)
	values := []string{}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "- "))
		if value != "" && value != "None" {
			values = append(values, value)
		}
	}
	return values
}

func readSection(content, section string) string {
	re := regexp.MustCompile(`(?ms)^## ` + regexp.QuoteMeta(section) + `\s*(.*?)\n## `)
	if matches := re.FindStringSubmatch(content + "\n## END\n"); len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
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
