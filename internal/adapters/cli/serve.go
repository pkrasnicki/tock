package cli

import (
	"fmt"
	"net/http"
	"strings"

	httpAdapter "github.com/kriuchkov/tock/internal/adapters/http"
	"github.com/kriuchkov/tock/internal/jira"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	var port int
	var corsOrigins string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		Long:  "Start an HTTP server that exposes activity management endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := getService(cmd)
			cfg := getConfig(cmd)

			// Initialize Jira client if configured
			var jiraClient *jira.Client
			if cfg.Jira.URL != "" && cfg.Jira.Username != "" && cfg.Jira.APIToken != "" {
				jiraClient = jira.NewClient(cfg.Jira.URL, cfg.Jira.Username, cfg.Jira.APIToken)
				fmt.Println("Jira integration enabled")
			}

			// Build handler options
			handlerOpts := httpAdapter.HandlerOptions{
				Service:    svc,
				JiraClient: jiraClient,
			}

			// Configure CORS if specified
			if corsOrigins != "" {
				// Parse comma-separated origins
				origins := strings.Split(corsOrigins, ",")
				for i := range origins {
					origins[i] = strings.TrimSpace(origins[i])
				}

				corsConfig := httpAdapter.CORSConfig{
					AllowedOrigins:   origins,
					AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
					ExposedHeaders:   []string{"Link"},
					AllowCredentials: false,
					MaxAge:           300,
				}
				handlerOpts.CORSConfig = &corsConfig
			}

			handler := httpAdapter.NewHandlerWithOptions(handlerOpts)
			handler.RegisterRoutes()

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("Starting HTTP Server on %s\n", addr)
			return http.ListenAndServe(addr, nil)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVar(&corsOrigins, "cors-origins", "", "Comma-separated list of allowed CORS origins (default: * for all origins)")

	return cmd
}
