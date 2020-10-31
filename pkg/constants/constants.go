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

package constants

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultLogLevel is the default global verbosity
	DefaultLogLevel = logrus.InfoLevel

	// DefaultRegion is the default region id
	DefaultRegion = "us-east-1"

	// EmptyString is the empty string
	EmptyString = ""

	// DefaultRegionVariable is the default region id
	DefaultRegionVariable = "AWS_DEFAULT_REGION"

	// StringText is "string"
	StringText = "string"

	// GoRunTime means golang runtime
	GoRunTime = "go1.x"

	// DefaultS3Bucket indicates default s3 bucket for lambda function
	DefaultS3Bucket = "devopsartfactory"

	// DefaultS3Key indicates default s3 key for lambda function
	DefaultS3Key = "bigshot/code/lambda.zip"

	// DefaultShooter means default shooter
	DefaultShooter = "trace"

	// DefaultWorkerDuration means default duration of lambda
	DefaultWorkerDuration = 1 * time.Second

	// CommonNamePrefix means prefix of common resource
	CommonNamePrefix = "bigshot"

	// WorkerNamePrefix means prefix of lambda resources
	WorkerNamePrefix = "bigshot-lambda"

	// ControllerNamePrefix means prefix of controller resource
	ControllerNamePrefix = "bigshot-controller"

	// DefaultPrimaryKey indicates default primary key of bigshot dynamodb table
	DefaultPrimaryKey = "name"

	// DefaultReadCapacity means default value of read capacity
	DefaultReadCapacity = int64(5)

	// DefaultWriteCapacity means default value of write capacity
	DefaultWriteCapacity = int64(5)

	// DefaultProfile indicates default profile
	DefaultProfile = "default"

	// WorkerMode indicates lambda mode
	WorkerMode = "lambda"

	// ControllerMode indicates controller mode
	ControllerMode = "controller"

	// HTTPS means https protocol
	HTTPS = "HTTPS"

	// HTTP means http protocol
	HTTP = "HTTP"

	// BigShotSlackURLs
	BigShotSlackURLs = "slack_urls"
)

var (
	// AllAWSRegions is a list of all AWS Region
	AllAWSRegions = []string{
		"ap-northeast-2",
		"ap-south-1",
		"eu-north-1",
		"eu-west-3",
		"eu-west-2",
		"eu-west-1",
		"ap-northeast-1",
		"sa-east-1",
		"ca-central-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"eu-central-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}

	// AWSCredentialsPath is the file path of aws credentials
	AWSCredentialsPath = HomeDir() + "/.aws/credentials"

	// AWSConfigPath is the file path of aws config
	AWSConfigPath = HomeDir() + "/.aws/config"

	// AllowedMethods means a list of methods allowed
	AllowedMethods = []string{
		"GET",
		"POST",
	}
)

// Get Home Directory
func HomeDir() string {
	if h := os.Getenv("HOME"); h != EmptyString {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
