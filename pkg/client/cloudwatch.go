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
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/sirupsen/logrus"
)

type CloudWatch struct {
	EventClient *cloudwatchevents.CloudWatchEvents
}

// NewCloudWatchClient creates CloudWatch client
func NewCloudWatchClient(region string) *CloudWatch {
	session := GetAwsSession()
	return &CloudWatch{
		EventClient: GetCloudWatchEventClientFn(session, region, nil),
	}
}

// GetCloudWatchEventClientFn creates a new AWS cloudwatch client
func GetCloudWatchEventClientFn(sess client.ConfigProvider, region string, creds *credentials.Credentials) *cloudwatchevents.CloudWatchEvents {
	if creds == nil {
		return cloudwatchevents.New(sess, &aws.Config{Region: aws.String(region)})
	}
	return cloudwatchevents.New(sess, &aws.Config{Region: aws.String(region), Credentials: creds})
}

// PutRule creates or update the rule
func (c *CloudWatch) PutRule(name, cron string) (*string, error) {
	input := &cloudwatchevents.PutRuleInput{
		Description:        aws.String(fmt.Sprintf("BigShot cloudwatch rule: %s", name)),
		Name:               aws.String(name),
		ScheduleExpression: aws.String(cron),
	}

	result, err := c.EventClient.PutRule(input)
	if err != nil {
		return nil, err
	}

	return result.RuleArn, nil
}

// PutTarget creates a target for the rule
func (c *CloudWatch) PutTarget(ruleName string, lambdaArn *string) error {
	input := &cloudwatchevents.PutTargetsInput{
		Rule: aws.String(ruleName),
		Targets: []*cloudwatchevents.Target{
			{
				Arn: lambdaArn,
				Id:  aws.String("1"),
			},
		},
	}

	_, err := c.EventClient.PutTargets(input)
	if err != nil {
		return err
	}

	return nil
}

// DisableRule disables a cloudwatch rule
func (c *CloudWatch) DisableRule(name string) error {
	input := &cloudwatchevents.DisableRuleInput{
		Name: aws.String(name),
	}

	_, err := c.EventClient.DisableRule(input)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRule deletes a cloudwatch rule
func (c *CloudWatch) DeleteRule(name string) error {
	input := &cloudwatchevents.DeleteRuleInput{
		Name:  aws.String(name),
		Force: aws.Bool(true),
	}

	_, err := c.EventClient.DeleteRule(input)
	if err != nil {
		return err
	}

	return nil
}

// RemoveTarget deletes a target from the rule
func (c *CloudWatch) RemoveTarget(ruleName string, targets []*cloudwatchevents.Target) error {
	ids := []*string{}
	for _, target := range targets {
		ids = append(ids, target.Id)
	}

	input := &cloudwatchevents.RemoveTargetsInput{
		Rule: aws.String(ruleName),
		Ids:  ids,
	}

	_, err := c.EventClient.RemoveTargets(input)
	if err != nil {
		return err
	}

	return nil
}

// ListTargetsByRule gets list of  targets by the rule
func (c *CloudWatch) ListTargetsByRule(ruleName string) ([]*cloudwatchevents.Target, error) {
	input := &cloudwatchevents.ListTargetsByRuleInput{
		Rule: aws.String(ruleName),
	}

	result, err := c.EventClient.ListTargetsByRule(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == cloudwatchevents.ErrCodeResourceNotFoundException {
				logrus.Warn(aerr.Error())
				return nil, nil
			}
		}
		return nil, err
	}

	return result.Targets, nil
}
