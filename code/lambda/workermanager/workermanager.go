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
	"sync"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/code/lambda/env"
	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
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
	var template schema.Template
	if err := dynamodbattribute.UnmarshalMap(item, &template); err != nil {
		return err
	}

	logLevel := *template.Log
	interval := *template.Interval/len(template.Regions) - 1
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

	f := func(regionData schema.Region, target schema.Target, ch chan error) {
		timeout := constants.DefaultTargetTimeout
		if template.Timeout != nil {
			timeout = *template.Timeout
		}
		logrus.Infof("%s, %s, %d", *target.URL, *target.Port, timeout)

		data := map[string]interface{}{
			"target":                   *target.URL,
			"port":                     *target.Port,
			"method":                   *target.Method,
			"timeout":                  timeout,
			constants.BigShotSlackURLs: template.SlackURLs,
		}

		if len(logLevel) > 0 {
			data["log_level"] = logLevel
		}

		if target.Body != nil {
			body := map[string]string{}
			for k, v := range target.Body {
				body[k] = v
			}
			data["body"] = body
		}

		if target.Header != nil {
			header := map[string]string{}
			for k, v := range target.Header {
				header[k] = v
			}
			data["header"] = header
		}

		payload, err := json.Marshal(data)
		if err != nil {
			ch <- err
			return
		}

		logrus.Infof("Lambda will be triggered in %s: %s", *regionData.Region, *target.URL)

		lambdaClient := client.NewLambdaClient(*regionData.Region)

		internal := false
		if target.Internal != nil {
			internal = *target.Internal
		}
		err = lambdaClient.Trigger(regionData.Region, template.Name, payload, internal)
		if err == nil {
			logrus.Infof("function is successfully triggered: %s, %s, %s", *regionData.Region, *target.Port, *target.URL)
		}

		ch <- err
	}

	for _, target := range template.Targets {
		for _, region := range template.Regions {
			if target.Internal != nil && *target.Internal {
				if !tools.IsStringInArray(*region.Region, target.Regions) {
					continue
				}
			}

			wg.Add(1)
			go f(region, target, input)
		}
	}

	wg.Wait()
	close(input)

	errors := <-output

	for _, err := range errors {
		logrus.Errorln(err.Error())
	}

	return nil
}
