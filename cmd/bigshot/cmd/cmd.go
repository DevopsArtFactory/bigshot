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
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
	"github.com/DevopsArtFactory/bigshot/pkg/version"
)

var (
	cfgFile string
	v       string
)

// Get root command
func NewRootCommand(out, stderr io.Writer) *cobra.Command {
	cobra.OnInitialize(initConfig)
	rootCmd := &cobra.Command{
		Use:           "bigshot",
		Short:         "Open source synthetic application on AWS",
		Long:          "Open source synthetic application on AWS",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().SetOutput(out)

			// Setup logs
			if err := tools.SetUpLogs(stderr, v); err != nil {
				return err
			}

			version := version.Get()

			logrus.Debugf("bigshot %+v", version)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	//Group by commands
	rootCmd.AddCommand(NewUpdateCodeCommand())
	rootCmd.AddCommand(NewUpdateTemplateCommand())
	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewRunCommand())
	rootCmd.AddCommand(NewStopCommand())
	rootCmd.AddCommand(NewDeleteCommand())
	rootCmd.AddCommand(NewDestroyCommand())
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(NewCmdVersion())

	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", constants.DefaultLogLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
	}

	viper.AutomaticEnv() // read in environment variables that match
}
