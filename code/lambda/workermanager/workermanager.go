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

package workermanager

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/code/lambda/env"
	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type WorkerManager struct{}

// New creates a new worker manager
func New() *WorkerManager {
	return &WorkerManager{}
}

// Run executes worker manager role
func (w *WorkerManager) Run(envs env.Env) error {
	dynamoDB := client.NewDynamoDBClient(envs.Region)

	item, err := dynamoDB.GetTemplate(envs.Template, tools.GenerateNewTableName())
	if err != nil {
		return err
	}

	return Trigger(item)
}

// RunTest executes workermanager role for tes
func (w *WorkerManager) RunTest() error {
	sample := "base-production"
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
	template := *item["name"].S
	regions := item["regions"]
	totalInterval, err := strconv.Atoi(*item["interval"].N)
	if err != nil {
		return err
	}

	interval := totalInterval/len(regions.L) - 1
	logrus.Infof("Interval: %d", interval)

	var wg sync.WaitGroup
	input := make(chan error)
	output := make(chan []error)
	defer close(output)

	go func(input chan error, output chan []error, wg *sync.WaitGroup) {
		var ret []error
		for err := range input {
			if err != nil {
				ret = append(ret, err)
			}
			wg.Done()
		}

		output <- ret
	}(input, output, &wg)

	var slackURLs []string
	for _, slack := range item[constants.BigShotSlackURLs].SS {
		slackURLs = append(slackURLs, *slack)
	}

	f := func(regionData map[string]*dynamodb.AttributeValue, target *dynamodb.AttributeValue, ch chan error) {
		data := target.M
		m := map[string]interface{}{
			"target":                   *data["url"].S,
			"method":                   *data["method"].S,
			constants.BigShotSlackURLs: slackURLs,
		}

		if _, ok := data["body"]; ok {
			body := map[string]string{}
			for k, v := range data["body"].M {
				body[k] = *v.S
			}
			m["body"] = body
		}

		if _, ok := data["header"]; ok {
			header := map[string]string{}
			for k, v := range data["header"].M {
				header[k] = *v.S
			}
			m["header"] = header
		}

		payload, err := json.Marshal(m)
		if err != nil {
			ch <- err
			return
		}

		regionID := *regionData["region"].S
		logrus.Infof("Lambda will be triggered in %s: %s", regionID, *data["url"].S)

		lambdaClient := client.NewLambdaClient(regionID)
		err = lambdaClient.Trigger(regionID, template, payload)
		ch <- err
	}

	for _, region := range regions.L {
		regionData := region.M
		for _, target := range item["targets"].L {
			wg.Add(1)
			go f(regionData, target, input)
		}

		logrus.Infof("region %s is done", *regionData["region"].S)
	}

	wg.Wait()
	close(input)

	errors := <-output

	for _, err := range errors {
		logrus.Errorln(err.Error())
	}

	return nil
}
