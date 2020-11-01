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

package runner

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/generator"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
	"github.com/DevopsArtFactory/bigshot/pkg/worker"
)

type Runner struct {
	Builder   *builder.Builder
	Generator *generator.Generator
}

// New creates a new Runner
func New(b *builder.Builder) *Runner {
	return &Runner{
		Builder: b,
	}
}

// Init creates global lambda functions for command line
func (r *Runner) Init() error {
	var wg sync.WaitGroup
	logrus.Info("Initiate bigshot infrastructures")

	if r.Builder.Config == nil {
		return errors.New("configuration is required for initialization")
	}

	if err := r.SetGenerator(); err != nil {
		return err
	}

	// Setup Controller
	if r.Generator.Controller != nil {
		err := r.Generator.Controller.Setup()
		if err != nil {
			return err
		}
	}

	openChannel := openChannel()
	defer close(r.Generator.Channel.Output)

	go openChannel(r.Generator.Channel, &wg)

	createIAMRole := makeCreateIAMRoleFunc()
	attachIAMPolicy := makeAttachIAMRolePolicyFunc()
	createLambdaWorker := makeCreateLambdaWorkerFunc()

	for _, w := range r.Generator.Workers {
		wg.Add(1)
		go createIAMRole(w, r.Generator.Channel.Input)
	}
	wg.Wait()

	tools.Wait(15, "Waiting %d seconds until IAM role is in effective...")

	for _, w := range r.Generator.Workers {
		wg.Add(1)
		go attachIAMPolicy(w, r.Generator.Channel.Input)
	}
	wg.Wait()

	for _, w := range r.Generator.Workers {
		if w.Error == nil {
			wg.Add(1)
			go createLambdaWorker(w, r.Generator.Channel.Input)
		}
	}
	wg.Wait()
	close(r.Generator.Channel.Input)

	PrintErrors(<-r.Generator.Channel.Output)

	return nil
}

// Destroy deletes all resources of bigshot
func (r *Runner) Destroy() error {
	var wg sync.WaitGroup
	logrus.Info("Destroying bigshot infrastructures")

	if err := r.SetGenerator(); err != nil {
		return err
	}

	openChannel := openChannel()
	defer close(r.Generator.Channel.Output)

	go openChannel(r.Generator.Channel, &wg)

	detachIAMPolicy := makeDetachIAMRolePolicyFunc()
	deleteIAMRole := makeDeleteIAMRoleFunc()
	deleteLambdaWorker := makeDeleteLambdaWorkerFunc()

	for _, w := range r.Generator.Workers {
		wg.Add(1)
		go detachIAMPolicy(w, r.Generator.Channel.Input)
	}
	wg.Wait()

	for _, w := range r.Generator.Workers {
		wg.Add(1)
		go deleteIAMRole(w, r.Generator.Channel.Input)
	}
	wg.Wait()

	for _, w := range r.Generator.Workers {
		if w.Error == nil {
			wg.Add(1)
		}
		go deleteLambdaWorker(w, r.Generator.Channel.Input)
	}
	wg.Wait()
	close(r.Generator.Channel.Input)

	errors := <-r.Generator.Channel.Output
	PrintErrors(errors)

	// Delete cloudwatch rules
	if err := r.Delete(); err != nil {
		return err
	}

	return nil
}

// UpdateCode updates global lambda function code
func (r *Runner) UpdateCode() error {
	var wg sync.WaitGroup
	logrus.Info("Update code of bigshot infrastructures")

	if err := r.SetGenerator(); err != nil {
		return err
	}

	openChannel := openChannel()
	defer close(r.Generator.Channel.Output)

	go openChannel(r.Generator.Channel, &wg)

	updateLambdaWorkerCode := makeUpdateLambdaWorkerCodeFunc()

	for _, w := range r.Generator.Workers {
		if w.Error == nil {
			wg.Add(1)
			go updateLambdaWorkerCode(w, r.Generator.Channel.Input)
		}
	}
	wg.Wait()
	close(r.Generator.Channel.Input)

	errors := <-r.Generator.Channel.Output
	PrintErrors(errors)

	return nil
}

// UpdateTemplate updates template of controller
func (r *Runner) UpdateTemplate(args []string) error {
	if len(args) != 1 && (len(args) == 0 && len(r.Builder.Flags.Config) == 0) {
		return errors.New("change name of template: bigshot update-template <template-name>\nchange template config: bigshot update-template --config=config.yaml")
	}

	logrus.Info("Update template of bigshot controller")

	if r.Builder.Config == nil {
		r.Builder.Config = &schema.Config{
			Name: args[0],
		}
	}

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}
	r.Builder.Flags.Region = region

	if err := r.SetGenerator(); err != nil {
		return err
	}

	// Setup Controller
	if r.Generator.Controller != nil {
		err := r.Generator.Controller.Setup()
		if err != nil {
			return err
		}
	}

	for _, w := range r.Generator.Workers {
		if w.Mode == constants.ControllerMode {
			if err := w.UpdateWorkerTemplate(r.Builder.Config); err != nil {
				return err
			}
		}
	}

	return nil
}

// Run creates a cloudwatch rule to start bigshot
func (r *Runner) Run() error {
	logrus.Info("Running bigshot infrastructures")

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	cw := client.NewCloudWatchClient(region)
	lambda := client.NewLambdaClient(region)

	min := r.Builder.Flags.Interval / 60
	cron := tools.CreateCronExpression(min)
	name := tools.GenerateRuleName(region)

	ruleArn, err := cw.PutRule(name, cron)
	if err != nil {
		return err
	}
	logrus.Infof("Cloudwatch rule is successfully created: %s", *ruleArn)

	funcName := tools.GenerateNewWorkerName(region, constants.ControllerMode)
	funcARN, err := lambda.GetFunctionARN(funcName)
	if err != nil {
		return err
	}
	logrus.Infof("Target lambda function is found %s", *funcARN)

	if err := cw.PutTarget(name, funcARN); err != nil {
		return err
	}
	logrus.Info("Attached the target function to the rule")

	if err := lambda.AddPermission(*ruleArn, funcName); err != nil {
		return err
	}
	logrus.Info("Permission is successfully granted")

	logrus.Infof("Bigshot will be run every %d minutes", min)

	return nil
}

// Stop stops a cloudwatch rule
func (r *Runner) Stop() error {
	logrus.Info("Stopping the bigshot rule")

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	cw := client.NewCloudWatchClient(region)
	rule := tools.GenerateRuleName(region)
	logrus.Infof("Trying to disable rule: %s", rule)
	if err := cw.DisableRule(rule); err != nil {
		return err
	}
	logrus.Infof("Bigshot is successfully stop running: %s", rule)

	return nil
}

// Delete removes a cloudwatch rule
func (r *Runner) Delete() error {
	logrus.Info("Deleting the bigshot rule")

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	cw := client.NewCloudWatchClient(region)
	lambda := client.NewLambdaClient(region)

	funcName := tools.GenerateNewWorkerName(region, constants.ControllerMode)
	if err := lambda.DeletePermission(funcName); err != nil {
		return err
	}
	logrus.Info("Permission is successfully removed")

	rule := tools.GenerateRuleName(region)
	targets, err := cw.ListTargetsByRule(rule)
	if err != nil {
		return err
	}
	logrus.Infof("Targets are found: %d", len(targets))

	if len(targets) > 0 {
		if err := cw.RemoveTarget(rule, targets); err != nil {
			return err
		}

		logrus.Infof("Trying to delete the rule: %s", rule)
	}

	if err := cw.DeleteRule(rule); err != nil {
		return err
	}
	logrus.Infof("Rule is successfully deleted: %s", rule)

	return nil
}

// SetGenerator setup a new Generator
func (r *Runner) SetGenerator() error {
	gn := generator.New()

	err := gn.Init(r.Builder.Flags, r.Builder.Config)
	if err != nil {
		return err
	}

	r.Generator = gn

	return nil
}

// openChannel opens channel with input
func openChannel() func(*generator.Channel, *sync.WaitGroup) {
	var result []error
	return func(ch *generator.Channel, wg *sync.WaitGroup) {
		for re := range ch.Input {
			result = append(result, re)
			wg.Done()
		}

		ch.Output <- result
	}
}

// makeCreateIAMRoleFunc creates a go routine function for creating IAM Role
func makeCreateIAMRoleFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		if err := w.CreateWorkerRole(); err != nil {
			w.Error = err
			ch <- err
			return
		}
		logrus.Infof("IAM role for lambda is ready in %s", w.GetRegion())
		ch <- nil
	}
}

// makeAttachIAMRolePolicyFunc creates a go routine function for attaching IAM Role policy
func makeAttachIAMRolePolicyFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		if err := w.AttachWorkerRolePolicy(); err != nil {
			w.Error = err
			ch <- err
			return
		}
		logrus.Infof("IAM role policy is successfully attached in %s", w.GetRegion())
		ch <- nil
	}
}

// makeCreateLambdaWorkerFunc creates a go routine function for creating Lambda lambda
func makeCreateLambdaWorkerFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		if err := w.CreateWorker(); err != nil {
			w.Error = err
			ch <- err
		}
		logrus.Infof("Worker function is ready in %s", w.GetRegion())
		ch <- nil
	}
}

// makeDeleteIAMRoleFunc creates a go routine function for deleting IAM Role
func makeDeleteIAMRoleFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		err := w.DeleteWorkerRole()
		if err != nil {
			w.Error = err
		}
		ch <- err
	}
}

// makeDetachIAMRolePolicyFunc creates a go routine function for detaching IAM Role policy
func makeDetachIAMRolePolicyFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		if err := w.DetachWorkerRolePolicy(); err != nil {
			w.Error = err
			ch <- err
			return
		}
		logrus.Infof("IAM role policy is successfully detached in %s", w.GetRegion())
		ch <- nil
	}
}

// makeDeleteLambdaWorkerFunc creates a go routine function for creating Lambda lambda
func makeDeleteLambdaWorkerFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		err := w.DeleteWorker()
		if err != nil {
			w.Error = err
		}
		ch <- err
	}
}

// makeUpdateLambdaWorkerCodeFunc create a go routine function for updating Lambda lambda
func makeUpdateLambdaWorkerCodeFunc() func(w *worker.Worker, ch chan error) {
	return func(w *worker.Worker, ch chan error) {
		if err := w.UpdateWorkerCode(); err != nil {
			w.Error = err
			ch <- err
		}
		logrus.Infof("Worker update is done in %s", w.GetRegion())
		ch <- nil
	}
}

// PrintErrors prints errors
func PrintErrors(errors []error) {
	for _, err := range errors {
		if err != nil {
			logrus.Errorf(err.Error())
		}
	}
}
