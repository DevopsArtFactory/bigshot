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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/code/lambda/controller"
	"github.com/DevopsArtFactory/bigshot/code/lambda/env"
	"github.com/DevopsArtFactory/bigshot/code/lambda/event"
	"github.com/DevopsArtFactory/bigshot/code/lambda/worker"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
)

type Flags struct {
	Type     string
	Test     bool
	Mode     string
	SlackURL string
}

// main starts lambda function
func main() {
	f := GetFlags()
	if f.Test {
		if err := RunTest(f); err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	lambda.Start(HandleRequest)
}

// Lambda handler
func HandleRequest(ctx context.Context, evt event.Event) error {
	fmt.Println(evt)
	if err := Run(evt); err != nil {
		return err
	}
	return nil
}

// GetFlags returns flags
func GetFlags() *Flags {
	t := flag.String("type", "vegeta", "type of shot checker")
	test := flag.Bool("test", false, "whether or not this run is test")
	mode := flag.String("mode", constants.WorkerMode, "mode of lambda")
	slack := flag.String("slack-url", constants.EmptyString, "slack URLs to test (comma delimiter)")

	flag.Parse()

	return &Flags{
		Type:     *t,
		Test:     *test,
		Mode:     *mode,
		SlackURL: *slack,
	}
}

// Run executes main process of lambda
func Run(evt event.Event) error {
	envs := env.GetEnvs()

	if len(envs.Region) == 0 {
		return fmt.Errorf("region is not specified. please check environment variables")
	}
	logrus.Infof("this is lambda function in %s", envs.Region)

	switch envs.Mode {
	case constants.ControllerMode:
		return controller.NewController().Run(envs)
	case constants.WorkerMode:
		return worker.NewWorker().Run(envs, evt)
	}

	return nil
}

// RunTest executes main process of lambda for test
func RunTest(flags *Flags) error {
	switch flags.Mode {
	case constants.ControllerMode:
		return controller.NewController().RunTest()
	case constants.WorkerMode:
		slackURL := flags.SlackURL
		var slacks []string
		if len(slackURL) > 0 {
			slacks = append(slacks, slackURL)
		}
		return worker.NewWorker().RunTest(flags.Type, slacks)
	}

	return nil
}
