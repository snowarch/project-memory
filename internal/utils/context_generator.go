package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AgentContext provides comprehensive project context for code agents
type AgentContext struct {
	ProjectName      string            `json:"project_name"`
	ProjectPath      string            `json:"project_path"`
	ProjectType      string            `json:"project_type"`
	Status           string            `json:"status"`
	Progress         int               `json:"progress"`
	Technologies     []TechnologyInfo  `json:"technologies"`
	RecentFiles      []FileInfo        `json:"recent_files"`
	ActivityLevel    string            `json:"activity_level"`
	LastModified     time.Time         `json:"last_modified"`
	GitInfo          *GitInfo          `json:"git_info,omitempty"`
	QuickStart       *QuickStartInfo   `json:"quick_start"`
	ImportantFiles   []string          `json:"important_files"`
	Dependencies     []DependencyInfo  `json:"dependencies"`
	BuildCommands    []string          `json:"build_commands"`
	TestCommands     []string          `json:"test_commands"`
	DevCommands      []string          `json:"dev_commands"`
	Notes            string            `json:"notes,omitempty"`
	GeneratedAt      time.Time         `json:"generated_at"`
}

type TechnologyInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version,omitempty"`
	Type         string `json:"type"`
	DetectedFrom string `json:"detected_from"`
}

type FileInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	Modified     time.Time `json:"modified"`
	Extension    string    `json:"extension"`
	Purpose      string    `json:"purpose,omitempty"`
}

type GitInfo struct {
	Remote           string   `json:"remote,omitempty"`
	Branch           string   `json:"branch,omitempty"`
	HasUncommitted   bool     `json:"has_uncommitted"`
	RecentCommits    []string `json:"recent_commits,omitempty"`
	LastCommitTime   time.Time `json:"last_commit_time,omitempty"`
}

type QuickStartInfo struct {
	SetupCommands    []string `json:"setup_commands"`
	DevCommand       string   `json:"dev_command,omitempty"`
	TestCommand      string   `json:"test_command,omitempty"`
	BuildCommand     string   `json:"build_command,omitempty"`
	InstallCommand   string   `json:"install_command,omitempty"`
}

type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type"`
}

// ContextGenerator creates comprehensive context for code agents
type ContextGenerator struct {
	projectPath string
}

func NewContextGenerator(projectPath string) *ContextGenerator {
	return &ContextGenerator{projectPath: projectPath}
}

func (cg *ContextGenerator) GenerateAgentContext(projectName, status string, progress int, notes string) (*AgentContext, error) {
	context := &AgentContext{
		ProjectName: projectName,
		ProjectPath: cg.projectPath,
		Status:      status,
		Progress:    progress,
		Notes:       notes,
		GeneratedAt: time.Now(),
	}

	// Detect project type
	context.ProjectType = cg.detectProjectType()

	// Get recent files
	recentFiles, err := cg.getRecentFiles()
	if err == nil {
		context.RecentFiles = recentFiles
	}

	// Analyze activity
	context.ActivityLevel = cg.analyzeActivityLevel()
	context.LastModified = cg.getLastModifiedTime()

	// Get Git information
	if gitInfo := cg.getGitInfo(); gitInfo != nil {
		context.GitInfo = gitInfo
	}

	// Get technologies
	techs, err := cg.getTechnologies()
	if err == nil {
		context.Technologies = techs
	}

	// Get quick start info
	context.QuickStart = cg.getQuickStartInfo()

	// Get important files
	context.ImportantFiles = cg.getImportantFiles()

	// Get dependencies
	deps, err := cg.getDependencies()
	if err == nil {
		context.Dependencies = deps
	}

	// Get commands
	context.BuildCommands = cg.getBuildCommands()
	context.TestCommands = cg.getTestCommands()
	context.DevCommands = cg.getDevCommands()

	return context, nil
}

func (cg *ContextGenerator) detectProjectType() string {
	indicators := map[string]string{
		"package.json":     "nodejs",
		"go.mod":           "go",
		"Cargo.toml":       "rust",
		"requirements.txt": "python",
		"pom.xml":          "java-maven",
		"build.gradle":     "java-gradle",
		"*.sln":            "dotnet",
		"CMakeLists.txt":   "cmake",
		"Makefile":         "makefile",
		"composer.json":    "php",
		"Gemfile":          "ruby",
	}

	for file, projType := range indicators {
		fullPath := filepath.Join(cg.projectPath, file)
		if _, err := os.Stat(fullPath); err == nil {
			return projType
		}
	}

	return "unknown"
}

func (cg *ContextGenerator) getRecentFiles() ([]FileInfo, error) {
	var files []FileInfo
	cutoff := time.Now().AddDate(0, 0, -7) // Last 7 days

	err := filepath.Walk(cg.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden files and common directories
		if strings.HasPrefix(info.Name(), ".") || 
		   strings.Contains(path, "node_modules") ||
		   strings.Contains(path, "target") ||
		   strings.Contains(path, "build") ||
		   strings.Contains(path, "dist") ||
		   strings.Contains(path, ".git") {
			return nil
		}

		if info.ModTime().After(cutoff) {
			fileInfo := FileInfo{
				Path:      path,
				Name:      info.Name(),
				Size:      info.Size(),
				Modified:  info.ModTime(),
				Extension: strings.ToLower(filepath.Ext(path)),
				Purpose:   cg.inferFilePurpose(info.Name()),
			}
			files = append(files, fileInfo)
		}

		return nil
	})

	return files, err
}

func (cg *ContextGenerator) inferFilePurpose(filename string) string {
	lower := strings.ToLower(filename)
	
	switch {
	case strings.Contains(lower, "test"):
		return "test"
	case strings.Contains(lower, "config") || strings.HasSuffix(lower, ".config.js"):
		return "configuration"
	case strings.Contains(lower, "readme"):
		return "documentation"
	case strings.Contains(lower, "license"):
		return "legal"
	case strings.HasSuffix(lower, ".md"):
		return "documentation"
	case strings.Contains(lower, "docker"):
		return "containerization"
	case strings.Contains(lower, "makefile") || strings.Contains(lower, ".mk"):
		return "build"
	case strings.Contains(lower, "package.json") || strings.Contains(lower, "go.mod"):
		return "dependencies"
	default:
		return "source"
	}
}

func (cg *ContextGenerator) analyzeActivityLevel() string {
	recentFiles, _ := cg.getRecentFiles()
	
	if len(recentFiles) > 10 {
		return "high"
	} else if len(recentFiles) > 3 {
		return "moderate"
	} else if len(recentFiles) > 0 {
		return "low"
	}
	return "stale"
}

func (cg *ContextGenerator) getLastModifiedTime() time.Time {
	var latest time.Time
	
	filepath.Walk(cg.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") || 
		   strings.Contains(path, "node_modules") ||
		   strings.Contains(path, "target") ||
		   strings.Contains(path, "build") ||
		   strings.Contains(path, "dist") ||
		   strings.Contains(path, ".git") {
			return nil
		}

		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}

		return nil
	})

	return latest
}

func (cg *ContextGenerator) getGitInfo() *GitInfo {
	gitDir := filepath.Join(cg.projectPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	gitInfo := &GitInfo{}

	// Get remote
	configPath := filepath.Join(gitDir, "config")
	if data, err := os.ReadFile(configPath); err == nil {
		content := string(data)
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "url =") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					gitInfo.Remote = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Get branch
	headPath := filepath.Join(gitDir, "HEAD")
	if headData, err := os.ReadFile(headPath); err == nil {
		headContent := strings.TrimSpace(string(headData))
		if strings.HasPrefix(headContent, "ref: refs/heads/") {
			gitInfo.Branch = strings.TrimPrefix(headContent, "ref: refs/heads/")
		}
	}

	// Check for uncommitted changes
	if output, err := exec.Command("git", "-C", cg.projectPath, "status", "--porcelain").Output(); err == nil {
		gitInfo.HasUncommitted = len(strings.TrimSpace(string(output))) > 0
	}

	return gitInfo
}

func (cg *ContextGenerator) getTechnologies() ([]TechnologyInfo, error) {
	var techs []TechnologyInfo

	// Node.js
	if data, err := os.ReadFile(filepath.Join(cg.projectPath, "package.json")); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			techs = append(techs, TechnologyInfo{
				Name:         "Node.js",
				Type:         "runtime",
				DetectedFrom: "package.json",
			})

			// Add main dependencies
			for dep, ver := range pkg.Dependencies {
				techs = append(techs, TechnologyInfo{
					Name:         dep,
					Version:      ver,
					Type:         "dependency",
					DetectedFrom: "package.json",
				})
			}
		}
	}

	// Go
	if data, err := os.ReadFile(filepath.Join(cg.projectPath, "go.mod")); err == nil {
		content := string(data)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "go ") {
				version := strings.TrimSpace(strings.TrimPrefix(line, "go "))
				techs = append(techs, TechnologyInfo{
					Name:         "Go",
					Version:      version,
					Type:         "runtime",
					DetectedFrom: "go.mod",
				})
				break
			}
		}
	}

	// Rust
	if _, err := os.Stat(filepath.Join(cg.projectPath, "Cargo.toml")); err == nil {
		techs = append(techs, TechnologyInfo{
			Name:         "Rust",
			Type:         "runtime",
			DetectedFrom: "Cargo.toml",
		})
	}

	// Python
	if _, err := os.Stat(filepath.Join(cg.projectPath, "requirements.txt")); err == nil {
		techs = append(techs, TechnologyInfo{
			Name:         "Python",
			Type:         "runtime",
			DetectedFrom: "requirements.txt",
		})
	}

	return techs, nil
}

func (cg *ContextGenerator) getQuickStartInfo() *QuickStartInfo {
	info := &QuickStartInfo{}

	switch cg.detectProjectType() {
	case "nodejs":
		info.SetupCommands = []string{"npm install"}
		info.DevCommand = "npm run dev"
		info.TestCommand = "npm test"
		info.BuildCommand = "npm run build"
	case "go":
		info.SetupCommands = []string{"go mod download"}
		info.DevCommand = "go run ."
		info.TestCommand = "go test ./..."
		info.BuildCommand = "go build"
	case "rust":
		info.SetupCommands = []string{"cargo build"}
		info.DevCommand = "cargo run"
		info.TestCommand = "cargo test"
		info.BuildCommand = "cargo build --release"
	case "python":
		info.SetupCommands = []string{"pip install -r requirements.txt"}
		info.DevCommand = "python main.py"
		info.TestCommand = "python -m pytest"
		info.BuildCommand = "python setup.py build"
	}

	return info
}

func (cg *ContextGenerator) getImportantFiles() []string {
	var files []string
	importantFiles := []string{
		"README.md", "package.json", "go.mod", "Cargo.toml", "requirements.txt",
		"Dockerfile", "docker-compose.yml", "Makefile", ".gitignore",
		".env.example", "tsconfig.json", "webpack.config.js", "vite.config.js",
	}

	for _, file := range importantFiles {
		if _, err := os.Stat(filepath.Join(cg.projectPath, file)); err == nil {
			files = append(files, file)
		}
	}

	return files
}

func (cg *ContextGenerator) getDependencies() ([]DependencyInfo, error) {
	var deps []DependencyInfo

	// This is a simplified version - in production, you'd parse actual dependency files
	switch cg.detectProjectType() {
	case "nodejs":
		if data, err := os.ReadFile(filepath.Join(cg.projectPath, "package.json")); err == nil {
			var pkg struct {
				Dependencies map[string]string `json:"dependencies"`
			}
			if json.Unmarshal(data, &pkg) == nil {
				for name, version := range pkg.Dependencies {
					deps = append(deps, DependencyInfo{
						Name:    name,
						Version: version,
						Type:    "runtime",
					})
				}
			}
		}
	}

	return deps, nil
}

func (cg *ContextGenerator) getBuildCommands() []string {
	switch cg.detectProjectType() {
	case "nodejs":
		return []string{"npm run build", "npm run compile"}
	case "go":
		return []string{"go build", "go build ./..."}
	case "rust":
		return []string{"cargo build", "cargo build --release"}
	case "python":
		return []string{"python setup.py build", "python -m build"}
	default:
		return []string{"make build"}
	}
}

func (cg *ContextGenerator) getTestCommands() []string {
	switch cg.detectProjectType() {
	case "nodejs":
		return []string{"npm test", "npm run test"}
	case "go":
		return []string{"go test", "go test ./...", "go test -v ./..."}
	case "rust":
		return []string{"cargo test", "cargo test --release"}
	case "python":
		return []string{"python -m pytest", "python -m unittest", "python test.py"}
	default:
		return []string{"make test", "pytest"}
	}
}

func (cg *ContextGenerator) getDevCommands() []string {
	switch cg.detectProjectType() {
	case "nodejs":
		return []string{"npm run dev", "npm start", "npm run serve"}
	case "go":
		return []string{"go run .", "go run main.go", "go run cmd/main.go"}
	case "rust":
		return []string{"cargo run", "cargo run --bin main"}
	case "python":
		return []string{"python main.py", "python app.py", "python -m flask run"}
	default:
		return []string{"make run", "make dev"}
	}
}

// ExportToJSON exports context as JSON for agents
func (ctx *AgentContext) ExportToJSON() ([]byte, error) {
	return json.MarshalIndent(ctx, "", "  ")
}

// ExportToMarkdown exports context as readable markdown
func (ctx *AgentContext) ExportToMarkdown() string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# Project Context: %s\n\n", ctx.ProjectName))
	md.WriteString(fmt.Sprintf("**Generated:** %s  \n", ctx.GeneratedAt.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Path:** `%s`  \n", ctx.ProjectPath))
	md.WriteString(fmt.Sprintf("**Type:** %s  \n", ctx.ProjectType))
	md.WriteString(fmt.Sprintf("**Status:** %s  \n", ctx.Status))
	md.WriteString(fmt.Sprintf("**Progress:** %d%%  \n", ctx.Progress))
	md.WriteString(fmt.Sprintf("**Activity:** %s  \n\n", ctx.ActivityLevel))

	if ctx.QuickStart != nil {
		md.WriteString("## Quick Start\n\n")
		if len(ctx.QuickStart.SetupCommands) > 0 {
			md.WriteString("**Setup:**\n```bash\n")
			for _, cmd := range ctx.QuickStart.SetupCommands {
				md.WriteString(cmd + "\n")
			}
			md.WriteString("```\n\n")
		}
		if ctx.QuickStart.DevCommand != "" {
			md.WriteString(fmt.Sprintf("**Development:** ` %s `\n\n", ctx.QuickStart.DevCommand))
		}
	}

	if len(ctx.Technologies) > 0 {
		md.WriteString("## Technologies\n\n")
		for _, tech := range ctx.Technologies {
			if tech.Version != "" {
				md.WriteString(fmt.Sprintf("- **%s** %s (%s)\n", tech.Name, tech.Version, tech.Type))
			} else {
				md.WriteString(fmt.Sprintf("- **%s** (%s)\n", tech.Name, tech.Type))
			}
		}
		md.WriteString("\n")
	}

	if len(ctx.ImportantFiles) > 0 {
		md.WriteString("## Important Files\n\n")
		for _, file := range ctx.ImportantFiles {
			md.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		md.WriteString("\n")
	}

	if ctx.GitInfo != nil {
		md.WriteString("## Git Information\n\n")
		if ctx.GitInfo.Remote != "" {
			md.WriteString(fmt.Sprintf("**Remote:** %s  \n", ctx.GitInfo.Remote))
		}
		if ctx.GitInfo.Branch != "" {
			md.WriteString(fmt.Sprintf("**Branch:** %s  \n", ctx.GitInfo.Branch))
		}
		md.WriteString(fmt.Sprintf("**Uncommitted Changes:** %v  \n\n", ctx.GitInfo.HasUncommitted))
	}

	return md.String()
}

func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
