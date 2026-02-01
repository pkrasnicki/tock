package cli

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/tock/internal/config"
	"github.com/kriuchkov/tock/internal/core/dto"
	"github.com/kriuchkov/tock/internal/core/models"
	"github.com/kriuchkov/tock/internal/gitutil"

	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	var description string
	var project string
	var at string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new activity",
		RunE: func(cmd *cobra.Command, _ []string) error {
			service := getService(cmd)
			tf := getTimeFormatter(cmd)
			cfg := getConfig(cmd)

			startTime := time.Now()
			if at != "" {
				var err error
				startTime, err = tf.ParseTime(at)
				if err != nil {
					return errors.Wrap(err, "parse time")
				}
			}

			// Get attributes from git branch if in a git repository
			var attributes []models.Attribute
			if gitutil.IsGitRepository("") {
				if branch, err := gitutil.GetCurrentBranch(); err == nil {
					attributes = matchBranchPatterns(branch, cfg.AttributePatterns)
				}
			}

			req := dto.StartActivityRequest{
				Description: description,
				Project:     project,
				StartTime:   startTime,
				Attributes:  attributes,
			}

			activity, err := service.Start(context.Background(), req)
			if err != nil {
				return errors.Wrap(err, "start activity")
			}

			fmt.Printf(
				"Started activity: %s | %s at %s\n",
				activity.Project,
				activity.Description,
				activity.StartTime.Format(tf.GetDisplayFormat()),
			)

			// Print attributes if any
			if len(activity.Attributes) > 0 {
				fmt.Printf("Attributes: ")
				for i, attr := range activity.Attributes {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s=%s", attr.Key, attr.Value)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Activity description")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name")
	cmd.Flags().StringVarP(&at, "time", "t", "", "Start time (HH:MM)")
	if err := cmd.MarkFlagRequired("description"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}

	return cmd
}

// matchBranchPatterns matches the branch name against configured patterns and returns attributes
func matchBranchPatterns(branch string, patterns []config.AttributePattern) []models.Attribute {
	var attributes []models.Attribute

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			continue
		}

		matches := re.FindStringSubmatch(branch)
		if matches == nil {
			continue
		}

		// Add all attributes from the matched pattern
		// Replace $1, $2, etc. with captured groups
		for key, value := range pattern.Attributes {
			expandedValue := value

			// Replace $0 with full match, $1 with first group, etc.
			for i, match := range matches {
				placeholder := fmt.Sprintf("$%d", i)
				expandedValue = regexp.MustCompile(regexp.QuoteMeta(placeholder)).ReplaceAllString(expandedValue, match)
			}

			attributes = append(attributes, models.Attribute{
				Key:   key,
				Value: expandedValue,
			})
		}
	}

	return attributes
}
