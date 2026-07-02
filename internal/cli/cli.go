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
	"github.com/nextvibe/nextvibe/internal/state"
	"github.com/nextvibe/nextvibe/internal/taskgen"
	"github.com/nextvibe/nextvibe/internal/workspace"
)

type language string

const (
	languageEnglish language = "en"
	languageChinese language = "zh"
)

type commandOptions struct {
	jsonOut bool
	lang    language
}

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeUsage(stdout, languageEnglish)
		return 0
	}

	command := args[0]
	if command == "-h" || command == "--help" || command == "help" {
		writeUsage(stdout, languageEnglish)
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
		writeUsage(stderr, languageEnglish)
		return 1
	}
}

func runInit(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseCommandOptions("init", args)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := workspace.Ensure(root)
	if err != nil {
		return fail(stderr, err)
	}
	if options.jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeInit(stdout, result, options.lang)
	return 0
}

func runScan(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseCommandOptions("scan", args)
	if err != nil {
		return fail(stderr, err)
	}
	result, err := scanProject(root)
	if err != nil {
		return fail(stderr, err)
	}
	if err := state.SaveScan(root, result); err != nil {
		return fail(stderr, err)
	}
	if options.jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeScan(stdout, result, options.lang)
	return 0
}

func runSuggest(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseCommandOptions("suggest", args)
	if err != nil {
		return fail(stderr, err)
	}
	scan, err := scanProject(root)
	if err != nil {
		return fail(stderr, err)
	}
	if err := state.SaveScan(root, scan); err != nil {
		return fail(stderr, err)
	}
	result := planner.Suggest(scan)
	if err := state.SaveSuggestion(root, result); err != nil {
		return fail(stderr, err)
	}
	if options.jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeSuggestion(stdout, result, options.lang)
	return 0
}

func runTask(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseCommandOptions("task", args)
	if err != nil {
		return fail(stderr, err)
	}
	scan, err := scanProject(root)
	if err != nil {
		return fail(stderr, err)
	}
	if err := state.SaveScan(root, scan); err != nil {
		return fail(stderr, err)
	}
	suggestion := planner.Suggest(scan)
	if err := state.SaveSuggestion(root, suggestion); err != nil {
		return fail(stderr, err)
	}
	result, err := taskgen.CurrentOrCreate(root, suggestion)
	if err != nil {
		return fail(stderr, err)
	}
	if err := state.SaveTask(root, result.Task); err != nil {
		return fail(stderr, err)
	}
	if options.jsonOut {
		return writeJSON(stdout, stderr, result.Task)
	}
	writeTask(stdout, result.Task, result.Created, options.lang)
	return 0
}

func runCheck(root string, args []string, stdout, stderr io.Writer) int {
	options, err := parseCommandOptions("check", args)
	if err != nil {
		return fail(stderr, err)
	}
	result := checker.CheckCurrent(root)
	if err := state.SaveCheck(root, result); err != nil {
		return fail(stderr, err)
	}
	if options.jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeCheck(stdout, result, options.lang)
	return 0
}

func runInstall(root string, args []string, stdout, stderr io.Writer) int {
	options, target, err := parseInstallArgs(args)
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
	if options.jsonOut {
		return writeJSON(stdout, stderr, result)
	}
	writeInstall(stdout, result, options.lang)
	return 0
}

func parseInstallArgs(args []string) (commandOptions, string, error) {
	options := commandOptions{lang: languageEnglish}
	target := ""
	langValue := ""
	languageValue := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			options.jsonOut = true
		case arg == "--lang" || arg == "--language":
			if i+1 >= len(args) {
				return commandOptions{}, "", fmt.Errorf("%s requires a value", arg)
			}
			i++
			if arg == "--lang" {
				langValue = args[i]
			} else {
				languageValue = args[i]
			}
		case strings.HasPrefix(arg, "--lang="):
			langValue = strings.TrimPrefix(arg, "--lang=")
		case strings.HasPrefix(arg, "--language="):
			languageValue = strings.TrimPrefix(arg, "--language=")
		case strings.HasPrefix(arg, "-"):
			return commandOptions{}, "", fmt.Errorf("unknown install flag %q", arg)
		case target == "":
			target = arg
		default:
			return commandOptions{}, "", fmt.Errorf("install accepts one target, got %q and %q", target, arg)
		}
	}
	lang, err := selectLanguage(langValue, languageValue)
	if err != nil {
		return commandOptions{}, "", err
	}
	options.lang = lang
	return options, target, nil
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

func parseCommandOptions(command string, args []string) (commandOptions, error) {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOut := fs.Bool("json", false, "write machine-readable JSON")
	langValue := fs.String("lang", "", "write text output in en or zh")
	languageValue := fs.String("language", "", "alias for --lang")
	if err := fs.Parse(args); err != nil {
		return commandOptions{}, err
	}
	if fs.NArg() != 0 {
		return commandOptions{}, fmt.Errorf("%s does not accept positional arguments", command)
	}
	lang, err := selectLanguage(*langValue, *languageValue)
	if err != nil {
		return commandOptions{}, err
	}
	return commandOptions{jsonOut: *jsonOut, lang: lang}, nil
}

func selectLanguage(langValue, languageValue string) (language, error) {
	if langValue != "" && languageValue != "" {
		lang, err := parseLanguage(langValue)
		if err != nil {
			return "", err
		}
		alias, err := parseLanguage(languageValue)
		if err != nil {
			return "", err
		}
		if lang != alias {
			return "", fmt.Errorf("--lang and --language disagree: %q and %q", langValue, languageValue)
		}
		return lang, nil
	}
	if languageValue != "" {
		return parseLanguage(languageValue)
	}
	if langValue != "" {
		return parseLanguage(langValue)
	}
	return languageEnglish, nil
}

func parseLanguage(value string) (language, error) {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "_", "-")) {
	case "en", "en-us", "en-gb":
		return languageEnglish, nil
	case "zh", "zh-cn", "cn":
		return languageChinese, nil
	default:
		return "", fmt.Errorf("unsupported language %q; expected en or zh", value)
	}
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

func writeUsage(w io.Writer, lang language) {
	if lang == languageChinese {
		fmt.Fprintf(w, `%s

面向 AI 编程工具的 agent-native 项目导航。

用法:
  %s init [--json] [--lang en|zh]
  %s scan [--json] [--lang en|zh]
  %s suggest [--json] [--lang en|zh]
  %s task [--json] [--lang en|zh]
  %s check [--json] [--lang en|zh]
  %s install <codex|claude|cursor|all> [--json] [--lang en|zh]

选项:
  --lang, --language  选择文本输出语言：en 或 zh。JSON 输出保持稳定结构。

`, brand.ProjectName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName)
		return
	}
	fmt.Fprintf(w, `%s

Agent-native project navigation for AI coding tools.

Usage:
  %s init [--json] [--lang en|zh]
  %s scan [--json] [--lang en|zh]
  %s suggest [--json] [--lang en|zh]
  %s task [--json] [--lang en|zh]
  %s check [--json] [--lang en|zh]
  %s install <codex|claude|cursor|all> [--json] [--lang en|zh]

Options:
  --lang, --language  Choose text output language: en or zh. JSON output keeps its stable structure.

`, brand.ProjectName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName, brand.CommandName)
}

func writeInit(w io.Writer, result workspace.InitResult, lang language) {
	fmt.Fprintln(w, result.Message)
	writeStringList(w, lang, label(lang, "Created"), result.Created)
	writeStringList(w, lang, label(lang, "Existing"), result.Existing)
}

func writeScan(w io.Writer, result detector.Result, lang language) {
	fmt.Fprintf(w, "%s %s\n", brand.ProjectName, label(lang, "scan"))
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Project"), separator(lang), result.ProjectName)
	fmt.Fprintf(w, "%s%s %s (%s)\n", label(lang, "Stage"), separator(lang), result.Stage.Label, result.Stage.ID)
	writeStringList(w, lang, label(lang, "Stacks"), result.DetectedStacks)
	writeStringList(w, lang, label(lang, "Key files"), result.KeyFiles)
	writeStringList(w, lang, label(lang, "Key directories"), result.KeyDirectories)
	writeStringList(w, lang, label(lang, "Test commands"), result.TestCommands)
	writeStringList(w, lang, label(lang, "Risks"), result.Risks)
}

func writeSuggestion(w io.Writer, suggestion planner.Suggestion, lang language) {
	fmt.Fprintf(w, "%s %s\n", brand.ProjectName, label(lang, "suggestion"))
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Current stage"), separator(lang), suggestion.CurrentStage)
	fmt.Fprintf(w, "%s%s %s [%s]\n", label(lang, "Next"), separator(lang), suggestion.Recommendation.Title, suggestion.Recommendation.Priority)
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Reason"), separator(lang), suggestion.Recommendation.Reason)
	writeStringList(w, lang, label(lang, "Artifacts"), suggestion.Artifacts)
	writeStringList(w, lang, label(lang, "Acceptance criteria"), suggestion.AcceptanceCriteria)
	writeStringList(w, lang, label(lang, "Allowed actions"), suggestion.AgentInstructions.AllowedActions)
	writeStringList(w, lang, label(lang, "Forbidden actions"), suggestion.AgentInstructions.ForbiddenActions)
}

func writeTask(w io.Writer, task taskgen.Task, created bool, lang language) {
	state := "current"
	if created {
		state = "created"
	}
	fmt.Fprintf(w, "%s %s %s\n", brand.ProjectName, label(lang, "task"), label(lang, state))
	fmt.Fprintf(w, "%s%s %s - %s\n", label(lang, "Task"), separator(lang), task.TaskID, task.Title)
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Status"), separator(lang), task.Status)
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "File"), separator(lang), task.TaskFile)
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Goal"), separator(lang), task.Goal)
	writeStringList(w, lang, label(lang, "Allowed files"), task.AllowedFiles)
	writeStringList(w, lang, label(lang, "Forbidden changes"), task.ForbiddenChanges)
	writeStringList(w, lang, label(lang, "Acceptance criteria"), task.AcceptanceCriteria)
}

func writeCheck(w io.Writer, result checker.Result, lang language) {
	fmt.Fprintf(w, "%s %s\n", brand.ProjectName, label(lang, "check"))
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Task"), separator(lang), emptyAs(result.TaskID, label(lang, "(none)")))
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Status"), separator(lang), result.Status)
	for _, check := range result.Checks {
		marker := "FAIL"
		if check.Passed {
			marker = "PASS"
		}
		marker = label(lang, marker)
		if check.Details != "" {
			fmt.Fprintf(w, "- %s %s: %s\n", marker, check.Name, check.Details)
			continue
		}
		fmt.Fprintf(w, "- %s %s\n", marker, check.Name)
	}
	fmt.Fprintf(w, "%s%s %s\n", label(lang, "Next action"), separator(lang), result.NextAction)
}

func writeInstall(w io.Writer, result installer.Result, lang language) {
	fmt.Fprintln(w, result.Message)
	for _, file := range result.Files {
		fmt.Fprintf(w, "- %s %s\n", label(lang, file.Action), file.Path)
	}
}

func writeStringList(w io.Writer, lang language, title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(w, "%s%s\n", title, separator(lang))
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

func label(lang language, key string) string {
	if lang != languageChinese {
		return key
	}
	labels := map[string]string{
		"(none)":              "(无)",
		"Allowed actions":     "允许动作",
		"Allowed files":       "允许文件",
		"Artifacts":           "产物",
		"Created":             "已创建",
		"Existing":            "已存在",
		"Forbidden actions":   "禁止动作",
		"Forbidden changes":   "禁止变更",
		"Current stage":       "当前阶段",
		"Key directories":     "关键目录",
		"Key files":           "关键文件",
		"Next action":         "下一步",
		"Acceptance criteria": "验收标准",
		"File":                "文件",
		"Goal":                "目标",
		"Next":                "下一步",
		"PASS":                "通过",
		"FAIL":                "失败",
		"Project":             "项目",
		"Reason":              "原因",
		"Risks":               "风险",
		"Stacks":              "技术栈",
		"Stage":               "阶段",
		"Status":              "状态",
		"Task":                "任务",
		"Test commands":       "测试命令",
		"check":               "检查",
		"created":             "已创建",
		"current":             "当前",
		"scan":                "扫描",
		"suggestion":          "建议",
		"task":                "任务",
		"updated":             "已更新",
		"unchanged":           "未变化",
	}
	if value, ok := labels[key]; ok {
		return value
	}
	return key
}

func separator(lang language) string {
	if lang == languageChinese {
		return "："
	}
	return ":"
}
