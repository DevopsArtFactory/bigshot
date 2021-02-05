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
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/config"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/controller"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Worker struct {
	Mode         string
	RoleArn      *string
	ZipFile      []byte
	Template     string
	DryRun       bool
	Internal     bool
	Error        error
	Region       *schema.Region
	Config       *schema.Template
	LambdaClient *client.Lambda
	IAMClient    *client.IAM
}

// New creates a new lambda for specific region
func New(region string, zipFile []byte, config *schema.Template, dryRun, internal bool) *Worker {
	w := Worker{
		Region:       getWorkerRegion(region, config.Regions),
		RoleArn:      nil,
		Mode:         constants.WorkerMode,
		Config:       config,
		Template:     *config.Name,
		DryRun:       dryRun,
		Internal:     internal,
		LambdaClient: client.NewLambdaClient(region),
		IAMClient:    client.NewIAMClient(region),
	}

	if zipFile != nil {
		w.ZipFile = zipFile
	}
	return &w
}

// getWorkerRegion get worker region configuration from Regions configuration
func getWorkerRegion(region string, regions []schema.Region) *schema.Region {
	for _, reg := range regions {
		if *reg.Region == region {
			return &reg
		}
	}

	return &schema.Region{Region: &region}
}

// GetRegion returns region of workers
func (w *Worker) GetRegion() string {
	return *w.Region.Region
}

// SetMode sets mode of lambda
func (w *Worker) SetMode(mode string) {
	logrus.Debugf("Mode is changed to %s", mode)
	w.Mode = mode
}

// SetTemplate sets template of workermanager
func (w *Worker) SetTemplate(template string) {
	logrus.Debugf("Template is set with %s", template)
	w.Template = template
}

// CreateWorkerRole creates lambdaRole
func (w *Worker) CreateWorkerRole() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region.Region, w.Config.Name)

	if w.DryRun {
		logrus.Debugf("[V] IAM role created: %s, %s", *w.Region.Region, roleName)
		return nil
	}

	roleArn, err := w.IAMClient.FindIamRoleForLambda(roleName)
	if err != nil {
		return err
	}

	if roleArn == nil {
		_, err = w.IAMClient.CreateIamRoleForLambda(roleName)
		if err != nil {
			return err
		}
	}

	logrus.Debugf("IAM role for lambda is ready in %s", w.GetRegion())

	return nil
}

// AttachWorkerRolePolicy attaches IAM Policy to IAM role
func (w *Worker) AttachWorkerRolePolicy() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region.Region, w.Config.Name)

	if w.DryRun {
		logrus.Debugf("[%s]Lambda STS Policy will be attached to the role: %s", *w.Region.Region, roleName)
		return nil
	}

	err := w.IAMClient.AttachIAMPolicy(roleName)
	if err != nil {
		return err
	}

	roleArn, err := w.IAMClient.FindIamRoleForLambda(roleName)
	if err != nil {
		return err
	}

	w.RoleArn = roleArn

	logrus.Debugf("IAM role policy is successfully attached in %s", w.GetRegion())

	return nil
}

// CreateWorker creates lambda
func (w *Worker) CreateWorker() error {
	workerConfig := GetBaseWorkerConfig(w)

	if w.DryRun {
		logrus.Debugf("[%s]Lambda worker will be created: %s", *w.Region.Region, workerConfig.Name)
		return nil
	}

	_, err := w.LambdaClient.CreateFunction(workerConfig)
	if err != nil {
		return err
	}

	logrus.Debugf("Worker function is ready in %s", w.GetRegion())

	return nil
}

// DeleteWorkerRole deletes a lambdaRole
func (w *Worker) DeleteWorkerRole() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region.Region, w.Config.Name)

	err := w.IAMClient.DeleteIamRoleForLambda(roleName)
	if err != nil {
		return err
	}

	return nil
}

// DetachWorkerRolePolicy detaches IAM Policy from IAM role
func (w *Worker) DetachWorkerRolePolicy() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region.Region, w.Config.Name)

	err := w.IAMClient.DetachIAMPolicy(roleName)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWorker creates lambda
func (w *Worker) DeleteWorker() error {
	err := w.LambdaClient.DeleteFunction(tools.GenerateNewWorkerName(w.Region.Region, w.Config.Name, w.Mode, w.Internal))
	if err != nil {
		return err
	}

	return nil
}

// UpdateWorkerCode updates lambda
func (w *Worker) UpdateWorkerCode() error {
	funcName := tools.GenerateNewWorkerName(w.Region.Region, w.Config.Name, w.Mode, w.Internal)
	workerConfig := GetBaseWorkerConfig(w)

	if workerConfig.ZipFile != nil {
		if err := w.LambdaClient.UpdateFunctionCode(funcName, workerConfig.ZipFile); err != nil {
			return err
		}
	} else {
		logrus.Debugf("function code is not updated: %s", *w.Region.Region)
	}

	return nil
}

// UpdateWorkerTemplate updates lambda
func (w *Worker) UpdateWorkerTemplate(c *schema.Template) error {
	workerConfig := GetBaseWorkerConfig(w)

	if err := w.LambdaClient.UpdateTemplate(workerConfig); err != nil {
		return err
	}

	con, err := controller.New(c, *w.Region.Region)
	if err != nil {
		return err
	}

	// update the configuration
	tableName := tools.GenerateNewTableName()
	if err := con.DynamoDBClient.SaveItem(*con.Template, tableName); err != nil {
		return err
	}

	logrus.Debugln("Update template is done")

	return nil
}

// GetBaseWorkerConfig returns base lambda configuration
func GetBaseWorkerConfig(worker *Worker) config.Config {
	env := config.GetEnvironmentVariables(worker.Region.Region, worker.Mode, worker.Config.Name)
	cf := config.Config{
		Name:                 tools.GenerateNewWorkerName(worker.Region.Region, worker.Config.Name, worker.Mode, worker.Internal),
		Description:          tools.GenerateDescription(worker.Region.Region),
		EnvironmentVariables: env,
		Handler:              "out/code/lambda/handler",
		MemorySize:           int64(256),
		Tags:                 config.GetTags(worker.Region.Region, worker.Mode),
		Role:                 worker.RoleArn,
		Publish:              true,
		Internal:             worker.Internal,
		Runtime:              constants.GoRunTime,
	}

	if worker.Config.Timeout != nil {
		cf.Timeout = int64(*worker.Config.Timeout)
	}

	if worker.ZipFile != nil {
		cf.ZipFile = worker.ZipFile
	} else {
		cf.S3Bucket = constants.DefaultS3Bucket
		cf.S3Key = constants.DefaultS3Key
	}

	if worker.Internal {
		cf.SecurityGroups = worker.Region.SecurityGroups
		cf.Subnets = worker.Region.Subnets
	}

	return cf
}
