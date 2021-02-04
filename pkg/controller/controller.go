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
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/shot"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Controller struct {
	Template       *schema.Template
	DynamoDBClient *client.DynamoDB
	Region         string
}

// GetRegion returns region of controller
func (c *Controller) GetRegion() string {
	return c.Region
}

// New returns new controller
func New(config *schema.Template, region string) (*Controller, error) {
	var err error
	if len(region) == 0 {
		region, err = builder.GetDefaultRegion(constants.DefaultProfile)
		if err != nil {
			return nil, err
		}
	}
	return &Controller{
		Template:       config,
		Region:         region,
		DynamoDBClient: client.NewDynamoDBClient(region),
	}, nil
}

// SetupMetadataTable creates a controller metadata table
func (c *Controller) SetupMetadataTable() error {
	if c.Template == nil {
		logrus.Warn("Controller cannot be setup because there is no configuration file")
		return nil
	}

	// create dynamodb
	tableName := tools.GenerateNewTableName()
	if err := c.DynamoDBClient.CreateMetaDataTable(tableName); err != nil {
		return err
	}

	// create new configuration
	if err := c.DynamoDBClient.SaveItem(*c.Template, tableName); err != nil {
		return err
	}

	logrus.Debug("Controller setup is finished")
	return nil
}

// Run starts the synthetic with template
func Run(template string) error {
	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	dynamoDB := client.NewDynamoDBClient(region)

	item, err := dynamoDB.GetTemplate(template, tools.GenerateNewTableName())
	if err != nil {
		return err
	}

	logrus.Info(item)

	return nil
}

// ListItems retrieves all templates
func ListItems() ([]schema.Template, error) {
	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return nil, err
	}

	dynamoDB := client.NewDynamoDBClient(region)

	items, err := dynamoDB.Scan(tools.GenerateNewTableName())
	if err != nil {
		return nil, err
	}

	return ParseTemplates(items)
}

// GetDetail retrieves detailed information
func GetDetail(template string) (*schema.Template, error) {
	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return nil, err
	}

	dynamoDB := client.NewDynamoDBClient(region)

	item, err := dynamoDB.GetTemplate(template, tools.GenerateNewTableName())
	if err != nil {
		return nil, err
	}

	return ChangeItemToConfig(item)
}

// ModifyTemplate modifies template only
func ModifyTemplate(config schema.Template) error {
	tableName := tools.GenerateNewTableName()

	region, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return err
	}

	dynamoDB := client.NewDynamoDBClient(region)
	if err := dynamoDB.SaveItem(config, tableName); err != nil {
		return err
	}

	return nil
}

// RunTargetVerification
func RunTargetVerification(target schema.Target) (*schema.Result, error) {
	result, err := shot.Shoot(target, true)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)

	return result, nil
}
