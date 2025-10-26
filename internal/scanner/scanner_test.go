package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snowarch/project-memory/internal/models"
)

func TestIsProjectDirectory(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		expected  bool
	}{
		{
			name:     "Node.js project",
			files:    []string{"package.json"},
			expected: true,
		},
		{
			name:     "Python project",
			files:    []string{"requirements.txt"},
			expected: true,
		},
		{
			name:     "Go project",
			files:    []string{"go.mod"},
			expected: true,
		},
		{
			name:     "Rust project",
			files:    []string{"Cargo.toml"},
			expected: true,
		},
		{
			name:     "Git repository",
			files:    []string{".git"},
			expected: true,
		},
		{
			name:     "Not a project",
			files:    []string{"random.txt"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear directorio temporal
			tmpDir := t.TempDir()
			
			// Crear archivos de prueba
			for _, file := range tt.files {
				path := filepath.Join(tmpDir, file)
				if file == ".git" {
					if err := os.Mkdir(path, 0755); err != nil {
						t.Fatalf("Failed to create .git dir: %v", err)
					}
				} else {
					if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
						t.Fatalf("Failed to create file %s: %v", file, err)
					}
				}
			}

			s := New(tmpDir)
			result := s.isProjectDirectory(tmpDir)
			
			if result != tt.expected {
				t.Errorf("isProjectDirectory() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectTechnologies_NodeJS(t *testing.T) {
	tmpDir := t.TempDir()
	
	packageJSON := `{
		"dependencies": {
			"react": "^18.0.0",
			"next": "13.0.0",
			"lodash": "^4.17.21"
		}
	}`
	
	err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	s := New(tmpDir)
	techs, err := s.DetectTechnologies(tmpDir)
	if err != nil {
		t.Fatalf("DetectTechnologies() failed: %v", err)
	}

	// Verificar que se detectó Node.js
	foundNodeJS := false
	foundReact := false
	foundNext := false
	
	for _, tech := range techs {
		if tech.Name == "Node.js" && tech.Type == "runtime" {
			foundNodeJS = true
		}
		if tech.Name == "react" && tech.Type == "framework" {
			foundReact = true
		}
		if tech.Name == "next" && tech.Type == "framework" {
			foundNext = true
		}
	}

	if !foundNodeJS {
		t.Error("Node.js runtime not detected")
	}
	if !foundReact {
		t.Error("React framework not detected")
	}
	if !foundNext {
		t.Error("Next.js framework not detected")
	}
}

func TestDetectTechnologies_Go(t *testing.T) {
	tmpDir := t.TempDir()
	
	goMod := `module github.com/test/project

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)
`
	
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	s := New(tmpDir)
	techs, err := s.DetectTechnologies(tmpDir)
	if err != nil {
		t.Fatalf("DetectTechnologies() failed: %v", err)
	}

	foundGo := false
	var goVersion string
	
	for _, tech := range techs {
		if tech.Name == "Go" && tech.Type == "runtime" {
			foundGo = true
			goVersion = tech.Version
		}
	}

	if !foundGo {
		t.Error("Go runtime not detected")
	}
	
	if goVersion != "1.21" {
		t.Errorf("Go version = %s, want 1.21", goVersion)
	}
}

func TestDetectTechnologies_Rust(t *testing.T) {
	tmpDir := t.TempDir()
	
	cargoToml := `[package]
name = "test-project"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = "1.0"
`
	
	err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644)
	if err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	s := New(tmpDir)
	techs, err := s.DetectTechnologies(tmpDir)
	if err != nil {
		t.Fatalf("DetectTechnologies() failed: %v", err)
	}

	foundRust := false
	var edition string
	
	for _, tech := range techs {
		if tech.Name == "Rust" && tech.Type == "runtime" {
			foundRust = true
			edition = tech.Version
		}
	}

	if !foundRust {
		t.Error("Rust runtime not detected")
	}
	
	if edition != "2021" {
		t.Errorf("Rust edition = %s, want 2021", edition)
	}
}

func TestGenerateProjectID(t *testing.T) {
	path1 := "/home/user/project1"
	path2 := "/home/user/project2"
	path1Again := "/home/user/project1"

	id1 := generateProjectID(path1)
	id2 := generateProjectID(path2)
	id1Again := generateProjectID(path1Again)

	// IDs deben ser consistentes
	if id1 != id1Again {
		t.Error("generateProjectID() should return same ID for same path")
	}

	// IDs deben ser únicos
	if id1 == id2 {
		t.Error("generateProjectID() should return different IDs for different paths")
	}

	// ID debe tener longitud correcta (16 caracteres)
	if len(id1) != 16 {
		t.Errorf("generateProjectID() ID length = %d, want 16", len(id1))
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		readme   string
		expected string
	}{
		{
			name:     "Simple description",
			readme:   "# Project\n\nThis is a test project.",
			expected: "This is a test project.",
		},
		{
			name:     "Description with empty lines",
			readme:   "# Title\n\n\nDescription here",
			expected: "Description here",
		},
		{
			name:     "Long description truncated",
			readme:   "# Title\n\n" + string(make([]byte, 300)),
			expected: string(make([]byte, 200)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			readmePath := filepath.Join(tmpDir, "README.md")
			
			err := os.WriteFile(readmePath, []byte(tt.readme), 0644)
			if err != nil {
				t.Fatalf("Failed to create README.md: %v", err)
			}

			s := New(tmpDir)
			result := s.extractDescription(tmpDir)

			if tt.name == "Long description truncated" {
				if len(result) != 200 {
					t.Errorf("extractDescription() length = %d, want 200", len(result))
				}
			} else {
				if result != tt.expected {
					t.Errorf("extractDescription() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestAnalyzeProject(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-project"
	
	// Crear un README simple
	readme := "# Test Project\n\nThis is a test."
	err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0644)
	if err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	s := New(tmpDir)
	project, err := s.analyzeProject(tmpDir, projectName)
	if err != nil {
		t.Fatalf("analyzeProject() failed: %v", err)
	}

	if project.Name != projectName {
		t.Errorf("Project name = %s, want %s", project.Name, projectName)
	}

	if project.Path != tmpDir {
		t.Errorf("Project path = %s, want %s", project.Path, tmpDir)
	}

	if project.Status != models.StatusActive {
		t.Errorf("Project status = %s, want %s", project.Status, models.StatusActive)
	}

	if project.Description != "This is a test." {
		t.Errorf("Project description = %q, want 'This is a test.'", project.Description)
	}

	if project.ID == "" {
		t.Error("Project ID should not be empty")
	}
}
