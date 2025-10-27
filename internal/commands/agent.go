package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/utils"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent-friendly commands for programmatic project interaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		action, _ := cmd.Flags().GetString("action")
		projectName, _ := cmd.Flags().GetString("project")
		outputFormat, _ := cmd.Flags().GetString("format")
		ideName, _ := cmd.Flags().GetString("ide")

		// Get values from environment if not provided
		if action == "" {
			action = os.Getenv("PMEM_ACTION")
		}
		if projectName == "" {
			projectName = os.Getenv("PMEM_SELECTED_PROJECT")
		}

		if action == "" {
			return fmt.Errorf("action required (use --action or PMEM_ACTION)")
		}
		if projectName == "" {
			return fmt.Errorf("project name required (use --project or PMEM_SELECTED_PROJECT)")
		}

		projectRepo := repository.NewProjectRepository(db.Conn())
		
		// Find project
		projects, err := projectRepo.Search(projectName)
		if err != nil {
			return fmt.Errorf("failed to search projects: %w", err)
		}

		if len(projects) == 0 {
			return fmt.Errorf("project not found: %s", projectName)
		}

		project := projects[0]

		switch strings.ToLower(action) {
		case "open":
			return openProjectInIDE(project.Path, ideName)
		case "context":
			return generateAgentContext(&project, outputFormat)
		case "handoff":
			return generateHandoffDocument(&project)
		case "insights":
			return showProjectInsights(&project)
		case "status":
			fmt.Printf("%s|%s|%d\n", project.Name, project.Status, project.Progress)
			return nil
		case "path":
			fmt.Printf("%s\n", project.Path)
			return nil
		default:
			return fmt.Errorf("unknown action: %s (available: open, context, handoff, insights, status, path)", action)
		}
	},
}

func init() {
	agentCmd.Flags().String("action", "", "Action to perform (open, context, handoff, insights, status, path)")
	agentCmd.Flags().String("project", "", "Project name")
	agentCmd.Flags().String("format", "json", "Output format for context (json, markdown, both)")
	agentCmd.Flags().String("ide", "", "IDE to use for opening (vscode, windsurf, vim, nano)")
	rootCmd.AddCommand(agentCmd)
}

func generateAgentContext(project *models.Project, format string) error {
	generator := utils.NewContextGenerator(project.Path)
	context, err := generator.GenerateAgentContext(
		project.Name,
		string(project.Status),
		project.Progress,
		project.Notes,
	)
	if err != nil {
		return fmt.Errorf("failed to generate context: %w", err)
	}

	switch strings.ToLower(format) {
	case "json":
		jsonData, err := context.ExportToJSON()
		if err != nil {
			return fmt.Errorf("failed to export JSON: %w", err)
		}
		fmt.Print(string(jsonData))
	case "markdown":
		mdData := context.ExportToMarkdown()
		fmt.Print(mdData)
	default:
		return fmt.Errorf("unknown format: %s (available: json, markdown)", format)
	}

	return nil
}
