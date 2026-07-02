package planner

import (
	"github.com/nextvibe/nextvibe/internal/brand"
	"github.com/nextvibe/nextvibe/internal/detector"
)

type Recommendation struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Reason   string `json:"reason"`
	Priority string `json:"priority"`
}

type AgentInstructions struct {
	AllowedActions   []string `json:"allowedActions"`
	ForbiddenActions []string `json:"forbiddenActions"`
}

type Suggestion struct {
	CurrentStage       string            `json:"currentStage"`
	Recommendation     Recommendation    `json:"recommendation"`
	Artifacts          []string          `json:"artifacts"`
	AcceptanceCriteria []string          `json:"acceptanceCriteria"`
	AgentInstructions  AgentInstructions `json:"agentInstructions"`
}

func Suggest(scan detector.Result) Suggestion {
	switch scan.Stage.ID {
	case "project-goal-missing":
		return projectGoalSuggestion(scan)
	case "project-structure-missing":
		return projectStructureSuggestion(scan)
	case "frontend-prototype-backend-incomplete":
		return apiContractSuggestion(scan)
	case "api-contract-data-model-missing":
		return dataModelSuggestion(scan)
	case "backend-tests-missing":
		return testsSuggestion(scan)
	case "deployment-config-missing":
		return deploymentSuggestion(scan)
	case "agent-integration-missing":
		return agentIntegrationSuggestion(scan)
	default:
		return focusedSliceSuggestion(scan)
	}
}

func projectStructureSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "define-project-structure",
			Title:    "Define the first implementation structure",
			Reason:   "The project has a goal, but no frontend or backend implementation structure was detected.",
			Priority: "high",
		},
		Artifacts: []string{
			brand.WorkspaceDir + "/roadmap.md",
			brand.WorkspaceDir + "/project.md",
		},
		AcceptanceCriteria: []string{
			"The first implementation slice is named",
			"The expected directory structure is documented",
			"The next task has clear acceptance criteria",
			"No unrelated product surface is added",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"read README.md",
				"update " + brand.WorkspaceDir + "/project.md",
				"update " + brand.WorkspaceDir + "/roadmap.md",
			},
			ForbiddenActions: []string{
				"do not scaffold multiple frameworks",
				"do not add deployment infrastructure",
				"do not add model API integrations",
			},
		},
	}
}

func projectGoalSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "write-project-goal",
			Title:    "Write the project goal and README",
			Reason:   "The repository does not yet expose enough project structure for an agent to navigate safely.",
			Priority: "high",
		},
		Artifacts: []string{
			"README.md",
			brand.WorkspaceDir + "/project.md",
		},
		AcceptanceCriteria: []string{
			"README.md explains what the project is for",
			"The primary user and first workflow are named",
			"Out-of-scope work is captured so agents do not expand the task",
			"No product implementation is added in this task",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"read project files",
				"create or update README.md",
				"update " + brand.WorkspaceDir + "/project.md",
			},
			ForbiddenActions: []string{
				"do not add application features",
				"do not introduce new runtime dependencies",
				"do not create deployment infrastructure",
			},
		},
	}
}

func apiContractSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "design-api-contract",
			Title:    "Design the first API contract",
			Reason:   "The project has frontend signals but no stable API contract or backend integration boundary.",
			Priority: "high",
		},
		Artifacts: []string{
			"docs/api/openapi.yaml",
			"docs/api/api-design.md",
		},
		AcceptanceCriteria: []string{
			"Each core page has a corresponding API endpoint",
			"Each endpoint includes request and response examples",
			"No business implementation is added in this task",
			"Mock data structure is used as reference when mock data exists",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"read project files",
				"create docs/api/openapi.yaml",
				"create docs/api/api-design.md",
			},
			ForbiddenActions: []string{
				"do not implement backend logic",
				"do not modify frontend components",
				"do not remove mock data yet",
			},
		},
	}
}

func dataModelSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "design-data-model",
			Title:    "Design the first data model",
			Reason:   "The project has an API contract, but no database schema or durable data model is visible.",
			Priority: "high",
		},
		Artifacts: []string{
			"docs/data-model.md",
			"migrations/001_initial_schema.sql",
		},
		AcceptanceCriteria: []string{
			"Core entities are named and related to API resources",
			"Initial schema includes primary keys and required fields",
			"Open questions are documented",
			"No application persistence code is added in this task",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"read API contract files",
				"create docs/data-model.md",
				"create migrations/001_initial_schema.sql",
			},
			ForbiddenActions: []string{
				"do not implement database access code",
				"do not rewrite API contracts unless a mismatch is documented",
				"do not modify frontend UI",
			},
		},
	}
}

func testsSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "add-first-tests",
			Title:    "Add the first focused tests",
			Reason:   "Backend or business code exists, but no test structure was detected.",
			Priority: "medium",
		},
		Artifacts: []string{
			"tests/",
		},
		AcceptanceCriteria: []string{
			"At least one meaningful test covers a core behavior",
			"The project test command is documented",
			"The test can run locally without external services",
			"No unrelated refactor is included",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"read business logic",
				"create focused tests",
				"document the test command if missing",
			},
			ForbiddenActions: []string{
				"do not rewrite production architecture",
				"do not add broad test frameworks unless needed",
				"do not change deployment configuration",
			},
		},
	}
}

func deploymentSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "add-deployment-config",
			Title:    "Add a minimal deployment configuration",
			Reason:   "The project has implementation signals but no Dockerfile or compose configuration.",
			Priority: "medium",
		},
		Artifacts: []string{
			"Dockerfile",
			".dockerignore",
			"docs/deployment.md",
		},
		AcceptanceCriteria: []string{
			"Dockerfile builds the application or CLI",
			"Unneeded local files are excluded from the image",
			"Local build and run commands are documented",
			"No cloud account or hosted service is required",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"inspect build files",
				"create Dockerfile",
				"create .dockerignore",
				"create docs/deployment.md",
			},
			ForbiddenActions: []string{
				"do not add cloud provider resources",
				"do not introduce secrets",
				"do not change application behavior",
			},
		},
	}
}

func agentIntegrationSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "install-agent-integration",
			Title:    "Install agent integration files",
			Reason:   "No Codex, Claude Code, or Cursor integration files were detected.",
			Priority: "medium",
		},
		Artifacts: []string{
			"AGENTS.md",
			"CLAUDE.md",
			".cursor/rules/nextvibe.mdc",
			".claude/skills/nextvibe/SKILL.md",
		},
		AcceptanceCriteria: []string{
			"Codex instructions explain when to run " + brand.CommandName,
			"Claude Code instructions and skill are installed",
			"Cursor rules are installed",
			"Existing user-authored instructions are preserved",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"run " + brand.CommandName + " install all",
				"inspect generated integration files",
			},
			ForbiddenActions: []string{
				"do not overwrite user-authored agent instructions",
				"do not add model API keys",
				"do not create prompt-copy workflows",
			},
		},
	}
}

func focusedSliceSuggestion(scan detector.Result) Suggestion {
	return Suggestion{
		CurrentStage: scan.Stage.Label,
		Recommendation: Recommendation{
			ID:       "choose-next-product-slice",
			Title:    "Choose the next focused product slice",
			Reason:   "The project has enough structure to proceed with a scoped implementation task.",
			Priority: "medium",
		},
		Artifacts: []string{
			brand.WorkspaceDir + "/roadmap.md",
			brand.WorkspaceDir + "/current-task.md",
		},
		AcceptanceCriteria: []string{
			"The next slice has a clear user-visible outcome",
			"Acceptance criteria are written before implementation",
			"Task boundaries identify allowed and forbidden changes",
			"The task can be verified locally",
		},
		AgentInstructions: AgentInstructions{
			AllowedActions: []string{
				"read project state",
				"update " + brand.WorkspaceDir + "/roadmap.md",
				"update " + brand.WorkspaceDir + "/current-task.md",
			},
			ForbiddenActions: []string{
				"do not expand into multiple product slices",
				"do not add unrelated dependencies",
				"do not skip acceptance criteria",
			},
		},
	}
}
