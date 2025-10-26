package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/repository"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects with gum",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectRepo := repository.NewProjectRepository(db.Conn())
		
		projects, err := projectRepo.List(statusFilter, limit, 0)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found. Run 'pmem scan' first.")
			return nil
		}

		var choices []string
		for _, p := range projects {
			statusIcon := getStatusIcon(string(p.Status))
			line := fmt.Sprintf("%s %s | %s | Progress: %d%%", statusIcon, p.Name, p.Status, p.Progress)
			choices = append(choices, line)
		}

		gumCmd := exec.Command("gum", "choose", "--header=Select a project:")
		gumCmd.Stdin = strings.NewReader(strings.Join(choices, "\n"))
		gumCmd.Stderr = os.Stderr

		output, err := gumCmd.Output()
		if err != nil {
			return nil
		}

		selected := strings.TrimSpace(string(output))
		if selected == "" {
			return nil
		}

		parts := strings.Split(selected, " | ")
		if len(parts) == 0 {
			return nil
		}

		projectName := strings.TrimSpace(strings.TrimPrefix(parts[0], getStatusIcon(statusFilter)))
		projectName = strings.TrimSpace(projectName)

		for _, p := range projects {
			if p.Name == projectName {
				return showProjectDetails(p.ID)
			}
		}

		return nil
	},
}

func getStatusIcon(status string) string {
	switch status {
	case "active":
		return "●"
	case "paused":
		return "◐"
	case "archived":
		return "○"
	case "completed":
		return "✓"
	default:
		return "•"
	}
}

func showProjectDetails(projectID string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	techRepo := repository.NewTechnologyRepository(db.Conn())

	project, err := projectRepo.GetByID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	techs, err := techRepo.GetByProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to get technologies: %w", err)
	}

	details := fmt.Sprintf(`
Project: %s
Path: %s
Status: %s
Progress: %d%%

Description:
%s

Technologies:
`, project.Name, project.Path, project.Status, project.Progress, project.Description)

	for _, tech := range techs {
		if tech.Version != "" {
			details += fmt.Sprintf("  - %s: %s (%s)\n", tech.Name, tech.Version, tech.Type)
		} else {
			details += fmt.Sprintf("  - %s (%s)\n", tech.Name, tech.Type)
		}
	}

	if project.IsGitRepo {
		details += fmt.Sprintf("\nGit:\n  Branch: %s\n", project.GitBranch)
		if project.GitRemote != "" {
			details += fmt.Sprintf("  Remote: %s\n", project.GitRemote)
		}
	}

	if project.Notes != "" {
		details += fmt.Sprintf("\nNotes:\n%s\n", project.Notes)
	}

	gumCmd := exec.Command("gum", "pager")
	gumCmd.Stdin = strings.NewReader(details)
	gumCmd.Stdout = os.Stdout
	gumCmd.Stderr = os.Stderr
	
	return gumCmd.Run()
}
