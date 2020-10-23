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

package config

import (
	"github.com/aws/aws-sdk-go/aws"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Config struct {
	Name                 string
	Description          string
	EnvironmentVariables map[string]*string
	Handler              string
	Publish              bool
	MemorySize           int64
	Role                 *string
	Runtime              string
	Tags                 map[string]*string
	Timeout              int64
	S3Bucket             string
	S3Key                string
	ZipFile              []byte
}

// GetBaseWorkerConfig returns base lambda configuration
func GetBaseWorkerConfig(region, mode, template string, roleArn *string, zipFile []byte) Config {
	env := GetEnvironmentVariables(region, mode, template)
	cf := Config{
		Name:                 tools.GenerateNewWorkerName(region, mode),
		Description:          tools.GenerateDescription(region),
		EnvironmentVariables: env,
		Handler:              "out/code/lambda/handler",
		MemorySize:           int64(256),
		Tags:                 GetTags(region, mode),
		Role:                 roleArn,
		Publish:              true,
		Runtime:              constants.GoRunTime,
		Timeout:              int64(300),
	}
	if zipFile != nil {
		cf.ZipFile = zipFile
	} else {
		cf.S3Bucket = constants.DefaultS3Bucket
		cf.S3Key = constants.DefaultS3Key
	}

	return cf
}

// GetEnvironmentVariables makes environment variables
func GetEnvironmentVariables(region, mode, template string) map[string]*string {
	env := map[string]*string{
		"region": aws.String(region),
		"mode":   aws.String(mode),
	}

	if len(template) > 0 && mode == constants.ControllerMode {
		env["template"] = aws.String(template)
	}

	return env
}

// GetTags makes tags
func GetTags(region, mode string) map[string]*string {
	tags := map[string]*string{
		"region": aws.String(region),
		"mode":   aws.String(mode),
	}

	return tags
}
