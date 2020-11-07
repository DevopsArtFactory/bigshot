package cmd

import (
	"context"
	"github.com/DevopsArtFactory/bigshot/cmd/bigshot/cmd/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/executor"
	"github.com/spf13/cobra"
	"io"
)

// Create new server command
func NewServerCommand() *cobra.Command {
	return builder.NewCmd("server").
		WithDescription("Run bigshot workermanager as server").
		RunWithNoArgs(funcServer)
}

// funcServer run deployment
func funcServer(ctx context.Context, _ io.Writer) error {
	return executor.RunExecutor(ctx, func(executor executor.Executor) error {
		if err := executor.Runner.RunServer(); err != nil {
			return err
		}
		return nil
	})
}
