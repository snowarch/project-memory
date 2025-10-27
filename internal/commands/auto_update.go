package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/logger"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/scanner"
)

var autoUpdateCmd = &cobra.Command{
	Use:   "auto-update [directory]",
	Short: "Automatically analyze and update project states based on activity",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		directory := "/home/snowf/CascadeProjects" // Default
		if len(args) > 0 {
			directory = args[0]
		}

		forceUpdate, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		logger.Info("Starting automatic project state analysis...")
		logger.Info("Scanning directory: %s", directory)

		projectRepo := repository.NewProjectRepository(db.Conn())

		// Get existing projects
		projects, err := projectRepo.List("", 1000, 0)
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}

		updatedCount := 0
		insightsCount := 0

		for _, project := range projects {
			if project.Path != directory && !filepath.HasPrefix(project.Path, directory) {
				continue
			}

			logger.Debug("Analyzing project: %s", project.Name)

			// Analyze project activity
			analyzer := scanner.NewProjectStateAnalyzer(project.Path)
			activity, suggestedProgress, confidence, insights := analyzer.AnalyzeProjectActivity()

			// Determine if update is needed
			shouldUpdate := forceUpdate || 
				confidence > 0.7 || 
				(project.Progress != suggestedProgress && confidence > 0.5) ||
				(time.Since(project.UpdatedAt) > 24*time.Hour && len(insights) > 0)

			if shouldUpdate {
				if dryRun {
					fmt.Printf("[DRY RUN] Would update %s:\n", project.Name)
					fmt.Printf("  Activity: %s (confidence: %.2f)\n", activity, confidence)
					fmt.Printf("  Progress: %d%% → %d%%\n", project.Progress, suggestedProgress)
					if len(insights) > 0 {
						fmt.Printf("  Insights:\n")
						for _, insight := range insights {
							fmt.Printf("    • %s\n", insight)
						}
					}
					fmt.Println()
				} else {
					// Update project with new insights
					oldProgress := project.Progress
					project.Progress = suggestedProgress
					project.UpdatedAt = time.Now()
					
					// Add insights to notes
					if len(insights) > 0 {
						insightText := fmt.Sprintf("[%s] Auto-analysis: %s", 
							time.Now().Format("2006-01-02"), 
							fmt.Sprintf("Activity: %s, Insights: %v", activity, insights))
						if project.Notes != "" {
							project.Notes += "\n" + insightText
						} else {
							project.Notes = insightText
						}
					}

					if err := projectRepo.Update(&project); err != nil {
						logger.Warn("Failed to update project %s: %v", project.Name, err)
						continue
					}

					updatedCount++
					insightsCount += len(insights)

					fmt.Printf("✓ Updated %s: %s activity, %d%% progress", 
						project.Name, activity, suggestedProgress)
					if oldProgress != suggestedProgress {
						fmt.Printf(" (was %d%%)", oldProgress)
					}
					fmt.Println()
				}
			}
		}

		if dryRun {
			fmt.Printf("Dry run completed. %d projects would be updated.\n", updatedCount)
		} else {
			logger.Info("Auto-update completed: %d projects updated, %d insights generated", 
				updatedCount, insightsCount)
		}

		return nil
	},
}

func init() {
	autoUpdateCmd.Flags().BoolP("force", "f", false, "Force update even for low-confidence changes")
	autoUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	rootCmd.AddCommand(autoUpdateCmd)
}
