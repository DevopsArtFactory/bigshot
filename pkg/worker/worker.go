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
	Region       string
	Mode         string
	RoleArn      *string
	ZipFile      []byte
	Template     string
	Error        error
	LambdaClient *client.Lambda
	IAMClient    *client.IAM
}

// New creates a new lambda for specific region
func New(region string, zipFile []byte) *Worker {
	w := Worker{
		Region:       region,
		RoleArn:      nil,
		Mode:         constants.WorkerMode,
		LambdaClient: client.NewLambdaClient(region),
		IAMClient:    client.NewIAMClient(region),
	}

	if zipFile != nil {
		w.ZipFile = zipFile
	}
	return &w
}

// GetRegion returns region of workers
func (w *Worker) GetRegion() string {
	return w.Region
}

// SetMode sets mode of lambda
func (w *Worker) SetMode(mode string) {
	logrus.Debugf("Mode is changed to %s", mode)
	w.Mode = mode
}

// SetTemplate sets template of controller
func (w *Worker) SetTemplate(template string) {
	logrus.Debugf("Template is set with %s", template)
	w.Template = template
}

// CreateWorkerRole creates lambdaRole
func (w *Worker) CreateWorkerRole() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region)

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

	return nil
}

// AttachWorkerRolePolicy attaches IAM Policy to IAM role
func (w *Worker) AttachWorkerRolePolicy() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region)

	err := w.IAMClient.AttachIAMPolicy(roleName)
	if err != nil {
		return err
	}

	roleArn, err := w.IAMClient.FindIamRoleForLambda(roleName)
	if err != nil {
		return err
	}

	w.RoleArn = roleArn

	return nil
}

// CreateWorker creates lambda
func (w *Worker) CreateWorker() error {
	workerConfig := config.GetBaseWorkerConfig(w.Region, w.Mode, w.Template, w.RoleArn, w.ZipFile)

	_, err := w.LambdaClient.CreateFunction(workerConfig)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWorkerRole deletes a lambdaRole
func (w *Worker) DeleteWorkerRole() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region)

	err := w.IAMClient.DeleteIamRoleForLambda(roleName)
	if err != nil {
		return err
	}

	return nil
}

// DetachWorkerRolePolicy detaches IAM Policy from IAM role
func (w *Worker) DetachWorkerRolePolicy() error {
	roleName := tools.GenerateNewLambdaRoleName(w.Region)

	err := w.IAMClient.DetachIAMPolicy(roleName)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWorker creates lambda
func (w *Worker) DeleteWorker() error {
	err := w.LambdaClient.DeleteFunction(tools.GenerateNewWorkerName(w.Region, w.Mode))
	if err != nil {
		return err
	}

	return nil
}

// UpdateWorkerCode updates lambda
func (w *Worker) UpdateWorkerCode() error {
	funcName := tools.GenerateNewWorkerName(w.Region, w.Mode)
	workerConfig := config.GetBaseWorkerConfig(w.Region, w.Mode, w.Template, w.RoleArn, w.ZipFile)

	if workerConfig.ZipFile != nil {
		if err := w.LambdaClient.UpdateFunctionCode(funcName, workerConfig.ZipFile); err != nil {
			return err
		}
	} else {
		logrus.Infof("function code is not updated: %s", w.Region)
	}

	return nil
}

// UpdateWorkerTemplate updates lambda
func (w *Worker) UpdateWorkerTemplate(c *schema.Config) error {
	workerConfig := config.GetBaseWorkerConfig(w.Region, w.Mode, w.Template, w.RoleArn, w.ZipFile)

	if err := w.LambdaClient.UpdateTemplate(workerConfig); err != nil {
		return err
	}

	con, err := controller.New(c)
	if err != nil {
		return err
	}

	// update the configuration
	tableName := tools.GenerateNewTableName()
	if err := con.DynamoDBClient.CreateItem(*con.Config, tableName); err != nil {
		return err
	}

	logrus.Info("Update template is done")

	return nil
}
