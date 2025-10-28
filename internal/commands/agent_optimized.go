package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/repository"
	"github.com/snowarch/project-memory/internal/scanner"
	"github.com/snowarch/project-memory/internal/utils"
)

var agentOptimizedCmd = &cobra.Command{
	Use:   "agent-optimized",
	Short: "Zero-input optimized commands for AI agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		action, _ := cmd.Flags().GetString("action")
		project, _ := cmd.Flags().GetString("project")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		batch, _ := cmd.Flags().GetString("batch")

		// Check environment variables for agent integration
		if action == "" {
			action = os.Getenv("PMEM_ACTION")
		}
		if project == "" {
			project = os.Getenv("PMEM_PROJECT")
		}
		if format == "" {
			format = os.Getenv("PMEM_FORMAT")
		}
		if output == "" {
			output = os.Getenv("PMEM_OUTPUT")
		}

		// Default values
		if format == "" {
			format = "json"
		}

		switch action {
		case "list":
			return listProjectsOptimized(format, output)
		case "discover":
			return discoverProjectsOptimized(output)
		case "context":
			if project == "" {
				return fmt.Errorf("project name required for context action")
			}
			return getProjectContextOptimized(project, format, output)
		case "batch-context":
			if batch == "" {
				return fmt.Errorf("batch file required for batch-context action")
			}
			return getBatchContextOptimized(batch, format, output)
		case "open":
			if project == "" {
				return fmt.Errorf("project name required for open action")
			}
			return openProjectOptimized(project)
		case "status":
			if project == "" {
				return fmt.Errorf("project name required for status action")
			}
			return getProjectStatusOptimized(project)
		case "search":
			if project == "" {
				return fmt.Errorf("search query required")
			}
			return searchProjectsOptimized(project, format, output)
		case "technologies":
			return listTechnologiesOptimized(format, output)
		case "health":
			return healthCheckOptimized()
		default:
			return showAgentHelp()
		}
	},
}

func listProjectsOptimized(format, output string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	
	projects, err := projectRepo.List("", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(projects, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return writeOutput(data, output)
	case "csv":
		var csv strings.Builder
		csv.WriteString("ID,Name,Path,Status,Progress,CreatedAt,UpdatedAt\n")
		for _, p := range projects {
			csv.WriteString(fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s\n",
				p.ID, p.Name, p.Path, p.Status, p.Progress,
				p.CreatedAt.Format("2006-01-02 15:04:05"),
				p.UpdatedAt.Format("2006-01-02 15:04:05")))
		}
		return writeOutput([]byte(csv.String()), output)
	case "simple":
		var simple strings.Builder
		for _, p := range projects {
			simple.WriteString(fmt.Sprintf("%s|%s|%s|%d\n", p.Name, p.Path, p.Status, p.Progress))
		}
		return writeOutput([]byte(simple.String()), output)
	default:
		return fmt.Errorf("unsupported format: %s (use json, csv, simple)", format)
	}
}

func discoverProjectsOptimized(output string) error {
	// Scan current directory for new projects
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	scanner := scanner.New(cwd)
	
	projects, err := scanner.ScanProjects()
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return writeOutput(data, output)
}

func getProjectContextOptimized(projectName, format, output string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	
	projects, err := projectRepo.Search(projectName)
	if err != nil {
		return fmt.Errorf("failed to search projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("project not found: %s", projectName)
	}

	project := projects[0]
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
	case "json":
		data, err := json.MarshalIndent(context, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return writeOutput(data, output)
	case "markdown":
		mdData := context.ExportToMarkdown()
		return writeOutput([]byte(mdData), output)
	case "summary":
		summary := fmt.Sprintf(`Project: %s
Type: %s
Status: %s
Progress: %d%%
Activity: %s
Technologies: %d
Recent Files: %d
Git: %v
Path: %s`,
			context.ProjectName,
			context.ProjectType,
			context.Status,
			context.Progress,
			context.ActivityLevel,
			len(context.Technologies),
			len(context.RecentFiles),
			context.GitInfo != nil,
			context.ProjectPath)
		return writeOutput([]byte(summary), output)
	default:
		return fmt.Errorf("unsupported format: %s (use json, markdown, summary)", format)
	}
}

func getBatchContextOptimized(batchFile, format, output string) error {
	// Read batch file with project names (one per line)
	data, err := os.ReadFile(batchFile)
	if err != nil {
		return fmt.Errorf("failed to read batch file: %w", err)
	}

	projectNames := strings.Split(strings.TrimSpace(string(data)), "\n")
	var contexts []map[string]interface{}

	projectRepo := repository.NewProjectRepository(db.Conn())

	for _, name := range projectNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		projects, err := projectRepo.Search(name)
		if err != nil || len(projects) == 0 {
			continue
		}

		project := projects[0]
		generator := utils.NewContextGenerator(project.Path)
		
		context, err := generator.GenerateAgentContext(
			project.Name,
			string(project.Status),
			project.Progress,
			project.Notes,
		)
		if err != nil {
			continue
		}

		var contextData interface{}
		switch format {
		case "json":
			contextData = context
		case "markdown":
			contextData = context.ExportToMarkdown()
		case "summary":
			contextData = fmt.Sprintf("%s|%s|%s|%d|%s", 
				context.ProjectName, context.ProjectType, context.Status,
				context.Progress, context.ActivityLevel)
		}

		contexts = append(contexts, map[string]interface{}{
			"project": project.Name,
			"context": contextData,
		})
	}

	result := map[string]interface{}{
		"contexts":  contexts,
		"count":     len(contexts),
		"format":    format,
		"timestamp": utils.GetCurrentTimestamp(),
	}

	resultData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	return writeOutput(resultData, output)
}

func openProjectOptimized(projectName string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	
	projects, err := projectRepo.Search(projectName)
	if err != nil {
		return fmt.Errorf("failed to search projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("project not found: %s", projectName)
	}

	project := projects[0]
	detector := utils.NewIDEDetector()
	
	ide := detector.GetPreferredIDE()
	if ide == nil {
		return fmt.Errorf("no IDEs detected on system")
	}

	err = detector.LaunchProjectInBackground(project.Path, ide.Name)
	if err != nil {
		return fmt.Errorf("failed to open project: %w", err)
	}

	// Silent success for agent use
	return nil
}

func getProjectStatusOptimized(projectName string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	
	projects, err := projectRepo.Search(projectName)
	if err != nil {
		return fmt.Errorf("failed to search projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("project not found: %s", projectName)
	}

	project := projects[0]
	
	// Output machine-readable status
	fmt.Printf("%s|%s|%d|%s|%d\n", 
		project.Name, 
		project.Status, 
		project.Progress,
		project.UpdatedAt.Format("2006-01-02 15:04:05"),
		int(project.UpdatedAt.Unix()))

	return nil
}

func searchProjectsOptimized(query, format, output string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	
	projects, err := projectRepo.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search projects: %w", err)
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(projects, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return writeOutput(data, output)
	case "names":
		var names strings.Builder
		for _, p := range projects {
			names.WriteString(p.Name + "\n")
		}
		return writeOutput([]byte(names.String()), output)
	case "paths":
		var paths strings.Builder
		for _, p := range projects {
			paths.WriteString(p.Path + "\n")
		}
		return writeOutput([]byte(paths.String()), output)
	default:
		return fmt.Errorf("unsupported format: %s (use json, names, paths)", format)
	}
}

func listTechnologiesOptimized(format, output string) error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	techRepo := repository.NewTechnologyRepository(db.Conn())
	
	projects, err := projectRepo.List("", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	techMap := make(map[string][]string)
	for _, p := range projects {
		techs, err := techRepo.GetByProject(p.ID)
		if err != nil {
			continue
		}
		for _, tech := range techs {
			techMap[tech.Name] = append(techMap[tech.Name], p.Name)
		}
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(techMap, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return writeOutput(data, output)
	case "list":
		var list strings.Builder
		for tech, projects := range techMap {
			list.WriteString(fmt.Sprintf("%s: %s\n", tech, strings.Join(projects, ", ")))
		}
		return writeOutput([]byte(list.String()), output)
	case "count":
		fmt.Printf("%d\n", len(techMap))
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (use json, list, count)", format)
	}
}

func healthCheckOptimized() error {
	projectRepo := repository.NewProjectRepository(db.Conn())
	
	projects, err := projectRepo.List("", 1, 0)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return err
	}

	fmt.Printf("OK|projects:%d|timestamp:%d\n", len(projects), utils.GetCurrentTimestamp())
	return nil
}

func showAgentHelp() error {
	helpText := `Project Memory Agent Interface

Actions:
  list                 - List all projects
  discover             - Discover projects in current directory
  context              - Get project context
  batch-context        - Get context for multiple projects
  open                 - Open project in IDE
  status               - Get project status
  search <query>       - Search projects
  technologies         - List all technologies
  health               - Health check

Formats:
  json                 - JSON output (default)
  csv                  - CSV format for list
  simple               - Simple pipe-separated format
  markdown             - Markdown format
  summary              - Human-readable summary
  names                - Project names only
  paths                - Project paths only
  list                 - List format
  count                - Count only

Environment Variables:
  PMEM_ACTION          - Action to perform
  PMEM_PROJECT         - Project name
  PMEM_FORMAT          - Output format
  PMEM_OUTPUT          - Output file

Examples:
  pmem agent-optimized --action list --format json
  pmem agent-optimized --action context --project myproject --format summary
  PMEM_ACTION=context PMEM_PROJECT=myproject pmem agent-optimized
  pmem agent-optimized --action batch-context --batch projects.txt --format json

Flags:
  --action string       - Action to perform
  --project string      - Project name
  --format string       - Output format (default: json)
  --output string       - Output file
  --batch string        - Batch file for multiple projects
`
	fmt.Print(helpText)
	return nil
}

func writeOutput(data []byte, output string) error {
	if output != "" {
		return os.WriteFile(output, data, 0644)
	}
	fmt.Print(string(data))
	return nil
}

func init() {
	agentOptimizedCmd.Flags().String("action", "", "Action to perform")
	agentOptimizedCmd.Flags().String("project", "", "Project name")
	agentOptimizedCmd.Flags().String("format", "json", "Output format")
	agentOptimizedCmd.Flags().String("output", "", "Output file")
	agentOptimizedCmd.Flags().String("batch", "", "Batch file for multiple projects")
	rootCmd.AddCommand(agentOptimizedCmd)
}
