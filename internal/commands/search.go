package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search projects by name, description, or path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		projectRepo := repository.NewProjectRepository(db.Conn())

		projects, err := projectRepo.Search(query)
		if err != nil {
			return fmt.Errorf("failed to search projects: %w", err)
		}

		// Filter by status if provided
		if status != "" {
			var filtered []models.Project
			for _, project := range projects {
				if string(project.Status) == status {
					filtered = append(filtered, project)
				}
			}
			projects = filtered
		}

		// Apply limit
		if limit > 0 && len(projects) > limit {
			projects = projects[:limit]
		}

		if len(projects) == 0 {
			fmt.Printf("No projects found matching: %s\n", query)
			return nil
		}

		fmt.Printf("Found %d projects matching: %s\n\n", len(projects), query)
		for i, project := range projects {
			statusIcon := getStatusIcon(string(project.Status))
			fmt.Printf("%d. %s %s | %s | Progress: %d%%\n", 
				i+1, statusIcon, project.Name, project.Status, project.Progress)
			fmt.Printf("   Path: %s\n", project.Path)
			if project.Description != "" {
				// Truncate description for display
				desc := project.Description
				if len(desc) > 80 {
					desc = desc[:77] + "..."
				}
				fmt.Printf("   Description: %s\n", desc)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	searchCmd.Flags().StringP("status", "s", "", "Filter by status (active, paused, completed, archived)")
	searchCmd.Flags().IntP("limit", "l", 0, "Limit number of results")
	rootCmd.AddCommand(searchCmd)
}

func getStatusIcon(status string) string {
	switch status {
	case "active":
		return "ACTIVE"
	case "paused":
		return "PAUSED"
	case "archived":
		return "ARCHIVED"
	case "completed":
		return "DONE"
	default:
		return "UNKNOWN"
	}
}
