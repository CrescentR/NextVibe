package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/checker"
	"github.com/nextvibe/nextvibe/internal/detector"
	"github.com/nextvibe/nextvibe/internal/installer"
	"github.com/nextvibe/nextvibe/internal/output"
	"github.com/nextvibe/nextvibe/internal/planner"
	"github.com/nextvibe/nextvibe/internal/scanner"
	"github.com/nextvibe/nextvibe/internal/taskgen"
	"github.com/nextvibe/nextvibe/internal/workspace"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeUsage(stdout)
		return 0
	}

	command := args[0]
	if command == "-h" || command == "--help" || command == "help" {
		writeUsage(stdout)
		return 0
	}

	root, err := os.Getwd()
	if err != nil {
		return fail(stderr, err)
	}

	switch command {
	case "init":
		return runInit(root, args[1:], stdout, stderr)
	case "scan":
		return runScan(root, args[1:], stdout, stderr)
	case "suggest":
		return runSuggest(root, args[1:], stdout, stderr)
	case "task":
		return runTask(root, args[1:], stdout, stderr)
	case "check":
		return runCheck(root, args[1:], stdout, stderr)
	case "install":
		return runInstall(root, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n\n", command)
		writeUsage(stderr)
		return 1
	}
}

func runInit(root string, args []string, stdout, stderr io.Writer) int {
	jsonOut, err := parseJSONFlag("init", args)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := workspace.Ensure(root)
	if err != nil {
		return fail(stderr, err)
	}
	if jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeInit(stdout, result)
	return 0
}

func runScan(root string, args []string, stdout, stderr io.Writer) int {
	jsonOut, err := parseJSONFlag("scan", args)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := scanProject(root)
	if err != nil {
		return fail(stderr, err)
	}
	if jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeScan(stdout, result)
	return 0
}

func runSuggest(root string, args []string, stdout, stderr io.Writer) int {
	jsonOut, err := parseJSONFlag("suggest", args)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := suggestProject(root)
	if err != nil {
		return fail(stderr, err)
	}
	if jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeSuggestion(stdout, result)
	return 0
}

func runTask(root string, args []string, stdout, stderr io.Writer) int {
	jsonOut, err := parseJSONFlag("task", args)
	if err != nil {
		return fail(stderr, err)
	}
	suggestion, err := suggestProject(root)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := taskgen.CurrentOrCreate(root, suggestion)
	if err != nil {
		return fail(stderr, err)
	}
	if jsonOut {
		return writeJSON(stdout, stderr, result.Task)
	}
	writeTask(stdout, result.Task, result.Created)
	return 0
}

func runCheck(root string, args []string, stdout, stderr io.Writer) int {
	jsonOut, err := parseJSONFlag("check", args)
	if err != nil {
		return fail(stderr, err)
	}
	result := checker.CheckCurrent(root)
	if jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeCheck(stdout, result)
	return 0
}

func runInstall(root string, args []string, stdout, stderr io.Writer) int {
	jsonOut, target, err := parseInstallArgs(args)
	if err != nil {
		return fail(stderr, err)
	}
	if target == "" {
		return fail(stderr, fmt.Errorf("install target required: codex, claude, cursor, or all"))
	}
	result, err := installer.Install(root, target)
	if err != nil {
		return fail(stderr, err)
	}
	if jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeInstall(stdout, result)
	return 0
}

func parseInstallArgs(args []string) (bool, string, error) {
	jsonOut := false
	target := ""
	for _, arg := range args {
		switch {
		case arg == "--json":
			jsonOut = true
		case strings.HasPrefix(arg, "-"):
			return false, "", fmt.Errorf("unknown install flag %q", arg)
		case target == "":
			target = arg
		default:
			return false, "", fmt.Errorf("install accepts one target, got %q and %q", target, arg)
		}
	}
	return jsonOut, target, nil
}

func scanProject(root string) (detector.Result, error) {
	inventory, err := scanner.Scan(root)
	if err != nil {
		return detector.Result{}, err
	}
	return detector.Detect(inventory), nil
}

func suggestProject(root string) (planner.Suggestion, error) {
	scan, err := scanProject(root)
	if err != nil {
		return planner.Suggestion{}, err
	}
	return planner.Suggest(scan), nil
}

func parseJSONFlag(command string, args []string) (bool, error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut := fs.Bool("json", false, "write machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return false, err
	}
	if fs.NArg() != 0 {
		return false, fmt.Errorf("%s does not accept positional arguments", command)
	}
	return *jsonOut, nil
}

func writeJSON(stdout, stderr io.Writer, value any) int {
	if err := output.JSON(stdout, value); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func fail(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "error: %v\n", err)
	return 1
}

func writeUsage(w io.Writer) {
	fmt.Fprintf(w, `%s

Agent-native project navigation for AI coding tools.

Usage:
  %s init [--json]
  %s scan [--json]
  %s suggest [--json]
  %s task [--json]
  %s check [--json]
  %s install <codex|claude|cursor|all> [--json]

`, brand.ProjectName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName)
}

func writeInit(w io.Writer, result workspace.InitResult) {
	fmt.Fprintln(w, result.Message)
	writeStringList(w, "Created", result.Created)
	writeStringList(w, "Existing", result.Existing)
}

func writeScan(w io.Writer, result detector.Result) {
	fmt.Fprintf(w, "%s scan\n", brand.ProjectName)
	fmt.Fprintf(w, "Project: %s\n", result.ProjectName)
	fmt.Fprintf(w, "Stage: %s (%s)\n", result.Stage.Label, result.Stage.ID)
	writeStringList(w, "Stacks", result.DetectedStacks)
	writeStringList(w, "Key files", result.KeyFiles)
	writeStringList(w, "Key directories", result.KeyDirectories)
	writeStringList(w, "Risks", result.Risks)
}

func writeSuggestion(w io.Writer, suggestion planner.Suggestion) {
	fmt.Fprintf(w, "%s suggestion\n", brand.ProjectName)
	fmt.Fprintf(w, "Current stage: %s\n", suggestion.CurrentStage)
	fmt.Fprintf(w, "Next: %s [%s]\n", suggestion.Recommendation.Title, suggestion.Recommendation.Priority)
	fmt.Fprintf(w, "Reason: %s\n", suggestion.Recommendation.Reason)
	writeStringList(w, "Artifacts", suggestion.Artifacts)
	writeStringList(w, "Acceptance criteria", suggestion.AcceptanceCriteria)
	writeStringList(w, "Allowed actions", suggestion.AgentInstructions.AllowedActions)
	writeStringList(w, "Forbidden actions", suggestion.AgentInstructions.ForbiddenActions)
}

func writeTask(w io.Writer, task taskgen.Task, created bool) {
	state := "current"
	if created {
		state = "created"
	}
	fmt.Fprintf(w, "%s task %s\n", brand.ProjectName, state)
	fmt.Fprintf(w, "Task: %s - %s\n", task.TaskID, task.Title)
	fmt.Fprintf(w, "Status: %s\n", task.Status)
	fmt.Fprintf(w, "File: %s\n", task.TaskFile)
	fmt.Fprintf(w, "Goal: %s\n", task.Goal)
	writeStringList(w, "Allowed files", task.AllowedFiles)
	writeStringList(w, "Forbidden changes", task.ForbiddenChanges)
	writeStringList(w, "Acceptance criteria", task.AcceptanceCriteria)
}

func writeCheck(w io.Writer, result checker.Result) {
	fmt.Fprintf(w, "%s check\n", brand.ProjectName)
	fmt.Fprintf(w, "Task: %s\n", emptyAs(result.TaskID, "(none)"))
	fmt.Fprintf(w, "Status: %s\n", result.Status)
	for _, check := range result.Checks {
		marker := "FAIL"
		if check.Passed {
			marker = "PASS"
		}
		if check.Details != "" {
			fmt.Fprintf(w, "- %s %s: %s\n", marker, check.Name, check.Details)
			continue
		}
		fmt.Fprintf(w, "- %s %s\n", marker, check.Name)
	}
	fmt.Fprintf(w, "Next action: %s\n", result.NextAction)
}

func writeInstall(w io.Writer, result installer.Result) {
	fmt.Fprintln(w, result.Message)
	for _, file := range result.Files {
		fmt.Fprintf(w, "- %s %s\n", file.Action, file.Path)
	}
}

func writeStringList(w io.Writer, title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(w, "%s:\n", title)
	for _, value := range values {
		fmt.Fprintf(w, "- %s\n", value)
	}
}

func emptyAs(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
