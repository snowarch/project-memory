package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/utils"
)

var contextCmd = &cobra.Command{
	Use:   "context <project-name>",
	Short: "Generate project context for code agents (no API required)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		outputFormat, _ := cmd.Flags().GetString("format")
		outputFile, _ := cmd.Flags().GetString("output")

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

		// Generate context
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

		// Output context
		switch strings.ToLower(outputFormat) {
		case "json":
			jsonData, err := context.ExportToJSON()
			if err != nil {
				return fmt.Errorf("failed to export JSON: %w", err)
			}
			
			if outputFile != "" {
				err = os.WriteFile(outputFile, jsonData, 0644)
				if err != nil {
					return fmt.Errorf("failed to write JSON file: %w", err)
				}
				fmt.Printf("📋 Context exported to: %s\n", outputFile)
			} else {
				fmt.Print(string(jsonData))
			}
			
		case "markdown":
			mdData := context.ExportToMarkdown()
			
			if outputFile != "" {
				err = os.WriteFile(outputFile, []byte(mdData), 0644)
				if err != nil {
					return fmt.Errorf("failed to write Markdown file: %w", err)
				}
				fmt.Printf("📋 Context exported to: %s\n", outputFile)
			} else {
				fmt.Print(mdData)
			}
			
		case "both":
			jsonFile := strings.TrimSuffix(outputFile, ".md") + "-context.json"
			mdFile := strings.TrimSuffix(outputFile, ".json") + "-context.md"
			
			if outputFile == "" {
				jsonFile = fmt.Sprintf("/tmp/%s-context.json", project.Name)
				mdFile = fmt.Sprintf("/tmp/%s-context.md", project.Name)
			}
			
			jsonData, err := context.ExportToJSON()
			if err != nil {
				return fmt.Errorf("failed to export JSON: %w", err)
			}
			os.WriteFile(jsonFile, jsonData, 0644)
			
			mdData := context.ExportToMarkdown()
			os.WriteFile(mdFile, []byte(mdData), 0644)
			
			fmt.Printf("📋 Context exported to:\n")
			fmt.Printf("  JSON: %s\n", jsonFile)
			fmt.Printf("  Markdown: %s\n", mdFile)
			
		default:
			return fmt.Errorf("unknown format: %s (available: json, markdown, both)", outputFormat)
		}

		return nil
	},
}

func init() {
	contextCmd.Flags().StringP("format", "f", "json", "Output format (json, markdown, both)")
	contextCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(contextCmd)
}
