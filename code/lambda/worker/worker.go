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

package worker

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/code/lambda/env"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/shot"
)

type Worker struct{}

// NewWorker creates a new worker
func NewWorker() *Worker {
	return &Worker{}
}

// Run executes worker role
func (w *Worker) Run(envs env.Env, targets []string, slackURL []string) error {
	if len(envs.RunType) == 0 {
		envs.RunType = constants.DefaultShooter
	}
	return Shoot(envs.RunType, envs.Region, slackURL, targets)
}

// RunTest executes worker role for test
func (w *Worker) RunTest(workerType string, slackURL []string) error {
	example := []string{
		"https://weverse.io",
	}

	return Shoot(
		workerType,
		constants.DefaultRegion,
		slackURL,
		example,
	)
}

// Shoot runs api
func Shoot(t, region string, slackURL, targets []string) error {
	if len(targets) == 0 {
		return errors.New("no target specified")
	}

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

	f := func(url string, ch chan error) {
		shooter := shot.NewShooter(t, region)
		if shooter == nil {
			ch <- errors.New("cannot find the right shooter for lambda")
			return
		}
		shooter.SetTarget(url)
		shooter.SetRate(1)
		shooter.SetSlackURL(slackURL)
		err := shooter.Run()
		ch <- err
	}

	for _, target := range targets {
		wg.Add(1)
		go f(target, input)
	}
	wg.Wait()
	close(input)

	result := <-output

	for _, e := range result {
		logrus.Error(e.Error())
	}

	return nil
}
