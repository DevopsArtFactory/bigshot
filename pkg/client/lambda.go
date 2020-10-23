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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/config"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Lambda struct {
	Client *lambda.Lambda
}

// NewLambdaClient creates Lambda client
func NewLambdaClient(region string) *Lambda {
	session := GetAwsSession()
	return &Lambda{
		Client: GetLambdaClientFn(session, region, nil),
	}
}

// GetLambdaClientFn creates a new AWS lambda client
func GetLambdaClientFn(sess client.ConfigProvider, region string, creds *credentials.Credentials) *lambda.Lambda {
	if creds == nil {
		return lambda.New(sess, &aws.Config{Region: aws.String(region)})
	}
	return lambda.New(sess, &aws.Config{Region: aws.String(region), Credentials: creds})
}

// CreateFunction creates AWS lambda function
func (l Lambda) CreateFunction(config config.Config) (*string, error) {
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String(config.Name),
		Description:  aws.String(config.Description),
		Environment: &lambda.Environment{
			Variables: config.EnvironmentVariables,
		},
		Handler:    aws.String(config.Handler),
		MemorySize: aws.Int64(config.MemorySize),
		Publish:    aws.Bool(config.Publish),
		Runtime:    aws.String(config.Runtime),
		Timeout:    aws.Int64(config.Timeout),
		Tags:       config.Tags,
		Role:       config.Role,
	}

	if config.ZipFile != nil {
		input.Code = &lambda.FunctionCode{
			ZipFile: config.ZipFile,
		}
	} else {
		input.Code = &lambda.FunctionCode{
			S3Bucket: aws.String(constants.DefaultS3Bucket),
			S3Key:    aws.String(constants.DefaultS3Key),
		}
	}

	result, err := l.Client.CreateFunction(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == lambda.ErrCodeResourceConflictException {
				logrus.Warnf("Lambda function is already created: %s", config.Name)
				return nil, nil
			}
		}
		return nil, err
	}

	logrus.Infof("lambda function is newly created: %s", *result.FunctionArn)

	return result.FunctionArn, nil
}

// DeleteFunction deletes AWS lambda function
func (l Lambda) DeleteFunction(name string) error {
	input := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	}

	_, err := l.Client.DeleteFunction(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == lambda.ErrCodeResourceNotFoundException {
				logrus.Infof("Lambda function is already deleted: %s", name)
				return nil
			}
		}

		return err
	}

	logrus.Infof("Lambda function is successfully deleted: %s", name)

	return nil
}

// UpdateFunctionCode updates AWS lambda function code
func (l Lambda) UpdateFunctionCode(name string, zipFile []byte) error {
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(name),
	}

	if zipFile != nil {
		input.ZipFile = zipFile
	} else {
		input.S3Bucket = aws.String(constants.DefaultS3Bucket)
		input.S3Key = aws.String(constants.DefaultS3Key)
	}

	_, err := l.Client.UpdateFunctionCode(input)
	if err != nil {
		return err
	}

	logrus.Infof("Function code is successfully updated: %s", name)

	return nil
}

// UpdateTemplate updates AWS lambda function code
func (l Lambda) UpdateTemplate(config config.Config) error {
	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(config.Name),
		Environment: &lambda.Environment{
			Variables: config.EnvironmentVariables,
		},
	}

	_, err := l.Client.UpdateFunctionConfiguration(input)
	if err != nil {
		return err
	}

	logrus.Infof("Template of function is successfully updated: %s", config.Name)

	return nil
}

// Trigger will invoke lambda function
func (l *Lambda) Trigger(region string, payload []byte) error {
	functionName := tools.GenerateNewWorkerName(region, constants.WorkerMode)

	input := &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	}

	_, err := l.Client.Invoke(input)
	if err != nil {
		return err
	}

	logrus.Infof("Lambda is successfully invoked: %s", functionName)

	return nil
}
