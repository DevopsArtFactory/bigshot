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
	Mode       string
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
func (g *Generator) Init(flags builder.Flags, template *schema.Template) error {
	// Variables
	var ws []*worker.Worker
	var zipFile []byte
	var err error

	if err := CheckInternalSetting(template); err != nil {
		return err
	}

	regions, err := GetRegionsFromTemplate(template, flags.AllRegion)
	if err != nil {
		return err
	}

	// Check if regions in template.Targets exist in regions configuration
	regionIDs := getRegionIDs(regions)
	if template != nil {
		if err := CheckAvailableRegions(regionIDs, template.Targets); err != nil {
			return err
		}
	}

	// Worker Setup
	if len(flags.ZipFile) > 0 {
		zipFile, err = tools.ReadZipFile(flags.ZipFile)
		if err != nil {
			return err
		}
	}

	// for normal lambda ( Not provisioned in VPC )
	for _, region := range regions {
		ws = append(ws, worker.New(*region.Region, zipFile, template, flags.DryRun, false))
	}

	// setup controller for managing workers
	cont, err := controller.New(template)
	if err != nil {
		return err
	}

	if cont != nil {
		g.SetController(cont)
		controllerWorker := worker.New(cont.GetRegion(), zipFile, template, flags.DryRun, false)
		controllerWorker.SetMode(constants.ManagerMode)
		controllerWorker.SetTemplate(*cont.Template.Name)
		ws = append(ws, controllerWorker)
	}

	g.SetWorkers(ws)

	// for internal lambda -> provisioned in VPC
	if err := g.InitInternalWorkers(flags, template, zipFile); err != nil {
		return nil
	}

	return nil
}

// InitInternalWorkers initiates workers for bigshot
func (g *Generator) InitInternalWorkers(flags builder.Flags, template *schema.Template, zipFile []byte) error {
	var ws []*worker.Worker
	var regionIDs []string

	if template.Targets != nil {
		regionIDs = getInternalNeedsRegion(template.Targets)
	} else {
		regionIDs = constants.AllAWSRegions
	}

	// for internal lambda -> provisioned in VPC
	for _, region := range regionIDs {
		ws = append(ws, worker.New(region, zipFile, template, flags.DryRun, true))
	}
	g.AddWorkers(ws)

	return nil
}

// GetRegionsFromTemplate retrieves region list from configuration
func GetRegionsFromTemplate(template *schema.Template, allRegion bool) ([]schema.Region, error) {
	var regions []schema.Region

	if template == nil && template.Regions == nil {
		if allRegion {
			for _, region := range constants.AllAWSRegions {
				regions = append(regions, schema.Region{
					Region: &region,
				})
			}
		} else {
			defaultRegion, err := builder.GetDefaultRegion(constants.DefaultProfile)
			if err != nil {
				return nil, err
			}
			regions = append(regions, schema.Region{
				Region: &defaultRegion,
			})
		}
	} else {
		regions = template.Regions

		if allRegion {
			for _, region := range constants.AllAWSRegions {
				regions = append(regions, schema.Region{
					Region: &region,
				})
			}
		}
	}

	return regions, nil
}

// SetWorkers sets lambda structures
func (g *Generator) SetWorkers(ws []*worker.Worker) {
	g.Workers = ws
}

// AddWorkers adds lambda structures
func (g *Generator) AddWorkers(ws []*worker.Worker) {
	g.Workers = append(g.Workers, ws...)
}

// SetController sets controller structure
func (g *Generator) SetController(c *controller.Controller) {
	g.Controller = c
}

// getTargetRegions returns target regions for workers
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

// checkAvailableRegions checks if regions in target configuration exist in regions configuration
func CheckAvailableRegions(regions []string, targets []schema.Target) error {
	for _, target := range targets {
		if target.Regions != nil && len(target.Regions) > 0 {
			for _, region := range target.Regions {
				if !tools.IsStringInArray(region, regions) {
					return fmt.Errorf("%s is not in the region list: %s", region, *target.URL)
				}
			}
		}
	}

	return nil
}

// CheckInternalSetting checks if regions configuration has internal settings
func CheckInternalSetting(template *schema.Template) error {
	if template == nil && template.Targets == nil {
		return nil
	}

	for _, target := range template.Targets {
		if target.Internal != nil && *target.Internal {
			for _, region := range target.Regions {
				for _, regionObj := range template.Regions {
					if *regionObj.Region == region {
						if len(regionObj.SecurityGroups) == 0 || len(regionObj.Subnets) == 0 {
							return fmt.Errorf("%s region has no security groups or subnet settings in Region configuration", region)
						}
						break
					}
				}
			}
		}
	}

	return nil
}

// getRegionIDs retrieves regions IDs
func getRegionIDs(regions []schema.Region) []string {
	var ret []string
	for _, region := range regions {
		ret = append(ret, *region.Region)
	}

	return ret
}

// getInternalNeedsRegion retrieves region list for internal lambda
func getInternalNeedsRegion(targets []schema.Target) []string {
	var ret []string
	for _, target := range targets {
		if target.Internal != nil && *target.Internal {
			for _, region := range target.Regions {
				if !tools.IsStringInArray(region, ret) {
					ret = append(ret, region)
				}
			}
		}
	}

	return ret
}
