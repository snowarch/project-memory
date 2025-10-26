package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/snowarch/project-memory/internal/database"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project memory database",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		fmt.Printf("Database initialized successfully at: %s\n", dbPath)
		return nil
	},
}
