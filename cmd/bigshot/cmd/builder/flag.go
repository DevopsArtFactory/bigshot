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

package builder

import (
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Flag struct {
	Name               string
	Shorthand          string
	Usage              string
	Value              interface{}
	DefValue           interface{}
	DefValuePerCommand map[string]interface{}
	FlagAddMethod      string
	DefinedOn          []string
	Hidden             bool

	pflag *pflag.Flag
}

// FlagRegistry is a list of all bigshot CLI flags.
var FlagRegistry = []Flag{
	{
		Name:          "config",
		Shorthand:     "c",
		Usage:         "Configuration file for bigshot",
		Value:         aws.String(constants.EmptyString),
		DefValue:      constants.EmptyString,
		FlagAddMethod: "StringVar",
		DefinedOn:     []string{"init", "update-template"},
	},
	{
		Name:          "region",
		Shorthand:     "r",
		Usage:         "Run command to specific region",
		Value:         aws.String(constants.EmptyString),
		DefValue:      constants.EmptyString,
		FlagAddMethod: "StringVar",
		DefinedOn:     []string{"init", "destroy", "update-code"},
	},
	{
		Name:          "all",
		Shorthand:     "A",
		Usage:         "Select all regions of provider",
		Value:         aws.Bool(false),
		DefValue:      false,
		FlagAddMethod: "BoolVar",
		DefinedOn:     []string{"init", "destroy", "update-code"},
	},
	{
		Name:          "zip-file",
		Shorthand:     "z",
		Usage:         "Specify the zip file of code for lambda lambda",
		Value:         aws.String(constants.EmptyString),
		DefValue:      constants.EmptyString,
		FlagAddMethod: "StringVar",
		DefinedOn:     []string{"init", "update-code"},
	},
	{
		Name:          "interval",
		Shorthand:     "i",
		Usage:         "time of interval to run bigshot in seconds",
		Value:         aws.Int(300),
		DefValue:      300,
		FlagAddMethod: "IntVar",
		DefinedOn:     []string{"run"},
	},
}

// flag sets pflag.Flag with custom Flag struct
func (fl *Flag) flag() *pflag.Flag {
	if fl.pflag != nil {
		return fl.pflag
	}

	inputs := []interface{}{fl.Value, fl.Name}
	if fl.FlagAddMethod != "Var" {
		inputs = append(inputs, fl.DefValue)
	}
	inputs = append(inputs, fl.Usage)

	fs := pflag.NewFlagSet(fl.Name, pflag.ContinueOnError)
	reflect.ValueOf(fs).MethodByName(fl.FlagAddMethod).Call(reflectValueOf(inputs))
	f := fs.Lookup(fl.Name)
	f.Shorthand = fl.Shorthand
	f.Hidden = fl.Hidden

	fl.pflag = f
	return f
}

func reflectValueOf(values []interface{}) []reflect.Value {
	var results []reflect.Value
	for _, v := range values {
		results = append(results, reflect.ValueOf(v))
	}
	return results
}

//Add command flags
func SetCommandFlags(cmd *cobra.Command) {
	var flagsForCommand []*Flag
	for i := range FlagRegistry {
		fl := &FlagRegistry[i]

		if tools.IsStringInArray(cmd.Use, fl.DefinedOn) {
			cmd.Flags().AddFlag(fl.flag())
			flagsForCommand = append(flagsForCommand, fl)
		}
	}

	// Apply command-specific default values to flags.
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Update default values.
		for _, fl := range flagsForCommand {
			viper.BindPFlag(fl.Name, cmd.Flags().Lookup(fl.Name))
		}

		// Since PersistentPreRunE replaces the parent's PersistentPreRunE,
		// make sure we call it, if it is set.
		if parent := cmd.Parent(); parent != nil {
			if preRun := parent.PersistentPreRunE; preRun != nil {
				if err := preRun(cmd, args); err != nil {
					return err
				}
			} else if preRun := parent.PersistentPreRun; preRun != nil {
				preRun(cmd, args)
			}
		}

		return nil
	}
}
