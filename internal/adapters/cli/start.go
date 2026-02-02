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

			// Get git context if in a git repository
			var attributes []models.Attribute
			gitProject := ""
			gitDescription := ""

			if gitutil.IsGitRepository("") {
				branch, err := gitutil.GetCurrentBranch()
				remotes := make(map[string]string)

				if err == nil {
					// Get remotes for variable expansion
					if r, err := gitutil.GetRemotes(); err == nil {
						remotes = r
					}

					// Match patterns and get project, description, and attributes
					gitProject, gitDescription, attributes = matchBranchPatterns(branch, remotes, cfg.AttributePatterns)
				}
			}

			// Use git-detected values if not provided via flags
			if project == "" && gitProject != "" {
				project = gitProject
			}
			if description == "" && gitDescription != "" {
				description = gitDescription
			}

			// Validate required fields
			if project == "" {
				return errors.New("project is required (provide via -p flag or git pattern)")
			}
			if description == "" {
				return errors.New("description is required (provide via -d flag or git pattern)")
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

	return cmd
}

// matchBranchPatterns matches the branch name against configured patterns and returns project, description, and attributes
func matchBranchPatterns(branch string, remotes map[string]string, patterns []config.AttributePattern) (string, string, []models.Attribute) {
	var project, description string
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

		// Build variables map for expansion
		variables := make(map[string]string)

		// Add capture groups ($0, $1, $2, etc.)
		for i, match := range matches {
			variables[fmt.Sprintf("$%d", i)] = match
		}

		// Add remote variables ($remote_origin, $remote_upstream, etc.)
		for remoteName, remoteURL := range remotes {
			varName := fmt.Sprintf("$remote_%s", remoteName)

			// Check if there's a mapping for this remote URL
			mappedValue := remoteURL
			if pattern.RemoteMappings != nil {
				for urlPattern, mappedName := range pattern.RemoteMappings {
					// Support both exact match and regex pattern
					if urlPattern == remoteURL || regexp.MustCompile(urlPattern).MatchString(remoteURL) {
						mappedValue = mappedName
						break
					}
				}
			}

			variables[varName] = mappedValue
		}

		// Expand project if specified in pattern
		if pattern.Project != "" && project == "" {
			project = expandVariables(pattern.Project, variables)
		}

		// Expand description if specified in pattern
		if pattern.Description != "" && description == "" {
			description = expandVariables(pattern.Description, variables)
		}

		// Add all attributes from the matched pattern
		for key, value := range pattern.Attributes {
			attributes = append(attributes, models.Attribute{
				Key:   key,
				Value: expandVariables(value, variables),
			})
		}
	}

	return project, description, attributes
}

// expandVariables replaces all variable placeholders in the template string
func expandVariables(template string, variables map[string]string) string {
	result := template
	for varName, varValue := range variables {
		result = regexp.MustCompile(regexp.QuoteMeta(varName)).ReplaceAllString(result, varValue)
	}
	return result
}
