package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/scanner"
)

var handoffCmd = &cobra.Command{
	Use:   "handoff <project-name> [output-file]",
	Short: "Generate comprehensive developer handoff documentation",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		outputFile := ""
		if len(args) > 1 {
			outputFile = args[1]
		}

		includeCode, _ := cmd.Flags().GetBool("include-code")

		projectRepo := repository.NewProjectRepository(db.Conn())
		techRepo := repository.NewTechnologyRepository(db.Conn())

		// Find project
		projects, err := projectRepo.Search(projectName)
		if err != nil {
			return fmt.Errorf("failed to search projects: %w", err)
		}

		if len(projects) == 0 {
			return fmt.Errorf("project not found: %s", projectName)
		}

		project := projects[0]

		// Generate handoff documentation
		handoffDoc, err := generateHandoffDoc(&project, techRepo, includeCode)
		if err != nil {
			return fmt.Errorf("failed to generate handoff: %w", err)
		}

		// Output documentation
		if outputFile != "" {
			err = os.WriteFile(outputFile, []byte(handoffDoc), 0644)
			if err != nil {
				return fmt.Errorf("failed to write handoff file: %w", err)
			}
			fmt.Printf("📄 Handoff documentation saved to: %s\n", outputFile)
		} else {
			fmt.Print(handoffDoc)
		}

		return nil
	},
}

func generateHandoffDoc(project *models.Project, techRepo *repository.TechnologyRepository, includeCode bool) (string, error) {
	var doc strings.Builder

	// Header
	doc.WriteString(fmt.Sprintf("# Developer Handoff: %s\n\n", project.Name))
	doc.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	doc.WriteString(fmt.Sprintf("**Status:** %s  \n", project.Status))
	doc.WriteString(fmt.Sprintf("**Progress:** %d%%  \n", project.Progress))
	doc.WriteString(fmt.Sprintf("**Location:** `%s`  \n\n", project.Path))

	// Project Overview
	doc.WriteString("## 📋 Project Overview\n\n")
	if project.Description != "" {
		doc.WriteString(fmt.Sprintf("%s\n\n", project.Description))
	} else {
		doc.WriteString("*No description available*\n\n")
	}

	// Current State Analysis
	doc.WriteString("## 🔄 Current State Analysis\n\n")
	analyzer := scanner.NewProjectStateAnalyzer(project.Path)
	activity, suggestedProgress, confidence, insights := analyzer.AnalyzeProjectActivity()

	doc.WriteString(fmt.Sprintf("**Activity Level:** %s  \n", activity))
	doc.WriteString(fmt.Sprintf("**Confidence:** %.1f%%  \n", confidence*100))
	doc.WriteString(fmt.Sprintf("**Suggested Progress:** %d%%  \n\n", suggestedProgress))

	if len(insights) > 0 {
		doc.WriteString("**Recent Insights:**\n")
		for _, insight := range insights {
			doc.WriteString(fmt.Sprintf("- %s\n", insight))
		}
		doc.WriteString("\n")
	}

	// Technologies
	doc.WriteString("## 🛠️ Technology Stack\n\n")
	techs, err := techRepo.GetByProject(project.ID)
	if err == nil && len(techs) > 0 {
		for _, tech := range techs {
			if tech.Version != "" {
				doc.WriteString(fmt.Sprintf("- **%s** %s (%s)\n", tech.Name, tech.Version, tech.Type))
			} else {
				doc.WriteString(fmt.Sprintf("- **%s** (%s)\n", tech.Name, tech.Type))
			}
		}
	} else {
		doc.WriteString("*No technologies detected*\n")
	}
	doc.WriteString("\n")

	// Git Information
	if project.IsGitRepo {
		doc.WriteString("## 📦 Git Information\n\n")
		if project.GitRemote != "" {
			doc.WriteString(fmt.Sprintf("**Remote:** %s  \n", project.GitRemote))
		}
		if project.GitBranch != "" {
			doc.WriteString(fmt.Sprintf("**Branch:** %s  \n", project.GitBranch))
		}
		doc.WriteString("\n")
	}

	// Quick Start Instructions
	doc.WriteString("## 🚀 Quick Start\n\n")
	doc.WriteString("```bash\n")
	doc.WriteString(fmt.Sprintf("cd %s\n", project.Path))
	
	// Detect common setup commands
	if _, err := os.Stat(filepath.Join(project.Path, "package.json")); err == nil {
		doc.WriteString("npm install\nnpm run dev\n")
	} else if _, err := os.Stat(filepath.Join(project.Path, "go.mod")); err == nil {
		doc.WriteString("go mod download\n")
		if _, err := os.Stat(filepath.Join(project.Path, "main.go")); err == nil {
			doc.WriteString("go run .\n")
		} else {
			doc.WriteString("go build ./...\n")
		}
	} else if _, err := os.Stat(filepath.Join(project.Path, "Cargo.toml")); err == nil {
		doc.WriteString("cargo build\n")
		doc.WriteString("cargo run\n")
	} else if _, err := os.Stat(filepath.Join(project.Path, "requirements.txt")); err == nil {
		doc.WriteString("pip install -r requirements.txt\n")
		doc.WriteString("python main.py\n")
	} else {
		doc.WriteString("# Check project-specific setup instructions\n")
	}
	
	doc.WriteString("```\n\n")

	// Project Structure
	doc.WriteString("## 📁 Project Structure\n\n")
	doc.WriteString("```\n")
	err = filepath.Walk(project.Path, func(path string, info os.FileInfo, err error) error {
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

		relPath, _ := filepath.Rel(project.Path, path)
		if strings.Count(relPath, string(filepath.Separator)) <= 2 { // Only show top 2 levels
			doc.WriteString(fmt.Sprintf("%s/\n", relPath))
		}

		return nil
	})
	doc.WriteString("```\n\n")

	// Important Files
	doc.WriteString("## 📄 Important Files\n\n")
	importantFiles := []string{"README.md", "package.json", "go.mod", "Cargo.toml", "requirements.txt", "Dockerfile", "docker-compose.yml"}
	
	for _, file := range importantFiles {
		if _, err := os.Stat(filepath.Join(project.Path, file)); err == nil {
			doc.WriteString(fmt.Sprintf("- `%s` - Configuration/Documentation\n", file))
		}
	}
	doc.WriteString("\n")

	// Notes and Context
	if project.Notes != "" {
		doc.WriteString("## 📝 Notes & Context\n\n")
		doc.WriteString(fmt.Sprintf("%s\n\n", project.Notes))
	}

	// Next Steps
	doc.WriteString("## 🎯 Recommended Next Steps\n\n")
	doc.WriteString("1. **Review Current State** - Check recent changes and uncommitted work\n")
	doc.WriteString("2. **Run Tests** - Ensure everything is working correctly\n")
	doc.WriteString("3. **Check Dependencies** - Update if necessary\n")
	doc.WriteString("4. **Review Issues** - Check for any blockers or known problems\n")
	doc.WriteString("5. **Setup Development Environment** - Follow quick start instructions\n\n")

	// Contact Information
	doc.WriteString("## 📞 Handoff Information\n\n")
	doc.WriteString(fmt.Sprintf("**Handoff Date:** %s  \n", time.Now().Format("2006-01-02")))
	doc.WriteString(fmt.Sprintf("**Project Age:** %s  \n", time.Since(project.CreatedAt).Round(24*time.Hour)))
	doc.WriteString(fmt.Sprintf("**Last Updated:** %s  \n", project.UpdatedAt.Format("2006-01-02 15:04:05")))
	doc.WriteString("\n---\n")
	doc.WriteString("*Generated by Project Memory Bank (pmem)*\n")

	return doc.String(), nil
}

func init() {
	handoffCmd.Flags().BoolP("include-code", "c", false, "Include code examples in handoff")
	rootCmd.AddCommand(handoffCmd)
}
