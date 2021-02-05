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
)

type Config struct {
	Name                 string
	Description          string
	Handler              string
	Runtime              string
	S3Bucket             string
	S3Key                string
	EnvironmentVariables map[string]*string
	Tags                 map[string]*string
	Timeout              int64
	MemorySize           int64
	Role                 *string
	SecurityGroups       []string
	Subnets              []string
	ZipFile              []byte
	Publish              bool
	Internal             bool
}

// GetEnvironmentVariables makes environment variables
func GetEnvironmentVariables(region *string, mode string, template *string) map[string]*string {
	env := map[string]*string{
		"region": region,
		"mode":   &mode,
	}

	if template != nil && mode == constants.ManagerMode {
		env["template"] = template
	}

	return env
}

// GetTags makes tags
func GetTags(region *string, mode string) map[string]*string {
	tags := map[string]*string{
		"region": region,
		"mode":   aws.String(mode),
	}

	return tags
}
