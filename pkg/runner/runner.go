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
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/color"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/generator"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/server"
	"github.com/DevopsArtFactory/bigshot/pkg/templates"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
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
	logrus.Info("Initiates bigshot infrastructures")

	if r.Builder.Config == nil {
		return errors.New("configuration is required for initialization")
	}

	// setup generator
	if err := r.SetGenerator(); err != nil {
		return err
	}

	// setup controller metadata table
	if r.Generator.Controller != nil {
		err := r.Generator.Controller.SetupMetadataTable()
		if err != nil {
			return err
		}
	}

	for _, w := range r.Generator.Workers {
		if err := w.CreateWorkerRole(); err != nil {
			fmt.Println(err.Error())
		}
	}

	if !r.Builder.Flags.DryRun {
		tools.Wait(15, "Waiting %d seconds until IAM role is in effective...")
	}

	for _, w := range r.Generator.Workers {
		if err := w.AttachWorkerRolePolicy(); err != nil {
			fmt.Println(err.Error())
		}
	}

	for _, w := range r.Generator.Workers {
		if err := w.CreateWorker(); err != nil {
			fmt.Println(err.Error())
		}
	}

	if err := r.CreateTrigger(r.Builder.Config); err != nil {
		return err
	}
	return nil
}

// Destroy deletes all resources of bigshot
func (r *Runner) Destroy(args []string) error {
	name, err := r.GetTargetFunctionName(args)
	if err != nil {
		return err
	}
	logrus.Infof("Destroying bigshot infrastructures: %s", name)

	// name update
	r.OverrideName(&name)

	if err := r.SetGenerator(); err != nil {
		return err
	}

	for _, w := range r.Generator.Workers {
		if err := w.DetachWorkerRolePolicy(); err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	for _, w := range r.Generator.Workers {
		if err := w.DeleteWorkerRole(); err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	for _, w := range r.Generator.Workers {
		if err := w.DeleteWorker(); err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	// Delete cloudwatch rules
	if err := r.Delete([]string{name}); err != nil {
		return err
	}

	return nil
}

// UpdateCode updates global lambda function code
func (r *Runner) UpdateCode(args []string) error {
	name, err := r.GetTargetFunctionName(args)
	if err != nil {
		return err
	}
	logrus.Infof("Update code of bigshot worker and manager: %s", name)

	r.OverrideName(&name)

	if err := r.SetGenerator(); err != nil {
		return err
	}

	for _, w := range r.Generator.Workers {
		if err := w.UpdateWorkerCode(); err != nil {
			fmt.Println(err.Error())
		}
	}

	return nil
}

// UpdateTemplate updates template of workermanager
func (r *Runner) UpdateTemplate(args []string) error {
	if len(args) != 1 && (len(args) == 0 && len(r.Builder.Flags.Config) == 0) {
		return errors.New("change name of template: bigshot update-template <template-name>\nchange template config: bigshot update-template --config=config.yaml")
	}

	logrus.Info("Update template of bigshot workermanager")

	if r.Builder.Config == nil {
		r.Builder.Config = &schema.Template{
			Name: &args[0],
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
		err := r.Generator.Controller.SetupMetadataTable()
		if err != nil {
			return err
		}
	}

	for _, w := range r.Generator.Workers {
		if w.Mode == constants.ManagerMode {
			if err := w.UpdateWorkerTemplate(r.Builder.Config); err != nil {
				return err
			}
		}
	}

	return nil
}

// Run triggers a single worker manager
func (r *Runner) Run(args []string) error {
	name, err := r.GetTargetFunctionName(args)
	if err != nil {
		return err
	}
	logrus.Infof("Running bigshot infrastructures: %s", name)

	return nil
}

// CreateTrigger creates a cloudwatch rule to start bigshot
func (r *Runner) CreateTrigger(config *schema.Template) error {
	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	cw := client.NewCloudWatchClient(region)
	lambda := client.NewLambdaClient(region)

	min := *config.Interval / 60
	cron := tools.CreateCronExpression(min)
	logrus.Infof("cron expression made: %s", cron)
	ruleName := tools.GenerateRuleName(region, *config.Name)

	ruleArn, err := cw.PutRule(ruleName, cron)
	if err != nil {
		return err
	}
	logrus.Infof("Cloudwatch rule is successfully created: %s", *ruleArn)

	funcName := tools.GenerateNewWorkerName(&region, config.Name, constants.ManagerMode, false)
	funcARN, err := lambda.GetFunctionARN(funcName)
	if err != nil {
		return err
	}
	logrus.Infof("Target lambda function is found %s", *funcARN)

	if err := cw.PutTarget(ruleName, funcARN); err != nil {
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
func (r *Runner) Stop(args []string) error {
	name, err := r.GetTargetFunctionName(args)
	if err != nil {
		return err
	}
	logrus.Infof("Stopping the bigshot rule: %s", name)

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	cw := client.NewCloudWatchClient(region)
	rule := tools.GenerateRuleName(region, name)
	logrus.Infof("Trying to disable rule: %s", rule)
	if err := cw.DisableRule(rule); err != nil {
		return err
	}
	logrus.Infof("Bigshot is successfully stop running: %s", rule)

	return nil
}

// Delete removes a cloudwatch rule
func (r *Runner) Delete(args []string) error {
	name, err := r.GetTargetFunctionName(args)
	if err != nil {
		return err
	}

	if len(name) == 0 {
		return errors.New("please choose or specify the worker name")
	}
	logrus.Debugf("Deleting the bigshot rule: %s", name)

	if err := r.SetGenerator(); err != nil {
		return err
	}

	r.OverrideName(&name)

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	cw := client.NewCloudWatchClient(region)
	lambda := client.NewLambdaClient(region)

	funcName := tools.GenerateNewWorkerName(&region, &name, constants.ManagerMode, false)
	if err := lambda.DeletePermission(funcName); err != nil {
		return err
	}
	logrus.Info("Permission is successfully removed")

	rule := tools.GenerateRuleName(region, name)
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

// List shows the worker status
func (r *Runner) List() error {
	names, err := r.FindAllNames()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Println("No worker exists")
		return nil
	}

	if err := r.PrintSummaryTemplates(names); err != nil {
		return err
	}

	return nil
}

// RunServer runs bigshot workermanager as server
func (r *Runner) RunServer() error {
	logrus.Infof("Booting up bigshot server")
	s := server.New()
	s.SetRouter()

	if err := s.SetDefaultSetting(r.Builder.Flags.LogFile); err != nil {
		return err
	}

	logrus.Infof("Server setting is done")

	addr := s.GetAddr()
	logrus.Infof("Start bigshot server")
	if err := http.ListenAndServe(addr, server.Wrapper(s.Router)); err != nil {
		logrus.Errorf(err.Error())
	}
	logrus.Infof("Shutting down bigshot server")

	return nil
}

// PrintSummary prints summary of template
func (r *Runner) PrintSummaryTemplates(names []string) error {
	dynamoDB := client.NewDynamoDBClient(r.Builder.DefaultRegion)
	var configs []*schema.Template
	for _, name := range names {
		item, err := dynamoDB.GetTemplate(name, tools.GenerateNewTableName())
		if err != nil {
			return err
		}

		conf, err := mapConfig(item)
		if err != nil {
			return err
		}

		configs = append(configs, conf)
	}

	if err := PrintTemplate(configs); err != nil {
		return err
	}

	return nil
}

// PrintTemplate prints the summary of templates
func PrintTemplate(configs []*schema.Template) error {
	var data = struct {
		Summary []*schema.Template
	}{
		Summary: configs,
	}

	funcMap := template.FuncMap{
		"decorate": color.DecorateAttr,
		"format":   tools.Formatting,
		"join":     tools.JoinString,
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 5, 3, ' ', tabwriter.TabIndent)
	t := template.Must(template.New("Template Information").Funcs(funcMap).Parse(templates.ListTemplate))

	err := t.Execute(w, data)
	if err != nil {
		return err
	}

	return w.Flush()
}

// mapConfig maps configuration with dynamodb item
func mapConfig(item map[string]*dynamodb.AttributeValue) (*schema.Template, error) {
	conf := schema.Template{}
	for k, v := range item {
		switch k {
		case "name":
			conf.Name = v.S
		case "interval":
			interval, _ := strconv.Atoi(*v.N)
			conf.Interval = &interval
		case "regions":
			var regions []schema.Region
			for _, region := range v.L {
				regions = append(regions, regionConfig(*region.M["region"].S))
			}
			conf.Regions = regions
		case "slack_urls":
			var slackURLs []string
			for _, url := range v.SS {
				slackURLs = append(slackURLs, *url)
			}
			conf.SlackURLs = slackURLs
		case "targets":
			var targets []schema.Target
			for _, target := range v.L {
				targets = append(targets, targetConfig(*target.M["method"].S, *target.M["url"].S))
			}
			conf.Targets = targets
		case "timeout":
			timeout, err := strconv.Atoi(*v.N)
			if err != nil {
				return nil, err
			}

			conf.Timeout = &timeout
		}
	}

	return &conf, nil
}

// regionConfig creates region config struct
func regionConfig(region string) schema.Region {
	return schema.Region{
		Region: &region,
	}
}

// targetConfig creates target config struct
func targetConfig(method, url string) schema.Target {
	return schema.Target{
		Method: &method,
		URL:    &url,
	}
}

// SetGenerator setup a new Generator
func (r *Runner) SetGenerator() error {
	r.Generator = generator.New()

	checkDryRun(r.Builder.Flags.DryRun)

	err := r.Generator.Init(r.Builder.Flags, r.Builder.Config)
	if err != nil {
		return err
	}

	return nil
}

// checkDryRun will check if this command is dry-run or not
func checkDryRun(dryRun bool) {
	if dryRun {
		logrus.Info("Dry run mode enabled")
	}
}

// GetTargetFunctionName returns target function name
func (r *Runner) GetTargetFunctionName(args []string) (string, error) {
	if len(args) > 1 {
		return constants.EmptyString, errors.New("only one argument is required")
	}

	if len(args) == 1 {
		return args[0], nil
	}

	names, err := r.FindAllNames()
	if err != nil {
		return constants.EmptyString, err
	}

	target := constants.EmptyString
	prompt := &survey.Select{
		Message: "Pick worker name:",
		Options: names,
	}
	survey.AskOne(prompt, &target)

	if len(target) == 0 {
		return constants.EmptyString, errors.New("please choose or specify the worker name")
	}

	return target, nil
}

// FindAllNames means finding names in slice of string
func (r *Runner) FindAllNames() ([]string, error) {
	tableName := tools.GenerateNewTableName()
	dynamodb := client.NewDynamoDBClient(r.Builder.DefaultRegion)
	names, err := dynamodb.GetAllNames(tableName)
	if err != nil {
		return nil, err
	}

	return names, nil
}

// OverrideName will override the configuration name
func (r *Runner) OverrideName(name *string) {
	if r.Builder.Config == nil {
		r.Builder.Config = &schema.Template{
			Name: name,
		}

		return
	}
	r.Builder.Config.Name = name
}
