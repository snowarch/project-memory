package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type IDE struct {
	Name        string
	Command     string
	Args        []string
	Priority    int
	Graphic     bool
	Description string
}

// IDEDetector finds available IDEs on the system
type IDEDetector struct {
	availableIDEs []IDE
}

func NewIDEDetector() *IDEDetector {
	detector := &IDEDetector{}
	detector.scanForIDEs()
	return detector
}

func (d *IDEDetector) scanForIDEs() {
	ideList := []IDE{
		{
			Name:        "VSCode",
			Command:     "code",
			Args:        []string{},
			Priority:    1,
			Graphic:     true,
			Description: "Visual Studio Code",
		},
		{
			Name:        "Windsurf",
			Command:     "windsurf",
			Args:        []string{},
			Priority:    2,
			Graphic:     true,
			Description: "Windsurf IDE",
		},
		{
			Name:        "Vim",
			Command:     "vim",
			Args:        []string{},
			Priority:    3,
			Graphic:     false,
			Description: "Vim editor",
		},
		{
			Name:        "Nano",
			Command:     "nano",
			Args:        []string{},
			Priority:    4,
			Graphic:     false,
			Description: "Nano editor",
		},
	}

	for _, ide := range ideList {
		if d.isIDEAvailable(ide.Command) {
			d.availableIDEs = append(d.availableIDEs, ide)
		}
	}
}

func (d *IDEDetector) isIDEAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func (d *IDEDetector) GetAvailableIDEs() []IDE {
	return d.availableIDEs
}

func (d *IDEDetector) GetPreferredIDE() *IDE {
	if len(d.availableIDEs) == 0 {
		return nil
	}
	
	// Return IDE with highest priority (lowest number)
	preferred := d.availableIDEs[0]
	for _, ide := range d.availableIDEs {
		if ide.Priority < preferred.Priority {
			preferred = ide
		}
	}
	return &preferred
}

func (d *IDEDetector) GetIDEByName(name string) *IDE {
	name = strings.ToLower(name)
	for _, ide := range d.availableIDEs {
		if strings.ToLower(ide.Name) == name {
			return &ide
		}
	}
	return nil
}

// LaunchProject opens a project in the specified IDE
func (d *IDEDetector) LaunchProject(projectPath string, ideName string) error {
	var ide *IDE
	
	if ideName != "" {
		ide = d.GetIDEByName(ideName)
		if ide == nil {
			return fmt.Errorf("IDE '%s' not found", ideName)
		}
	} else {
		ide = d.GetPreferredIDE()
		if ide == nil {
			return fmt.Errorf("no IDEs available")
		}
	}

	// Ensure project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", projectPath)
	}

	// Prepare command
	cmd := exec.Command(ide.Command, append(ide.Args, projectPath)...)
	
	// For GUI IDEs, we should run them in background
	if ide.Graphic {
		cmd.Start()
		return nil
	}

	// For terminal IDEs, we need to handle differently
	return cmd.Run()
}

// LaunchProjectInBackground opens project without blocking
func (d *IDEDetector) LaunchProjectInBackground(projectPath string, ideName string) error {
	var ide *IDE
	
	if ideName != "" {
		ide = d.GetIDEByName(ideName)
		if ide == nil {
			return fmt.Errorf("IDE '%s' not found", ideName)
		}
	} else {
		ide = d.GetPreferredIDE()
		if ide == nil {
			return fmt.Errorf("no IDEs available")
		}
	}

	// Ensure project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", projectPath)
	}

	// Prepare command
	cmd := exec.Command(ide.Command, append(ide.Args, projectPath)...)
	
	// Set up process group for better control
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	
	// Start in background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start IDE: %w", err)
	}

	// Don't wait for the command to complete
	return nil
}

// DetectProjectType determines the best way to open a project
func DetectProjectType(projectPath string) string {
	// Check for common project files
	files := []string{
		"package.json",    // Node.js
		"go.mod",          // Go
		"Cargo.toml",      // Rust
		"requirements.txt", // Python
		"pom.xml",         // Java Maven
		"build.gradle",    // Java Gradle
		"*.sln",           // .NET
		"CMakeLists.txt",  // C/C++
	}

	for _, file := range files {
		fullPath := filepath.Join(projectPath, file)
		if _, err := os.Stat(fullPath); err == nil {
			return file
		}
	}

	return "unknown"
}

// GetProjectOpenCommand returns the best command to open a project
func GetProjectOpenCommand(projectPath string, ideName string) (string, []string, error) {
	detector := NewIDEDetector()
	
	var ide *IDE
	if ideName != "" {
		ide = detector.GetIDEByName(ideName)
		if ide == nil {
			return "", nil, fmt.Errorf("IDE '%s' not found", ideName)
		}
	} else {
		ide = detector.GetPreferredIDE()
		if ide == nil {
			return "", nil, fmt.Errorf("no IDEs available")
		}
	}

	args := append(ide.Args, projectPath)
	return ide.Command, args, nil
}
