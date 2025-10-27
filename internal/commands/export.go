package commands

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
)

var exportCmd = &cobra.Command{
	Use:   "export <format> <output-file>",
	Short: "Export projects to JSON or CSV format",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := args[0]
		outputFile := args[1]
		status, _ := cmd.Flags().GetString("status")

		if format != "json" && format != "csv" {
			return fmt.Errorf("format must be 'json' or 'csv'")
		}

		projectRepo := repository.NewProjectRepository(db.Conn())
		techRepo := repository.NewTechnologyRepository(db.Conn())

		var projects []models.Project
		var err error

		if status != "" {
			projects, err = projectRepo.List(status, 1000, 0)
		} else {
			projects, err = projectRepo.List("", 1000, 0) // Large limit
		}

		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}

		if len(projects) == 0 {
			fmt.Printf("No projects found to export\n")
			return nil
		}

		switch format {
		case "json":
			err = exportToJSON(projects, techRepo, outputFile)
		case "csv":
			err = exportToCSV(projects, techRepo, outputFile)
		}

		if err != nil {
			return err
		}

		fmt.Printf("Exported %d projects to %s\n", len(projects), outputFile)
		return nil
	},
}

func exportToJSON(projects []models.Project, techRepo *repository.TechnologyRepository, outputFile string) error {
	type ExportProject struct {
		Name         string            `json:"name"`
		Path         string            `json:"path"`
		Description  string            `json:"description"`
		Status       string            `json:"status"`
		Progress     int               `json:"progress"`
		CreatedAt    int64             `json:"created_at"`
		UpdatedAt    int64             `json:"updated_at"`
		IsGitRepo    bool              `json:"is_git_repo"`
		GitRemote    string            `json:"git_remote,omitempty"`
		GitBranch    string            `json:"git_branch,omitempty"`
		Technologies []models.Technology `json:"technologies"`
	}

	var exportProjects []ExportProject
	for _, project := range projects {
		techs, _ := techRepo.GetByProject(project.ID)
		
		exportProject := ExportProject{
			Name:         project.Name,
			Path:         project.Path,
			Description:  project.Description,
			Status:       string(project.Status),
			Progress:     project.Progress,
			CreatedAt:    project.CreatedAt.Unix(),
			UpdatedAt:    project.UpdatedAt.Unix(),
			IsGitRepo:    project.IsGitRepo,
			GitRemote:    project.GitRemote,
			GitBranch:    project.GitBranch,
			Technologies: techs,
		}
		exportProjects = append(exportProjects, exportProject)
	}

	data, err := json.MarshalIndent(exportProjects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(outputFile, data, 0644)
}

func exportToCSV(projects []models.Project, techRepo *repository.TechnologyRepository, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Name", "Path", "Description", "Status", "Progress",
		"Created At", "Updated At", "Is Git Repo", "Git Remote", "Git Branch",
		"Technologies",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, project := range projects {
		techs, _ := techRepo.GetByProject(project.ID)
		var techNames []string
		for _, tech := range techs {
			if tech.Version != "" {
				techNames = append(techNames, fmt.Sprintf("%s %s", tech.Name, tech.Version))
			} else {
				techNames = append(techNames, tech.Name)
			}
		}

		row := []string{
			project.Name,
			project.Path,
			project.Description,
			string(project.Status),
			strconv.Itoa(project.Progress),
			project.CreatedAt.Format("2006-01-02 15:04:05"),
			project.UpdatedAt.Format("2006-01-02 15:04:05"),
			strconv.FormatBool(project.IsGitRepo),
			project.GitRemote,
			project.GitBranch,
			fmt.Sprintf("[%s]", strings.Join(techNames, ", ")),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

func init() {
	exportCmd.Flags().StringP("status", "s", "", "Filter by status (active, paused, completed, archived)")
	rootCmd.AddCommand(exportCmd)
}
