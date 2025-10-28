package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/ai"
	"github.com/snowarch/project-memory/internal/logger"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/scanner"
)

var insightsCmd = &cobra.Command{
	Use:   "insights <project-name>",
	Short: "Get comprehensive project insights and recommendations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		detailed, _ := cmd.Flags().GetBool("detailed")
		timeAvailable, _ := cmd.Flags().GetString("time")

		// Get API key
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GROQ_API_KEY not set. Set environment variable to use AI insights")
		}

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

		fmt.Printf("🔍 Analyzing project: %s\n", project.Name)
		fmt.Printf("📍 Current Status: %s | Progress: %d%%\n", project.Status, project.Progress)
		fmt.Printf("🕒 Last Updated: %s\n\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))

		// Get technologies
		techs, err := techRepo.GetByProject(project.ID)
		if err != nil {
			logger.Warn("Failed to get technologies: %v", err)
		}

		var techNames []string
		for _, tech := range techs {
			if tech.Version != "" {
				techNames = append(techNames, fmt.Sprintf("%s %s", tech.Name, tech.Version))
			} else {
				techNames = append(techNames, tech.Name)
			}
		}

		// Analyze project activity
		analyzer := scanner.NewProjectStateAnalyzer(project.Path)
		activity, suggestedProgress, confidence, insights := analyzer.AnalyzeProjectActivity()

		fmt.Printf("Activity Analysis:\n")
		fmt.Printf("   Activity Level: %s\n", activity)
		fmt.Printf("   Suggested Progress: %d%%\n", suggestedProgress)
		fmt.Printf("   Confidence: %.1f%%\n", confidence*100)
		if len(insights) > 0 {
			fmt.Printf("   Insights:\n")
			for _, insight := range insights {
				fmt.Printf("     - %s\n", insight)
			}
		}
		fmt.Println()

		// Enhanced AI analysis
		groqClient := ai.NewGroqClient(apiKey)
		
		// Read README
		readme := readREADME(project.Path)
		
		// Prepare activity insights for AI
		activityText := fmt.Sprintf("Activity: %s, Confidence: %.2f, Insights: %v", 
			activity, confidence, insights)

		if detailed {
			// Enhanced analysis
			result, tokens, err := groqClient.AnalyzeProjectEnhanced(
				project.Name,
				project.Description,
				strings.Join(techNames, ", "),
				readme,
				activityText,
			)
			if err != nil {
				return fmt.Errorf("AI analysis failed: %w", err)
			}

			fmt.Printf("🤖 Enhanced AI Analysis:\n%s\n\n", result)
			fmt.Printf("💰 Tokens Used: %d\n", tokens)

			// Generate next actions if time specified
			if timeAvailable != "" {
				fmt.Println()
				nextActions, tokens2, err := groqClient.SuggestNextActions(
					project.Name,
					string(project.Status),
					project.Progress,
					insights,
					timeAvailable,
				)
				if err != nil {
					logger.Warn("Failed to generate next actions: %v", err)
				} else {
					fmt.Printf("Next Actions (%s available):\n%s\n", timeAvailable, nextActions)
					fmt.Printf("💰 Additional Tokens: %d\n", tokens2)
				}
			}
		} else {
			// Quick summary
			summary, tokens, err := groqClient.GenerateProjectSummary(
				project.Name,
				string(project.Status),
				project.Progress,
				techNames,
				project.UpdatedAt,
				project.Notes,
			)
			if err != nil {
				return fmt.Errorf("AI summary failed: %w", err)
			}

			fmt.Printf("Developer Handoff Summary:\n%s\n\n", summary)
			fmt.Printf("💰 Tokens Used: %d\n", tokens)
		}

		// Show project statistics
		fmt.Printf("\n📈 Project Statistics:\n")
		fmt.Printf("   Age: %s\n", time.Since(project.CreatedAt).Round(24*time.Hour))
		fmt.Printf("   Technologies: %d\n", len(techs))
		fmt.Printf("   Git Repository: %v\n", project.IsGitRepo)
		if project.IsGitRepo {
			fmt.Printf("   Remote: %s\n", project.GitRemote)
			fmt.Printf("   Branch: %s\n", project.GitBranch)
		}

		return nil
	},
}

func init() {
	insightsCmd.Flags().BoolP("detailed", "d", false, "Show detailed AI analysis instead of quick summary")
	insightsCmd.Flags().String("time", "", "Available time for next actions (e.g., '2 hours', '30 minutes')")
	rootCmd.AddCommand(insightsCmd)
}
