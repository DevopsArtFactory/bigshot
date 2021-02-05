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

package client

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type DynamoDB struct {
	Client *dynamodb.DynamoDB
}

// NewDynamoDBClient creates Dynamodb clienth
func NewDynamoDBClient(region string) *DynamoDB {
	session := GetAwsSession()
	return &DynamoDB{
		Client: GetDynamoDBClientFn(session, region, nil),
	}
}

// GetDynamoDBClientFn creates a new AWS dynamodb client
func GetDynamoDBClientFn(sess client.ConfigProvider, region string, creds *credentials.Credentials) *dynamodb.DynamoDB {
	if creds == nil {
		return dynamodb.New(sess, &aws.Config{Region: aws.String(region)})
	}
	return dynamodb.New(sess, &aws.Config{Region: aws.String(region), Credentials: creds})
}

// CreateMetaDataTable creates a metadata table
func (d *DynamoDB) CreateMetaDataTable(name string) error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(constants.DefaultPrimaryKey),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(constants.DefaultPrimaryKey),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(constants.DefaultReadCapacity),
			WriteCapacityUnits: aws.Int64(constants.DefaultWriteCapacity),
		},
		TableName: aws.String(name),
	}

	logrus.Debugf("Metadata table creation is in progress: %s", name)
	_, err := d.Client.CreateTable(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceInUseException {
				logrus.Debugf("Metadata table is already created: %s", name)
				return nil
			}
		}
		return err
	}

	tools.Wait(10, "Wait %d seconds until table creation is done..")
	logrus.Debugf("Metadata table is successfully created: %s", name)

	return nil
}

// SaveItem creates new item for bigshot
func (d *DynamoDB) SaveItem(config schema.Template, tableName string) error {
	item, err := dynamodbattribute.MarshalMap(config)
	if err != nil {
		return fmt.Errorf("failed to DynamoDB marshal Record, %v", err)
	}

	input := &dynamodb.PutItemInput{
		Item:                   item,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(tableName),
	}

	_, err = d.Client.PutItem(input)
	if err != nil {
		return err
	}

	logrus.Debugf("Item is successfully created: %s", *config.Name)

	return nil
}

// GetTemplate retrieves template configurations from the metadata table
func (d *DynamoDB) GetTemplate(name, tableName string) (map[string]*dynamodb.AttributeValue, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			constants.DefaultPrimaryKey: {
				S: aws.String(name),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := d.Client.GetItem(input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("template does not exist: %s", name)
	}

	return result.Item, nil
}

// GetAllNames retrieves all names of bigshot template
func (d *DynamoDB) GetAllNames(table string) ([]string, error) {
	items, err := d.Scan(table)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, item := range items {
		names = append(names, *item["name"].S)
	}

	return names, nil
}

// Scan retrieves all dynamoDB data
func (d *DynamoDB) Scan(tableName string) ([]map[string]*dynamodb.AttributeValue, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := d.Client.Scan(input)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
