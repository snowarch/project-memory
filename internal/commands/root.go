package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/database"
	"github.com/snowarch/project-memory/internal/logger"
)

var (
	dbPath       string
	db           *database.DB
	rootPath     string
	verbose      bool
	quiet        bool
	statusFilter string
	limit        int
	offset       int
)

var rootCmd = &cobra.Command{
	Use:   "pmem",
	Short: "Project Memory Bank - Professional project tracking system",
	Long:  `A professional CLI tool for tracking, analyzing, and managing development projects with AI assistance.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configurar nivel de logging
		if verbose {
			logger.SetLevel(logger.LevelDebug)
		} else if quiet {
			logger.SetLevel(logger.LevelError)
		} else {
			logger.SetLevel(logger.LevelInfo)
		}

		if cmd.Name() == "init" {
			return
		}

		var err error
		db, err = database.New(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
			fmt.Fprintf(os.Stderr, "Run 'pmem init' first\n")
			os.Exit(1)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	defaultDBPath := filepath.Join(homeDir, ".local", "share", "pmem", "projects.db")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "Database file path")
	rootCmd.PersistentFlags().StringVarP(&rootPath, "path", "p", "", "Root path to scan for projects")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output")
	
	// Register all commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(statusCmd)
	
	// Configure command-specific flags
	analyzeCmd.Flags().StringVar(&apiKey, "api-key", "", "Groq API key (or set GROQ_API_KEY env var)")
	
	listCmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter by status (active, paused, archived, completed)")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum number of projects to show")
	listCmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
}
