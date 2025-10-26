package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
)

var statusCmd = &cobra.Command{
	Use:   "status <project-name> [new-status]",
	Short: "View or change project status",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		projectRepo := repository.NewProjectRepository(db.Conn())
		projects, err := projectRepo.Search(projectName)
		if err != nil {
			return fmt.Errorf("failed to search projects: %w", err)
		}

		if len(projects) == 0 {
			return fmt.Errorf("project not found: %s", projectName)
		}

		project := projects[0]

		if len(args) == 1 {
			fmt.Printf("Project: %s\n", project.Name)
			fmt.Printf("Status: %s\n", project.Status)
			fmt.Printf("Progress: %d%%\n", project.Progress)
			return nil
		}

		var newStatus models.ProjectStatus
		if len(args) == 2 {
			newStatus = models.ProjectStatus(args[1])
		} else {
			statusChoices := []string{"active", "paused", "archived", "completed"}
			gumCmd := exec.Command("gum", "choose", "--header=Select new status:")
			gumCmd.Stdin = strings.NewReader(strings.Join(statusChoices, "\n"))
			
			output, err := gumCmd.Output()
			if err != nil {
				return fmt.Errorf("status selection cancelled")
			}
			
			newStatus = models.ProjectStatus(strings.TrimSpace(string(output)))
		}

		if newStatus != models.StatusActive && 
		   newStatus != models.StatusPaused && 
		   newStatus != models.StatusArchived && 
		   newStatus != models.StatusCompleted {
			return fmt.Errorf("invalid status: %s (must be: active, paused, archived, completed)", newStatus)
		}

		project.Status = newStatus
		
		if newStatus == models.StatusCompleted && project.Progress < 100 {
			project.Progress = 100
		}

		if err := projectRepo.Update(&project); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		fmt.Printf("Status updated: %s → %s\n", project.Name, newStatus)
		return nil
	},
}
