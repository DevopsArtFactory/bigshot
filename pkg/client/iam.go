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
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sirupsen/logrus"
)

type IAM struct {
	Client *iam.IAM
}

// NewIAMClient creates IAM client
func NewIAMClient(region string) *IAM {
	session := GetAwsSession()
	return &IAM{
		Client: GetIAMClientFn(session, region, nil),
	}
}

// GetIAMClientFn creates a new AWS IAM client
func GetIAMClientFn(sess client.ConfigProvider, region string, creds *credentials.Credentials) *iam.IAM {
	if creds == nil {
		return iam.New(sess, &aws.Config{Region: aws.String(region)})
	}
	return iam.New(sess, &aws.Config{Region: aws.String(region), Credentials: creds})
}

// CreateIamRoleForLambda creates a lambda role
func (i IAM) CreateIamRoleForLambda(name string) (*string, error) {
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String("{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Sid\": \"\",\n      \"Effect\": \"Allow\",\n      \"Principal\": {\n        \"Service\": \"lambda.amazonaws.com\"\n      },\n      \"Action\": \"sts:AssumeRole\"\n    }\n  ]\n}"),
		Path:                     aws.String("/"),
		RoleName:                 aws.String(name),
	}

	result, err := i.Client.CreateRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == iam.ErrCodeEntityAlreadyExistsException {
				logrus.Debugf("Role is already created")
				return nil, nil
			}
		}
		return nil, err
	}

	logrus.Infof("New lambda role is created: %s", *result.Role.Arn)

	return result.Role.Arn, nil
}

// AttachIAMPolicy attaches IAM Policy to role
func (i IAM) AttachIAMPolicy(name string) error {
	input := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
		RoleName:  aws.String(name),
	}

	_, err := i.Client.AttachRolePolicy(input)
	if err != nil {
		return err
	}

	logrus.Infof("Policy is attached to IAM role: %s", name)

	return nil
}

// FindIamRoleForLambda finds a lambda role
func (i IAM) FindIamRoleForLambda(name string) (*string, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	result, err := i.Client.GetRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == iam.ErrCodeNoSuchEntityException {
				logrus.Debugf("Role is not found")
				return nil, nil
			}
		}
		return nil, err
	}

	logrus.Infof("IAM role for lambda function found: %s", *result.Role.Arn)

	return result.Role.Arn, nil
}

// DeleteIamRoleForLambda deletes a lambda role
func (i IAM) DeleteIamRoleForLambda(name string) error {
	input := &iam.DeleteRoleInput{
		RoleName: aws.String(name),
	}

	_, err := i.Client.DeleteRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == iam.ErrCodeNoSuchEntityException {
				logrus.Infof("Lambda role is already deleted: %s", name)
				return nil
			}
		}
		return err
	}

	logrus.Infof("Lambda role is successfully deleted: %s", name)

	return nil
}

// DetachIAMPolicy detaches IAM Policy from role
func (i IAM) DetachIAMPolicy(name string) error {
	input := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
		RoleName:  aws.String(name),
	}

	_, err := i.Client.DetachRolePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == iam.ErrCodeNoSuchEntityException {
				logrus.Debugf("Cannot find role for deleting policy: %s", name)
				return nil
			}
		}
		return err
	}

	logrus.Infof("Policy is detached from IAM role: %s", name)

	return nil
}
