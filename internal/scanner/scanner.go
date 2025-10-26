package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/snowarch/project-memory/internal/logger"
	"github.com/snowarch/project-memory/internal/models"
)

type Scanner struct {
	rootPath string
}

func New(rootPath string) *Scanner {
	return &Scanner{rootPath: rootPath}
}

func (s *Scanner) ScanProjects() ([]models.Project, error) {
	var projects []models.Project

	entries, err := os.ReadDir(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(s.rootPath, entry.Name())
		
		if s.isProjectDirectory(projectPath) {
			logger.Debug("Analyzing project: %s", projectPath)
			project, err := s.analyzeProject(projectPath, entry.Name())
			if err != nil {
				logger.Warn("Failed to analyze project %s: %v", entry.Name(), err)
				continue
			}
			projects = append(projects, project)
		}
	}

	return projects, nil
}

func (s *Scanner) isProjectDirectory(path string) bool {
	indicators := []string{
		"package.json",
		"requirements.txt",
		"go.mod",
		"Cargo.toml",
		"pom.xml",
		"build.gradle",
		".git",
		"README.md",
	}

	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(path, indicator)); err == nil {
			return true
		}
	}

	return false
}

func (s *Scanner) analyzeProject(projectPath, projectName string) (models.Project, error) {
	info, err := os.Stat(projectPath)
	if err != nil {
		return models.Project{}, err
	}

	projectID := generateProjectID(projectPath)
	
	description := s.extractDescription(projectPath)
	
	isGit, remote, branch := s.getGitInfo(projectPath)
	
	now := time.Now()
	
	project := models.Project{
		ID:            projectID,
		Name:          projectName,
		Path:          projectPath,
		Description:   description,
		Status:        models.StatusActive,
		Progress:      0,
		CreatedAt:     info.ModTime(),
		UpdatedAt:     now,
		LastScannedAt: &now,
		IsGitRepo:     isGit,
		GitRemote:     remote,
		GitBranch:     branch,
	}

	return project, nil
}

func (s *Scanner) extractDescription(projectPath string) string {
	readmePath := filepath.Join(projectPath, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 10 {
			if len(line) > 200 {
				return line[:200]
			}
			return line
		}
	}

	return ""
}

func (s *Scanner) getGitInfo(projectPath string) (bool, string, string) {
	gitDir := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false, "", ""
	}

	configPath := filepath.Join(gitDir, "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return true, "", ""
	}

	remote := ""
	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "url =") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				remote = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	headPath := filepath.Join(gitDir, "HEAD")
	headData, err := os.ReadFile(headPath)
	branch := ""
	if err == nil {
		headContent := strings.TrimSpace(string(headData))
		if strings.HasPrefix(headContent, "ref: refs/heads/") {
			branch = strings.TrimPrefix(headContent, "ref: refs/heads/")
		}
	}

	return true, remote, branch
}

func (s *Scanner) DetectTechnologies(projectPath string) ([]models.Technology, error) {
	var techs []models.Technology

	// Node.js detection
	packageJSON := filepath.Join(projectPath, "package.json")
	if data, err := os.ReadFile(packageJSON); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			techs = append(techs, models.Technology{
				Type:         "runtime",
				Name:         "Node.js",
				DetectedFrom: "package.json",
			})

			// Detectar frameworks populares
			frameworks := map[string]string{
				"next":         "framework",
				"react":        "framework",
				"vue":          "framework",
				"@angular/core": "framework",
				"express":      "framework",
				"nestjs":       "framework",
				"svelte":       "framework",
			}

			// Priorizar frameworks
			for dep, ver := range pkg.Dependencies {
				if fType, isFramework := frameworks[dep]; isFramework {
					techs = append(techs, models.Technology{
						Type:         fType,
						Name:         dep,
						Version:      strings.Trim(ver, "^~"),
						DetectedFrom: "package.json",
					})
				}
			}

			// Limitar dependencias regulares a las 10 más importantes
			count := 0
			for dep, ver := range pkg.Dependencies {
				if _, isFramework := frameworks[dep]; !isFramework && count < 10 {
					techs = append(techs, models.Technology{
						Type:         "dependency",
						Name:         dep,
						Version:      strings.Trim(ver, "^~"),
						DetectedFrom: "package.json",
					})
					count++
				}
			}
		}
	}

	requirementsTxt := filepath.Join(projectPath, "requirements.txt")
	if data, err := os.ReadFile(requirementsTxt); err == nil {
		techs = append(techs, models.Technology{
			Type:         "runtime",
			Name:         "Python",
			DetectedFrom: "requirements.txt",
		})

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Split(line, "==")
			name := strings.TrimSpace(parts[0])
			version := ""
			if len(parts) == 2 {
				version = strings.TrimSpace(parts[1])
			}
			techs = append(techs, models.Technology{
				Type:         "dependency",
				Name:         name,
				Version:      version,
				DetectedFrom: "requirements.txt",
			})
		}
	}

	// Go detection
	goMod := filepath.Join(projectPath, "go.mod")
	if data, err := os.ReadFile(goMod); err == nil {
		techs = append(techs, models.Technology{
			Type:         "runtime",
			Name:         "Go",
			DetectedFrom: "go.mod",
		})

		// Detectar versión de Go
		content := string(data)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "go ") {
				version := strings.TrimPrefix(strings.TrimSpace(line), "go ")
				techs[len(techs)-1].Version = version
				break
			}
		}
	}

	// Rust detection
	cargoToml := filepath.Join(projectPath, "Cargo.toml")
	if data, err := os.ReadFile(cargoToml); err == nil {
		techs = append(techs, models.Technology{
			Type:         "runtime",
			Name:         "Rust",
			DetectedFrom: "Cargo.toml",
		})

		// Buscar edición de Rust en el Cargo.toml
		content := string(data)
		if strings.Contains(content, "edition") {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "edition") && strings.Contains(line, "=") {
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						edition := strings.Trim(strings.TrimSpace(parts[1]), `"`)
						techs[len(techs)-1].Version = edition
						break
					}
				}
			}
		}
	}

	return techs, nil
}

func generateProjectID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16]
}
