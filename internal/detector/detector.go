package detector

import (
	"sort"

	"github.com/nextvibe/nextvibe/internal/protocol"
	"github.com/nextvibe/nextvibe/internal/scanner"
)

type Signals struct {
	HasFrontend          bool `json:"hasFrontend"`
	HasBackend           bool `json:"hasBackend"`
	HasMockData          bool `json:"hasMockData"`
	HasApiContract       bool `json:"hasApiContract"`
	HasDatabaseSchema    bool `json:"hasDatabaseSchema"`
	HasTests             bool `json:"hasTests"`
	HasDockerfile        bool `json:"hasDockerfile"`
	HasAgentInstructions bool `json:"hasAgentInstructions"`
}

type Stage struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Result struct {
	ProtocolVersion string   `json:"protocolVersion"`
	ProjectName     string   `json:"projectName"`
	DetectedStacks  []string `json:"detectedStacks"`
	KeyFiles        []string `json:"keyFiles"`
	KeyDirectories  []string `json:"keyDirectories"`
	TestCommands    []string `json:"testCommands,omitempty"`
	Signals         Signals  `json:"signals"`
	Stage           Stage    `json:"stage"`
	Risks           []string `json:"risks"`
}

func Detect(inventory scanner.Inventory) Result {
	signals := detectSignals(inventory)
	stacks := detectStacks(inventory)
	testCommands := detectTestCommands(inventory)
	hasReadme := inventory.Files["README.md"]
	stage := detectStage(signals, hasReadme)
	risks := detectRisks(inventory, signals, hasReadme)

	return Result{
		ProtocolVersion: protocol.Version,
		ProjectName:     inventory.ProjectName,
		DetectedStacks:  stacks,
		KeyFiles:        inventory.KeyFiles,
		KeyDirectories:  inventory.KeyDirectories,
		TestCommands:    testCommands,
		Signals:         signals,
		Stage:           stage,
		Risks:           risks,
	}
}

func detectSignals(inventory scanner.Inventory) Signals {
	hasFrontend := inventory.Directories["app"] ||
		inventory.Directories["pages"] ||
		inventory.Directories["components"] ||
		hasFrontendDependency(inventory.Package)

	hasBackend := inventory.Directories["api"] ||
		inventory.Directories["server"] ||
		inventory.Directories["backend"] ||
		inventory.Directories["internal"] ||
		inventory.Directories["cmd"] ||
		inventory.GoModule != ""

	hasApiContract := inventory.Directories["docs/api"] ||
		inventory.Files["docs/api/openapi.yaml"] ||
		inventory.Files["docs/api/openapi.yml"] ||
		inventory.Files["docs/api/api-design.md"] ||
		inventory.Files["openapi.yaml"] ||
		inventory.Files["openapi.yml"] ||
		inventory.Files["swagger.yaml"] ||
		inventory.Files["swagger.yml"]

	return Signals{
		HasFrontend:          hasFrontend,
		HasBackend:           hasBackend,
		HasMockData:          inventory.Directories["mock"] || inventory.Directories["mocks"],
		HasApiContract:       hasApiContract,
		HasDatabaseSchema:    inventory.Directories["migrations"] || inventory.Directories["sql"],
		HasTests:             inventory.Directories["test"] || inventory.Directories["tests"] || inventory.Directories["__tests__"] || len(inventory.TestFiles) > 0,
		HasDockerfile:        inventory.Files["Dockerfile"] || inventory.Files["docker-compose.yml"],
		HasAgentInstructions: inventory.Files["AGENTS.md"] || inventory.Files["CLAUDE.md"] || inventory.Directories[".cursor/rules"] || inventory.Directories[".claude"],
	}
}

func detectStacks(inventory scanner.Inventory) []string {
	stacks := map[string]bool{}
	if inventory.GoModule != "" || inventory.Files["go.mod"] {
		stacks["go"] = true
	}
	if inventory.Files["package.json"] {
		stacks["node"] = true
	}
	if inventory.Files["pom.xml"] {
		stacks["java"] = true
		stacks["maven"] = true
	}
	if inventory.Files["build.gradle"] {
		stacks["java"] = true
		stacks["gradle"] = true
	}
	if inventory.Files["Dockerfile"] || inventory.Files["docker-compose.yml"] {
		stacks["docker"] = true
	}

	addPackageStacks(stacks, inventory.Package)

	result := make([]string, 0, len(stacks))
	for stack := range stacks {
		result = append(result, stack)
	}
	sort.Strings(result)
	return result
}

func addPackageStacks(stacks map[string]bool, pkg scanner.PackageInfo) {
	deps := map[string]bool{}
	for name := range pkg.Dependencies {
		deps[name] = true
	}
	for name := range pkg.DevDependencies {
		deps[name] = true
	}

	switch {
	case deps["next"]:
		stacks["nextjs"] = true
		stacks["react"] = true
	case deps["react"]:
		stacks["react"] = true
	}
	if deps["vue"] {
		stacks["vue"] = true
	}
	if deps["svelte"] || deps["@sveltejs/kit"] {
		stacks["svelte"] = true
	}
	if deps["vite"] {
		stacks["vite"] = true
	}
	if deps["typescript"] {
		stacks["typescript"] = true
	}
}

func hasFrontendDependency(pkg scanner.PackageInfo) bool {
	deps := map[string]bool{}
	for name := range pkg.Dependencies {
		deps[name] = true
	}
	for name := range pkg.DevDependencies {
		deps[name] = true
	}
	return deps["react"] || deps["next"] || deps["vue"] || deps["svelte"] || deps["@sveltejs/kit"] || deps["vite"]
}

func detectTestCommands(inventory scanner.Inventory) []string {
	commands := []string{}
	seen := map[string]bool{}

	add := func(command string) {
		if command == "" || seen[command] {
			return
		}
		seen[command] = true
		commands = append(commands, command)
	}

	if inventory.GoModule != "" || inventory.Files["go.mod"] {
		add("go test ./...")
	}
	if inventory.Package.Scripts["test"] != "" {
		add("npm test")
	}
	if inventory.Files["pom.xml"] {
		add("mvn test")
	}
	if inventory.Files["build.gradle"] {
		add("gradle test")
	}

	return commands
}

func detectStage(signals Signals, hasReadme bool) Stage {
	switch {
	case !hasReadme:
		return Stage{ID: "project-goal-missing", Label: "Project goal is missing"}
	case !signals.HasFrontend && !signals.HasBackend:
		return Stage{ID: "project-structure-missing", Label: "Implementation structure needs definition"}
	case signals.HasFrontend && !signals.HasApiContract:
		return Stage{ID: "frontend-prototype-backend-incomplete", Label: "Frontend prototype exists, backend contract missing"}
	case signals.HasApiContract && !signals.HasDatabaseSchema:
		return Stage{ID: "api-contract-data-model-missing", Label: "API contract exists, data model missing"}
	case signals.HasBackend && !signals.HasTests:
		return Stage{ID: "backend-tests-missing", Label: "Backend code exists, tests missing"}
	case !signals.HasDockerfile:
		return Stage{ID: "deployment-config-missing", Label: "Deployment configuration missing"}
	case !signals.HasAgentInstructions:
		return Stage{ID: "agent-integration-missing", Label: "Agent integration files missing"}
	default:
		return Stage{ID: "ready-for-focused-development", Label: "Project has core navigation signals"}
	}
}

func detectRisks(inventory scanner.Inventory, signals Signals, hasReadme bool) []string {
	risks := []string{}
	if !hasReadme {
		risks = append(risks, "README or project goal missing")
	}
	if signals.HasMockData {
		risks = append(risks, "Mock data detected")
	}
	if projectNeedsAPIContract(inventory, signals) && !signals.HasApiContract {
		risks = append(risks, "No API contract found")
	}
	if projectNeedsDatabaseSchema(inventory, signals) && !signals.HasDatabaseSchema {
		risks = append(risks, "No database schema found")
	}
	if !signals.HasTests {
		risks = append(risks, "No tests found")
	}
	if !signals.HasDockerfile {
		risks = append(risks, "No deployment configuration found")
	}
	if !signals.HasAgentInstructions {
		risks = append(risks, "No agent integration files found")
	}
	return risks
}

func projectNeedsAPIContract(inventory scanner.Inventory, signals Signals) bool {
	return signals.HasFrontend ||
		signals.HasMockData ||
		signals.HasApiContract ||
		signals.HasDatabaseSchema ||
		inventory.Directories["api"] ||
		inventory.Directories["server"] ||
		inventory.Directories["backend"]
}

func projectNeedsDatabaseSchema(inventory scanner.Inventory, signals Signals) bool {
	return signals.HasFrontend ||
		signals.HasMockData ||
		signals.HasApiContract ||
		inventory.Directories["api"] ||
		inventory.Directories["server"] ||
		inventory.Directories["backend"]
}
