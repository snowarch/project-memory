package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/ai"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
)

var apiKey string

var analyzeCmd = &cobra.Command{
	Use:   "analyze <project-name>",
	Short: "Analyze project with AI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if apiKey == "" {
			apiKey = os.Getenv("GROQ_API_KEY")
			if apiKey == "" {
				return fmt.Errorf("GROQ_API_KEY not set. Use --api-key flag or set environment variable")
			}
		}

		projectRepo := repository.NewProjectRepository(db.Conn())
		techRepo := repository.NewTechnologyRepository(db.Conn())

		projects, err := projectRepo.Search(projectName)
		if err != nil {
			return fmt.Errorf("failed to search projects: %w", err)
		}

		if len(projects) == 0 {
			return fmt.Errorf("project not found: %s", projectName)
		}

		project := projects[0]

		fmt.Printf("Analyzing project: %s\n", project.Name)

		techs, err := techRepo.GetByProject(project.ID)
		if err != nil {
			return fmt.Errorf("failed to get technologies: %w", err)
		}

		techList := make([]string, len(techs))
		for i, tech := range techs {
			if tech.Version != "" {
				techList[i] = fmt.Sprintf("%s %s", tech.Name, tech.Version)
			} else {
				techList[i] = tech.Name
			}
		}

		readme := readREADME(project.Path)
		
		groqClient := ai.NewGroqClient(apiKey)
		
		result, tokens, err := groqClient.AnalyzeProject(
			project.Name,
			project.Description,
			strings.Join(techList, ", "),
			readme,
		)
		if err != nil {
			return fmt.Errorf("AI analysis failed: %w", err)
		}

		analysisRepo := repository.NewAnalysisRepository(db.Conn())
		analysis := &models.AIAnalysis{
			ProjectID:    project.ID,
			AnalysisType: "project_status",
			Result:       result,
			Model:        ai.Model,
			TokensUsed:   tokens,
			AnalyzedAt:   time.Now(),
		}

		if err := analysisRepo.Create(analysis); err != nil {
			fmt.Printf("Warning: Failed to save analysis: %v\n", err)
		}

		fmt.Printf("\nAnalysis Result:\n%s\n", result)
		fmt.Printf("\nTokens used: %d\n", tokens)

		return nil
	},
}

func readREADME(projectPath string) string {
	readmePath := filepath.Join(projectPath, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return "(No README.md found)"
	}

	content := string(data)
	
	// Truncate to 4000 characters to optimize tokens
	maxLen := 4000
	if len(content) > maxLen {
		return content[:maxLen] + "\n... (truncated)"
	}
	
	return content
}
