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
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Controller struct {
	Config         *schema.Config
	DynamoDBClient *client.DynamoDB
	Region         string
}

// GetRegion returns region of controller
func (c *Controller) GetRegion() string {
	return c.Region
}

// New returns new controller
func New(config *schema.Config) (*Controller, error) {
	defaultRegion, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return nil, err
	}

	return &Controller{
		Config:         config,
		Region:         defaultRegion,
		DynamoDBClient: client.NewDynamoDBClient(defaultRegion),
	}, nil
}

// Setup creates a controller function and metadata table
func (c *Controller) Setup() error {
	if c.Config == nil {
		logrus.Warn("Controller cannot be setup because there is no configuration file")
		return nil
	}

	// create dynamodb
	tableName := tools.GenerateNewTableName()
	if err := c.DynamoDBClient.CreateMetaDataTable(tableName); err != nil {
		return err
	}

	// create new configuration
	if err := c.DynamoDBClient.CreateItem(*c.Config, tableName); err != nil {
		return err
	}

	logrus.Debug("Controller setup is finished")
	return nil
}
