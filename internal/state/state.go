package state

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/checker"
	"github.com/nextvibe/nextvibe/internal/detector"
	"github.com/nextvibe/nextvibe/internal/planner"
	"github.com/nextvibe/nextvibe/internal/protocol"
	"github.com/nextvibe/nextvibe/internal/taskgen"
)

type Snapshot struct {
	ProtocolVersion string              `json:"protocolVersion"`
	ProjectName     string              `json:"projectName,omitempty"`
	Scan            *detector.Result    `json:"scan,omitempty"`
	Suggestion      *planner.Suggestion `json:"suggestion,omitempty"`
	CurrentTask     *taskgen.Task       `json:"currentTask,omitempty"`
	Check           *checker.Result     `json:"check,omitempty"`
}

func SaveScan(root string, scan detector.Result) error {
	return update(root, func(snapshot *Snapshot) {
		snapshot.ProjectName = scan.ProjectName
		snapshot.Scan = &scan
	})
}

func SaveSuggestion(root string, suggestion planner.Suggestion) error {
	return update(root, func(snapshot *Snapshot) {
		snapshot.Suggestion = &suggestion
	})
}

func SaveTask(root string, task taskgen.Task) error {
	return update(root, func(snapshot *Snapshot) {
		snapshot.CurrentTask = &task
	})
}

func SaveCheck(root string, check checker.Result) error {
	return update(root, func(snapshot *Snapshot) {
		snapshot.Check = &check
		if snapshot.CurrentTask != nil && snapshot.CurrentTask.TaskID == check.TaskID {
			snapshot.CurrentTask.Status = check.Status
		}
	})
}

func update(root string, apply func(*Snapshot)) error {
	path := filepath.Join(root, brand.WorkspaceDir, "state.json")
	snapshot := Snapshot{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &snapshot)
	} else if !os.IsNotExist(err) {
		return err
	}

	snapshot.ProtocolVersion = protocol.Version
	apply(&snapshot)

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
