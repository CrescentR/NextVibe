package taskgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/planner"
	"github.com/nextvibe/nextvibe/internal/protocol"
	"github.com/nextvibe/nextvibe/internal/workspace"
)

type Task struct {
	ProtocolVersion    string   `json:"protocolVersion"`
	TaskID             string   `json:"taskId"`
	Title              string   `json:"title"`
	Status             string   `json:"status"`
	Background         string   `json:"background"`
	Goal               string   `json:"goal"`
	AllowedFiles       []string `json:"allowedFiles"`
	ForbiddenChanges   []string `json:"forbiddenChanges"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	TaskFile           string   `json:"taskFile,omitempty"`
}

type Result struct {
	Task    Task `json:"task"`
	Created bool `json:"created"`
}

type HistoryEntry struct {
	TaskID   string `json:"taskId"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	TaskFile string `json:"taskFile"`
}

type History struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Tasks           []HistoryEntry `json:"tasks"`
}

func CurrentOrCreate(root string, suggestion planner.Suggestion) (Result, error) {
	if _, err := workspace.Ensure(root); err != nil {
		return Result{}, err
	}

	currentPath := filepath.Join(root, brand.WorkspaceDir, "current-task.md")
	if task, ok := readCurrentTask(root, currentPath); ok {
		if !taskIsComplete(task) {
			return Result{Task: task, Created: false}, nil
		}
	}

	task := taskFromSuggestion(root, suggestion)
	if err := writeTask(root, task); err != nil {
		return Result{}, err
	}
	if err := writeCurrentTask(root, task); err != nil {
		return Result{}, err
	}
	return Result{Task: task, Created: true}, nil
}

func CompleteCurrent(root string) (Task, error) {
	task, ok := LoadCurrent(root)
	if !ok {
		return Task{}, fmt.Errorf("no current task found")
	}
	task.Status = "complete"
	if err := writeTask(root, task); err != nil {
		return Task{}, err
	}
	if err := writeCurrentTask(root, task); err != nil {
		return Task{}, err
	}
	if err := RecordHistory(root, task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func LoadCurrent(root string) (Task, bool) {
	currentPath := filepath.Join(root, brand.WorkspaceDir, "current-task.md")
	return readCurrentTask(root, currentPath)
}

func LoadHistory(root string) (History, error) {
	path := filepath.Join(root, brand.WorkspaceDir, "task-history.json")
	history := History{
		ProtocolVersion: protocol.Version,
		Tasks:           []HistoryEntry{},
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return history, nil
	}
	if err != nil {
		return History{}, err
	}
	if err := json.Unmarshal(data, &history); err != nil {
		return History{}, err
	}
	history.ProtocolVersion = protocol.Version
	if history.Tasks == nil {
		history.Tasks = []HistoryEntry{}
	}
	return history, nil
}

func RecordHistory(root string, task Task) error {
	history, err := LoadHistory(root)
	if err != nil {
		return err
	}
	entry := HistoryEntry{
		TaskID:   task.TaskID,
		Title:    task.Title,
		Status:   task.Status,
		TaskFile: task.TaskFile,
	}
	replaced := false
	for i, existing := range history.Tasks {
		if existing.TaskID == task.TaskID {
			history.Tasks[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		history.Tasks = append(history.Tasks, entry)
	}
	return writeHistory(root, history)
}

func taskIsComplete(task Task) bool {
	return strings.EqualFold(strings.TrimSpace(task.Status), "complete")
}

func taskFromSuggestion(root string, suggestion planner.Suggestion) Task {
	id := nextTaskID(root)
	slug := slugify(suggestion.Recommendation.ID)
	if slug == "" {
		slug = slugify(suggestion.Recommendation.Title)
	}
	taskFile := filepath.ToSlash(filepath.Join(brand.WorkspaceDir, "tasks", id+"-"+slug+".md"))
	return Task{
		ProtocolVersion:    protocol.Version,
		TaskID:             id,
		Title:              suggestion.Recommendation.Title,
		Status:             "active",
		Background:         suggestion.Recommendation.Reason,
		Goal:               goalFor(suggestion),
		AllowedFiles:       suggestion.Artifacts,
		ForbiddenChanges:   suggestion.AgentInstructions.ForbiddenActions,
		AcceptanceCriteria: suggestion.AcceptanceCriteria,
		TaskFile:           taskFile,
	}
}

func goalFor(suggestion planner.Suggestion) string {
	switch suggestion.Recommendation.ID {
	case "design-api-contract":
		return "Create a first API contract that can guide backend and frontend integration."
	case "design-data-model":
		return "Create a first data model that can guide persistence and API implementation."
	case "add-first-tests":
		return "Add focused tests that make the next implementation step safer."
	case "add-deployment-config":
		return "Create a minimal local deployment path that can be verified without cloud services."
	case "install-agent-integration":
		return "Install agent integration files so coding tools can call " + brand.CommandName + " automatically."
	case "write-project-goal":
		return "Make the project goal explicit enough for coding agents to navigate safely."
	case "define-project-structure":
		return "Define the first implementation structure and next focused slice."
	case "plan-next-development-slice":
		return "Research the local project state and write the next bounded development plan before implementation."
	default:
		return "Define and complete the next focused product slice."
	}
}

func nextTaskID(root string) string {
	tasksDir := filepath.Join(root, brand.WorkspaceDir, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return "001"
	}
	maxID := 0
	re := regexp.MustCompile(`^(\d{3})-`)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		id, err := strconv.Atoi(matches[1])
		if err == nil && id > maxID {
			maxID = id
		}
	}
	return fmt.Sprintf("%03d", maxID+1)
}

func writeTask(root string, task Task) error {
	absPath := filepath.Join(root, filepath.FromSlash(task.TaskFile))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absPath, []byte(renderTask(task)), 0o644)
}

func writeCurrentTask(root string, task Task) error {
	absPath := filepath.Join(root, brand.WorkspaceDir, "current-task.md")
	content := strings.Join([]string{
		"# Current Task",
		"",
		"Task ID: " + task.TaskID,
		"Status: " + task.Status,
		"Title: " + task.Title,
		"Task file: " + task.TaskFile,
		"",
		"Run:",
		"",
		"```bash",
		brand.CommandName + " task --json",
		brand.CommandName + " check --json",
		"```",
		"",
	}, "\n")
	return os.WriteFile(absPath, []byte(content), 0o644)
}

func writeHistory(root string, history History) error {
	history.ProtocolVersion = protocol.Version
	if history.Tasks == nil {
		history.Tasks = []HistoryEntry{}
	}
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := filepath.Join(root, brand.WorkspaceDir, "task-history.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func renderTask(task Task) string {
	var builder strings.Builder
	builder.WriteString("# " + task.TaskID + " - " + task.Title + "\n\n")
	builder.WriteString("Task ID: " + task.TaskID + "\n")
	builder.WriteString("Status: " + task.Status + "\n")
	builder.WriteString("Title: " + task.Title + "\n\n")
	builder.WriteString("## Background\n\n" + task.Background + "\n\n")
	builder.WriteString("## Goal\n\n" + task.Goal + "\n\n")
	writeList(&builder, "## Allowed Files", task.AllowedFiles)
	writeList(&builder, "## Forbidden Changes", task.ForbiddenChanges)
	writeList(&builder, "## Acceptance Criteria", task.AcceptanceCriteria)
	return builder.String()
}

func writeList(builder *strings.Builder, title string, values []string) {
	builder.WriteString(title + "\n\n")
	if len(values) == 0 {
		builder.WriteString("- None\n\n")
		return
	}
	for _, value := range values {
		builder.WriteString("- " + value + "\n")
	}
	builder.WriteString("\n")
}

func readCurrentTask(root, currentPath string) (Task, bool) {
	data, err := os.ReadFile(currentPath)
	if err != nil {
		return Task{}, false
	}
	content := string(data)
	if strings.Contains(content, "No active task yet.") {
		return Task{}, false
	}
	taskFile := readField(content, "Task file:")
	if taskFile == "" {
		return Task{}, false
	}
	taskPath := filepath.Join(root, filepath.FromSlash(taskFile))
	task, ok := readTaskFile(taskPath)
	if !ok {
		return Task{}, false
	}
	task.TaskFile = filepath.ToSlash(taskFile)
	return task, true
}

func readTaskFile(path string) (Task, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, false
	}
	content := string(data)
	task := Task{
		ProtocolVersion:    protocol.Version,
		TaskID:             readField(content, "Task ID:"),
		Status:             readField(content, "Status:"),
		Title:              readField(content, "Title:"),
		Background:         readSection(content, "Background"),
		Goal:               readSection(content, "Goal"),
		AllowedFiles:       readListSection(content, "Allowed Files"),
		ForbiddenChanges:   readListSection(content, "Forbidden Changes"),
		AcceptanceCriteria: readListSection(content, "Acceptance Criteria"),
	}
	return task, task.TaskID != "" && task.Title != ""
}

func readField(content, prefix string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func readSection(content, section string) string {
	re := regexp.MustCompile(`(?ms)^## ` + regexp.QuoteMeta(section) + `\s*(.*?)\n## `)
	if matches := re.FindStringSubmatch(content + "\n## END\n"); len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
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

func slugify(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteRune('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}
