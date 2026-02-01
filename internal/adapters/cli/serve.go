package cli

import (
	"fmt"
	"net/http"

	httpAdapter "github.com/kriuchkov/tock/internal/adapters/http"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		Long:  "Start an HTTP server that exposes activity management endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := getService(cmd)

			handler := httpAdapter.NewHandler(svc)
			handler.RegisterRoutes()

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("Starting HTTP Server on %s\n", addr)
			return http.ListenAndServe(addr, nil)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")

	return cmd
}
