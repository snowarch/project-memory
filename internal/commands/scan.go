package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/logger"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/scanner"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan directory for projects",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanPath := rootPath
		if len(args) > 0 {
			scanPath = args[0]
		}

		if scanPath == "" {
			return fmt.Errorf("no path specified. Use --path flag or provide as argument")
		}

		if _, err := os.Stat(scanPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", scanPath)
		}

		logger.Info("Scanning projects in: %s", scanPath)

		s := scanner.New(scanPath)
		projects, err := s.ScanProjects()
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		logger.Info("Found %d projects", len(projects))

		projectRepo := repository.NewProjectRepository(db.Conn())
		techRepo := repository.NewTechnologyRepository(db.Conn())

		added := 0
		updated := 0

		for _, project := range projects {
			existing, err := projectRepo.GetByPath(project.Path)
			if err != nil {
				return fmt.Errorf("failed to check existing project: %w", err)
			}

			if existing == nil {
				if err := projectRepo.Create(&project); err != nil {
					logger.Error("Failed to add project %s: %v", project.Name, err)
					continue
				}
				added++
				logger.Debug("Added new project: %s", project.Name)
			} else {
				project.ID = existing.ID
				project.Status = existing.Status
				project.Progress = existing.Progress
				project.Notes = existing.Notes
				
				if err := projectRepo.Update(&project); err != nil {
					logger.Error("Failed to update project %s: %v", project.Name, err)
					continue
				}
				updated++
				logger.Debug("Updated existing project: %s", project.Name)
			}

			techs, err := s.DetectTechnologies(project.Path)
			if err != nil {
				logger.Error("Failed to detect technologies for %s: %v", project.Name, err)
				continue
			}

			if err := techRepo.DeleteByProject(project.ID); err != nil {
				logger.Warn("Failed to clear technologies for %s: %v", project.Name, err)
			}

			for i := range techs {
				techs[i].ProjectID = project.ID
				if err := techRepo.Create(&techs[i]); err != nil {
					logger.Warn("Failed to save technology %s: %v", techs[i].Name, err)
				}
			}

			logger.Progress("  %s (%s)", project.Name, project.Path)
		}

		logger.Info("\nScan complete: %d added, %d updated", added, updated)
		return nil
	},
}
