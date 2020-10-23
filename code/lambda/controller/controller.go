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

package controller

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/code/lambda/env"
	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Controller struct{}

// NewController creates a new controller
func NewController() *Controller {
	return &Controller{}
}

// Run executes controller role
func (c *Controller) Run(envs env.Env) error {
	dynamoDB := client.NewDynamoDBClient(envs.Region)

	item, err := dynamoDB.GetTemplate(envs.Template, tools.GenerateNewTableName())
	if err != nil {
		return err
	}

	return Trigger(item)
}

// RunTest executes controller role for tes
func (c *Controller) RunTest() error {
	sample := "sample-templates"
	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	dynamoDB := client.NewDynamoDBClient(region)
	item, err := dynamoDB.GetTemplate(sample, tools.GenerateNewTableName())
	if err != nil {
		return err
	}

	return Trigger(item)
}

// Trigger will invoke other regions' lambda
func Trigger(item map[string]*dynamodb.AttributeValue) error {
	regions := item["regions"]

	var targets []string
	for _, target := range item["targets"].L {
		data := target.M
		targets = append(targets, *data["url"].S)
	}

	payload, err := json.Marshal(map[string][]string{
		"targets":                 targets,
		constants.BigShotSlackURL: {*item[constants.BigShotSlackURL].S},
	})
	if err != nil {
		return err
	}

	interval := 300/len(regions.L) - 1

	for _, region := range regions.L {
		data := region.M
		regionID := *data["region"].S
		logrus.Infof("Lambda will be triggered in %s", regionID)

		lambdaClient := client.NewLambdaClient(regionID)
		if err := lambdaClient.Trigger(regionID, payload); err != nil {
			logrus.Error(err.Error())
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}
