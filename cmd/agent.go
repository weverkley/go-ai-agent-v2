package cmd

import (
	"go-ai-agent-v2/go-cli/pkg/server"
	"github.com/spf13/cobra"
)

// runAgentCmd will start the agent server.
func runAgentCmd(rootCmd *cobra.Command, cmd *cobra.Command, args []string) {
	srv := server.NewServer(chatService, SessionService)
	srv.Start(":8080")
}
