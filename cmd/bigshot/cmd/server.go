/*
Copyright 2020 The bigshot Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/DevopsArtFactory/bigshot/cmd/bigshot/cmd/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/executor"
)

// Create new server command
func NewServerCommand() *cobra.Command {
	return builder.NewCmd("server").
		WithDescription("Run bigshot workermanager as server").
		SetFlags().
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
