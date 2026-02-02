package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/tock/internal/adapters/syncstate"
	"github.com/kriuchkov/tock/internal/services/sync"

	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize activities with Jira",
		Long: `Synchronizes completed activities with Jira worklogs.
		
Activities with a 'jira' attribute will be synced to the corresponding Jira issue.
The command handles:
- Adding new worklogs for unsynced activities
- Updating existing worklogs when times change
- Deleting worklogs when activities are removed or jira attribute is removed
- Moving worklogs when jira attribute value changes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo := getRepository(cmd)
			cfg := getConfig(cmd)

			if cfg.Jira.URL == "" {
				return errors.New("Jira configuration is missing. Please configure jira.url, jira.username, and jira.api_token in ~/.config/tock/tock.yaml")
			}

			// Create sync state repository - use the data directory from file path
			dataDir := filepath.Dir(cfg.File.Path)
			syncRepo := syncstate.NewRepository(dataDir)

			syncService := sync.NewService(repo, syncRepo, cfg.Jira)

			fmt.Println("Synchronizing activities with Jira...")
			result, err := syncService.Sync(context.Background())
			if err != nil {
				return errors.Wrap(err, "sync failed")
			}

			// Print results
			fmt.Printf("\n✓ Sync completed:\n")
			fmt.Printf("  • Synced (new):    %d\n", result.Synced)
			fmt.Printf("  • Updated:         %d\n", result.Updated)
			fmt.Printf("  • Deleted:         %d\n", result.Deleted)
			fmt.Printf("  • Skipped:         %d\n", result.Skipped)

			if len(result.Errors) > 0 {
				fmt.Printf("\n✗ Errors:\n")
				for _, err := range result.Errors {
					fmt.Printf("  • %s\n", err)
				}
			}

			return nil
		},
	}

	return cmd
}
