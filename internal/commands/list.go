package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/models"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/utils"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects with interactive menu",
	RunE: func(cmd *cobra.Command, args []string) error {
		nonBlocking, _ := cmd.Flags().GetBool("non-blocking")
		autoIDE, _ := cmd.Flags().GetBool("auto-ide")
		ideName, _ := cmd.Flags().GetString("ide")
		exportContext, _ := cmd.Flags().GetString("export-context")

		projectRepo := repository.NewProjectRepository(db.Conn())
		
		projects, err := projectRepo.List(statusFilter, limit, 0)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found. Run 'pmem scan' first.")
			return nil
		}

		if nonBlocking {
			return showNonBlockingMenu(projects, autoIDE, ideName, exportContext)
		}

		return showInteractiveMenu(projects, autoIDE, ideName, exportContext)
	},
}

func showInteractiveMenu(projects []models.Project, autoIDE bool, ideName string, exportContext string) error {
	var choices []string
	for _, p := range projects {
		statusIcon := getStatusIcon(string(p.Status))
		line := fmt.Sprintf("%s %s | %s | Progress: %d%%", statusIcon, p.Name, p.Status, p.Progress)
		choices = append(choices, line)
	}

	// Add special options
	choices = append(choices, "🚀 Open in IDE", "📋 Generate Context", "📄 Export Handoff")

	gumCmd := exec.Command("gum", "choose", "--header=Select a project or action:")
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

	// Handle special actions
	switch selected {
	case "🚀 Open in IDE":
		return showIDESelectionMenu(projects)
	case "📋 Generate Context":
		return showContextGenerationMenu(projects)
	case "📄 Export Handoff":
		return showHandoffMenu(projects)
	}

	// Handle project selection
	parts := strings.Split(selected, " | ")
	if len(parts) == 0 {
		return nil
	}

	projectName := strings.TrimSpace(strings.TrimPrefix(parts[0], getStatusIcon(statusFilter)))
	projectName = strings.TrimSpace(projectName)

	for _, p := range projects {
		if p.Name == projectName {
			if autoIDE {
				return openProjectInIDE(p.Path, ideName)
			}
			return showProjectActions(p)
		}
	}

	return nil
}

func showNonBlockingMenu(projects []models.Project, autoIDE bool, ideName string, exportContext string) error {
	// Create a simple text-based menu for non-interactive environments
	fmt.Println("Available Projects:")
	fmt.Println("==================")
	
	for i, p := range projects {
		statusIcon := getStatusIcon(string(p.Status))
		fmt.Printf("%d. %s %s | %s | Progress: %d%%\n", 
			i+1, statusIcon, p.Name, p.Status, p.Progress)
	}
	
	fmt.Println("\nSpecial Actions:")
	fmt.Println("99. Open project in IDE")
	fmt.Println("98. Generate context for agents")
	fmt.Println("97. Export handoff documentation")
	
	// For non-blocking mode, we can't wait for user input
	// Instead, we provide instructions for programmatic use
	fmt.Println("\nNon-blocking mode active.")
	fmt.Println("Use environment variables or flags to specify actions:")
	fmt.Println("  export PMEM_SELECTED_PROJECT='project-name'")
	fmt.Println("  export PMEM_ACTION='open|context|handoff'")
	
	return nil
}

func showIDESelectionMenu(projects []models.Project) error {
	detector := utils.NewIDEDetector()
	availableIDEs := detector.GetAvailableIDEs()
	
	if len(availableIDEs) == 0 {
		fmt.Println("No IDEs detected on system.")
		return nil
	}

	var choices []string
	for _, p := range projects {
		choices = append(choices, p.Name)
	}

	gumCmd := exec.Command("gum", "choose", "--header=Select project to open:")
	gumCmd.Stdin = strings.NewReader(strings.Join(choices, "\n"))
	gumCmd.Stderr = os.Stderr

	output, err := gumCmd.Output()
	if err != nil {
		return nil
	}

	selectedProject := strings.TrimSpace(string(output))
	if selectedProject == "" {
		return nil
	}

	// Now select IDE
	var ideChoices []string
	for _, ide := range availableIDEs {
		ideChoices = append(ideChoices, fmt.Sprintf("%s (%s)", ide.Name, ide.Description))
	}

	gumCmd = exec.Command("gum", "choose", "--header=Select IDE:")
	gumCmd.Stdin = strings.NewReader(strings.Join(ideChoices, "\n"))
	gumCmd.Stderr = os.Stderr

	ideOutput, err := gumCmd.Output()
	if err != nil {
		return nil
	}

	selectedIDE := strings.TrimSpace(string(ideOutput))
	if selectedIDE == "" {
		return nil
	}

	// Extract IDE name
	ideName := strings.Split(selectedIDE, " ")[0]

	for _, p := range projects {
		if p.Name == selectedProject {
			return openProjectInIDE(p.Path, ideName)
		}
	}

	return nil
}

func showContextGenerationMenu(projects []models.Project) error {
	var choices []string
	for _, p := range projects {
		choices = append(choices, p.Name)
	}

	gumCmd := exec.Command("gum", "choose", "--header=Select project for context generation:")
	gumCmd.Stdin = strings.NewReader(strings.Join(choices, "\n"))
	gumCmd.Stderr = os.Stderr

	output, err := gumCmd.Output()
	if err != nil {
		return nil
	}

	selectedProject := strings.TrimSpace(string(output))
	if selectedProject == "" {
		return nil
	}

	// Select format
	formatChoices := []string{"JSON", "Markdown", "Both"}
	gumCmd = exec.Command("gum", "choose", "--header=Select output format:")
	gumCmd.Stdin = strings.NewReader(strings.Join(formatChoices, "\n"))
	gumCmd.Stderr = os.Stderr

	formatOutput, err := gumCmd.Output()
	if err != nil {
		return nil
	}

	selectedFormat := strings.TrimSpace(string(formatOutput))
	if selectedFormat == "" {
		return nil
	}

	for _, p := range projects {
		if p.Name == selectedProject {
			return generateProjectContext(&p, selectedFormat)
		}
	}

	return nil
}

func showHandoffMenu(projects []models.Project) error {
	var choices []string
	for _, p := range projects {
		choices = append(choices, p.Name)
	}

	gumCmd := exec.Command("gum", "choose", "--header=Select project for handoff:")
	gumCmd.Stdin = strings.NewReader(strings.Join(choices, "\n"))
	gumCmd.Stderr = os.Stderr

	output, err := gumCmd.Output()
	if err != nil {
		return nil
	}

	selectedProject := strings.TrimSpace(string(output))
	if selectedProject == "" {
		return nil
	}

	for _, p := range projects {
		if p.Name == selectedProject {
			return generateHandoffDocument(&p)
		}
	}

	return nil
}

func showProjectActions(project models.Project) error {
	choices := []string{
		"🚀 Open in IDE",
		"📋 View Details",
		"📄 Generate Handoff",
		"🤖 Generate Context",
		"📊 Show Insights",
		"🔄 Update Status",
	}

	gumCmd := exec.Command("gum", "choose", "--header=Select action:")
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

	switch selected {
	case "🚀 Open in IDE":
		return openProjectInIDE(project.Path, "")
	case "📋 View Details":
		return showProjectDetails(project.ID)
	case "📄 Generate Handoff":
		return generateHandoffDocument(&project)
	case "🤖 Generate Context":
		return generateProjectContext(&project, "Both")
	case "📊 Show Insights":
		return showProjectInsights(&project)
	case "🔄 Update Status":
		return updateProjectStatus(&project)
	}

	return nil
}

func openProjectInIDE(projectPath, ideName string) error {
	detector := utils.NewIDEDetector()
	
	if ideName == "" {
		ide := detector.GetPreferredIDE()
		if ide == nil {
			fmt.Println("No IDEs detected on system.")
			return nil
		}
		ideName = ide.Name
	}

	err := detector.LaunchProjectInBackground(projectPath, ideName)
	if err != nil {
		return fmt.Errorf("failed to open project in IDE: %w", err)
	}

	fmt.Printf("🚀 Opening project in %s...\n", ideName)
	return nil
}

func generateProjectContext(project *models.Project, format string) error {
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

	switch format {
	case "JSON":
		filename := fmt.Sprintf("/tmp/%s-context.json", project.Name)
		jsonData, err := context.ExportToJSON()
		if err != nil {
			return fmt.Errorf("failed to export JSON: %w", err)
		}
		err = os.WriteFile(filename, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write JSON file: %w", err)
		}
		fmt.Printf("📋 Context exported to: %s\n", filename)
		
	case "Markdown":
		filename := fmt.Sprintf("/tmp/%s-context.md", project.Name)
		mdData := context.ExportToMarkdown()
		err = os.WriteFile(filename, []byte(mdData), 0644)
		if err != nil {
			return fmt.Errorf("failed to write Markdown file: %w", err)
		}
		fmt.Printf("📋 Context exported to: %s\n", filename)
		
	case "Both":
		jsonFile := fmt.Sprintf("/tmp/%s-context.json", project.Name)
		mdFile := fmt.Sprintf("/tmp/%s-context.md", project.Name)
		
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
	}

	return nil
}

func generateHandoffDocument(project *models.Project) error {
	filename := fmt.Sprintf("/tmp/%s-handoff.md", project.Name)
	
	// Generate handoff document content
	handoffDoc, err := generateHandoffDocContent(project)
	if err != nil {
		return fmt.Errorf("failed to generate handoff: %w", err)
	}

	err = os.WriteFile(filename, []byte(handoffDoc), 0644)
	if err != nil {
		return fmt.Errorf("failed to write handoff file: %w", err)
	}

	fmt.Printf("📄 Handoff document generated: %s\n", filename)
	return nil
}

func showProjectInsights(project *models.Project) error {
	generator := utils.NewContextGenerator(project.Path)
	
	context, err := generator.GenerateAgentContext(
		project.Name,
		string(project.Status),
		project.Progress,
		project.Notes,
	)
	if err != nil {
		return fmt.Errorf("failed to generate insights: %w", err)
	}

	insights := fmt.Sprintf(`
📊 Project Insights: %s
========================

🎯 Overview:
  Type: %s
  Status: %s
  Progress: %d%%
  Activity Level: %s
  Last Modified: %s

🛠️ Technologies: %d detected
  %s

📁 Recent Activity: %d files in last 7 days
  %s

🚀 Quick Start:
  Setup: %v
  Dev: %s
  Test: %s

📦 Git Information:
  Has Uncommitted: %v
  Branch: %s

Generated at: %s
`, 
		project.Name,
		context.ProjectType,
		context.Status,
		context.Progress,
		context.ActivityLevel,
		context.LastModified.Format("2006-01-02 15:04:05"),
		len(context.Technologies),
		formatTechnologies(context.Technologies),
		len(context.RecentFiles),
		formatRecentFiles(context.RecentFiles),
		context.QuickStart.SetupCommands,
		context.QuickStart.DevCommand,
		context.QuickStart.TestCommand,
		context.GitInfo != nil && context.GitInfo.HasUncommitted,
		getGitBranch(context.GitInfo),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	gumCmd := exec.Command("gum", "pager")
	gumCmd.Stdin = strings.NewReader(insights)
	gumCmd.Stdout = os.Stdout
	gumCmd.Stderr = os.Stderr
	
	return gumCmd.Run()
}

func updateProjectStatus(project *models.Project) error {
	choices := []string{"active", "paused", "completed", "archived"}
	
	gumCmd := exec.Command("gum", "choose", "--header=Select new status:")
	gumCmd.Stdin = strings.NewReader(strings.Join(choices, "\n"))
	gumCmd.Stderr = os.Stderr

	output, err := gumCmd.Output()
	if err != nil {
		return nil
	}

	newStatus := strings.TrimSpace(string(output))
	if newStatus == "" {
		return nil
	}

	projectRepo := repository.NewProjectRepository(db.Conn())
	project.Status = models.ProjectStatus(newStatus)
	project.UpdatedAt = time.Now()
	
	if newStatus == "completed" {
		project.Progress = 100
	} else if newStatus == "active" && project.Progress == 100 {
		project.Progress = 90 // Reset from completed
	}

	err = projectRepo.Update(project)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("✅ Project status updated to: %s\n", newStatus)
	return nil
}

// Helper functions
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

func formatTechnologies(techs []utils.TechnologyInfo) string {
	var names []string
	for _, tech := range techs {
		if tech.Version != "" {
			names = append(names, fmt.Sprintf("%s %s", tech.Name, tech.Version))
		} else {
			names = append(names, tech.Name)
		}
	}
	return strings.Join(names, ", ")
}

func formatRecentFiles(files []utils.FileInfo) string {
	var names []string
	count := 0
	for _, file := range files {
		if count >= 5 { // Limit to 5 files
			break
		}
		names = append(names, file.Name)
		count++
	}
	if len(files) > 5 {
		names = append(names, fmt.Sprintf("... and %d more", len(files)-5))
	}
	return strings.Join(names, ", ")
}

func getGitBranch(gitInfo *utils.GitInfo) string {
	if gitInfo == nil {
		return "N/A"
	}
	return gitInfo.Branch
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

func init() {
	listCmd.Flags().BoolP("non-blocking", "n", false, "Run in non-blocking mode for agents")
	listCmd.Flags().BoolP("auto-ide", "a", false, "Automatically open selected project in IDE")
	listCmd.Flags().String("ide", "", "Specify IDE to use (vscode, windsurf, vim, nano)")
	listCmd.Flags().String("export-context", "", "Export context to specified file")
	rootCmd.AddCommand(listCmd)
}
