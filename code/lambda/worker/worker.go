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
	"net/url"

	"github.com/DevopsArtFactory/bigshot/code/lambda/env"
	"github.com/DevopsArtFactory/bigshot/code/lambda/event"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/shot"
)

type Worker struct{}

// NewWorker creates a new worker
func NewWorker() *Worker {
	return &Worker{}
}

// Run executes worker role
func (w *Worker) Run(envs env.Env, evt event.Event) error {
	if len(envs.RunType) == 0 {
		envs.RunType = constants.DefaultShooter
	}
	return Shoot(envs.RunType, envs.Region, evt)
}

// RunTest executes worker role for test
func (w *Worker) RunTest(workerType string, slackURLs []string) error {
	evts := []event.Event{
		{
			Target:    "https://www.google.com",
			Method:    "GET",
			SlackURLs: slackURLs,
		},
		{
			Target:    "https://www.amazon.com",
			Method:    "GET",
			SlackURLs: slackURLs,
		},
	}

	for _, evt := range evts {
		if err := Shoot(
			workerType,
			constants.DefaultRegion,
			evt,
		); err != nil {
			return err
		}
	}

	return nil
}

// Shoot runs api
func Shoot(t, region string, evt event.Event) error {
	if len(evt.Target) == 0 {
		return errors.New("no target specified")
	}

	shooter := shot.NewShooter(t, region)
	if shooter == nil {
		return errors.New("cannot find the right shooter for lambda")
	}
	_, err := url.Parse(evt.Target)
	if err == nil {
		shooter.SetTarget(evt.Target)
		shooter.SetMethod(evt.Method)
		if evt.Body != nil {
			shooter.SetBody(evt.Body)
		}
		if evt.Header != nil {
			shooter.SetHeader(evt.Header)
		}
		shooter.SetRate(1)
		shooter.SetSlackURL(evt.SlackURLs)
		if err := shooter.Run(); err != nil {
			return err
		}
	}

	return nil
}
