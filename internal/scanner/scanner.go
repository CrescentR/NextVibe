package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var knownFiles = []string{
	"package.json",
	"go.mod",
	"pom.xml",
	"build.gradle",
	"Dockerfile",
	"docker-compose.yml",
	"README.md",
	"AGENTS.md",
	"CLAUDE.md",
	"openapi.yaml",
	"openapi.yml",
	"swagger.yaml",
	"swagger.yml",
	filepath.ToSlash(filepath.Join("docs", "api", "openapi.yaml")),
	filepath.ToSlash(filepath.Join("docs", "api", "openapi.yml")),
	filepath.ToSlash(filepath.Join("docs", "api", "api-design.md")),
}

var knownDirectories = []string{
	filepath.ToSlash(filepath.Join(".cursor", "rules")),
	".claude",
	"docs",
	filepath.ToSlash(filepath.Join("docs", "api")),
	"src",
	"app",
	"pages",
	"components",
	"api",
	"mock",
	"mocks",
	"server",
	"backend",
	"internal",
	"cmd",
	"test",
	"tests",
	"__tests__",
	"migrations",
	"sql",
}

type Inventory struct {
	Root           string
	ProjectName    string
	KeyFiles       []string
	KeyDirectories []string
	Files          map[string]bool
	Directories    map[string]bool
	Package        PackageInfo
	GoModule       string
	TestFiles      []string
}

type PackageInfo struct {
	Name            string
	Dependencies    map[string]string
	DevDependencies map[string]string
	Scripts         map[string]string
}

func Scan(root string) (Inventory, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Inventory{}, err
	}

	inventory := Inventory{
		Root:        absRoot,
		ProjectName: filepath.Base(absRoot),
		Files:       map[string]bool{},
		Directories: map[string]bool{},
		TestFiles:   []string{},
	}

	for _, rel := range knownFiles {
		if fileExists(absRoot, rel) {
			slashRel := filepath.ToSlash(rel)
			inventory.Files[slashRel] = true
			inventory.KeyFiles = append(inventory.KeyFiles, slashRel)
		}
	}

	for _, rel := range knownDirectories {
		if dirExists(absRoot, rel) {
			slashRel := filepath.ToSlash(rel)
			inventory.Directories[slashRel] = true
			inventory.KeyDirectories = append(inventory.KeyDirectories, slashRel)
		}
	}

	if pkg, ok := readPackageJSON(filepath.Join(absRoot, "package.json")); ok {
		inventory.Package = pkg
		if pkg.Name != "" {
			inventory.ProjectName = pkg.Name
		}
	}

	if module := readGoModule(filepath.Join(absRoot, "go.mod")); module != "" {
		inventory.GoModule = module
		if inventory.ProjectName == filepath.Base(absRoot) {
			parts := strings.Split(module, "/")
			inventory.ProjectName = parts[len(parts)-1]
		}
	}

	testFiles := findTestFiles(absRoot)
	inventory.TestFiles = testFiles
	if len(testFiles) > 0 {
		inventory.Directories["tests"] = true
	}

	sort.Strings(inventory.KeyFiles)
	sort.Strings(inventory.KeyDirectories)
	sort.Strings(inventory.TestFiles)

	return inventory, nil
}

func fileExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil && !info.IsDir()
}

func dirExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil && info.IsDir()
}

func readPackageJSON(path string) (PackageInfo, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PackageInfo{}, false
	}
	var pkg PackageInfo
	if err := json.Unmarshal(data, &pkg); err != nil {
		return PackageInfo{}, false
	}
	if pkg.Dependencies == nil {
		pkg.Dependencies = map[string]string{}
	}
	if pkg.DevDependencies == nil {
		pkg.DevDependencies = map[string]string{}
	}
	if pkg.Scripts == nil {
		pkg.Scripts = map[string]string{}
	}
	return pkg, true
}

func readGoModule(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func findTestFiles(root string) []string {
	found := []string{}
	ignored := map[string]bool{
		".git":         true,
		".nextvibe":    true,
		".idea":        true,
		"node_modules": true,
		"dist":         true,
		"build":        true,
		"vendor":       true,
		"coverage":     true,
	}

	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() {
			if ignored[name] {
				return filepath.SkipDir
			}
			return nil
		}
		if !looksLikeTestFile(name) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})

	if len(found) > 50 {
		found = found[:50]
	}
	return found
}

func looksLikeTestFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, "_test.go") ||
		strings.HasSuffix(lower, ".test.js") ||
		strings.HasSuffix(lower, ".test.jsx") ||
		strings.HasSuffix(lower, ".test.ts") ||
		strings.HasSuffix(lower, ".test.tsx") ||
		strings.HasSuffix(lower, ".spec.js") ||
		strings.HasSuffix(lower, ".spec.jsx") ||
		strings.HasSuffix(lower, ".spec.ts") ||
		strings.HasSuffix(lower, ".spec.tsx")
}
