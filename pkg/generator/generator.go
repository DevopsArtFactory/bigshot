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

package generator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/controller"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
	"github.com/DevopsArtFactory/bigshot/pkg/worker"
)

type Generator struct {
	Controller *controller.Controller
	Workers    []*worker.Worker
	Channel    *Channel
}

type Channel struct {
	Input  chan error
	Output chan []error
}

// New creates a new Generator struct
func New() *Generator {
	gn := Generator{
		Controller: nil,
		Workers:    []*worker.Worker{},
		Channel: &Channel{
			Input:  make(chan error),
			Output: make(chan []error),
		},
	}

	return &gn
}

// Init initiates a generator for bigshot
func (g *Generator) Init(flags builder.Flags, config *schema.Config) error {
	// Variables
	var ws []*worker.Worker
	var zipFile []byte
	var err error

	// Worker Setup
	if len(flags.ZipFile) > 0 {
		zipFile, err = tools.ReadZipFile(flags.ZipFile)
		if err != nil {
			return err
		}
	}

	regions, err := GetTargetRegions(flags.Region, flags.AllRegion, false)
	if err != nil {
		return err
	}

	for _, region := range regions {
		ws = append(ws, worker.New(region, zipFile))
	}

	// Controller setup
	cont, err := controller.New(config)
	if err != nil {
		return err
	}

	if cont != nil {
		g.SetController(cont)
		controllerWorker := worker.New(cont.GetRegion(), zipFile)
		controllerWorker.SetMode(constants.ControllerMode)
		if cont.Config != nil {
			controllerWorker.SetTemplate(cont.Config.Name)
		}
		ws = append(ws, controllerWorker)
	}

	g.SetWorkers(ws)

	return nil
}

// SetWorkers sets lambda structures
func (g *Generator) SetWorkers(ws []*worker.Worker) {
	g.Workers = ws
}

// SetController sets controller structure
func (g *Generator) SetController(c *controller.Controller) {
	g.Controller = c
}

// getTargetRegions  returns target regions for workers
func GetTargetRegions(specified string, applyAllRegion, isTest bool) ([]string, error) {
	var targetRegions []string
	var err error

	if applyAllRegion {
		return constants.AllAWSRegions, nil
	}

	if len(specified) > 0 {
		split := strings.Split(specified, ",")
		for _, s := range split {
			if !tools.IsStringInArray(s, constants.AllAWSRegions) {
				return nil, fmt.Errorf("wrong region key is specified: %s", s)
			}
			targetRegions = append(targetRegions, s)
		}
	} else {
		targetRegions = constants.AllAWSRegions
		if !isTest {
			targetRegions, err = SelectFromCommand()
			if err != nil {
				return nil, err
			}
		}
	}

	return targetRegions, nil
}

// SelectFromCommand selects regions from the list
func SelectFromCommand() ([]string, error) {
	var targets []string
	prompt := &survey.MultiSelect{
		Message: "Pick regions:",
		Options: constants.AllAWSRegions,
	}
	survey.AskOne(prompt, &targets)

	if len(targets) == 0 {
		return nil, errors.New("you have to choose at least one region")
	}

	return targets, nil
}
