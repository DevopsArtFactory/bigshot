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

package schema

// Configuration for bigshot
type Template struct {
	// Bigshot template name. This will be key of dynamodb table
	Name *string `yaml:"name,omitempty" json:"name"`

	// Lambda log level
	Log *string `yaml:"log,omitempty" json:"log"`

	// Lambda execution timeout in seconds
	Timeout *int `yaml:"timeout,omitempty" json:"timeout"`

	// Synthetic interval in seconds
	Interval *int `yaml:"interval,omitempty" json:"interval"`

	// List of slack URLs for alert
	SlackURLs []string `yaml:"slack_urls,omitempty" json:"slack_urls"`

	// List of targets for api check
	Targets []Target `yaml:"targets,omitempty" json:"targets"`

	// List of regions.
	Regions []Region `yaml:"regions,omitempty" json:"regions"`
}

// Target configuration
type Target struct {
	// Target URL of API
	URL *string `yaml:"url,omitempty" json:"url"`

	// Target Port of API
	Port *string `yaml:"port,omitempty" json:"port"`

	// API method
	Method *string `yaml:"method,omitempty" json:"method"`

	// Body value of API
	Body map[string]string `yaml:"body,omitempty" json:"body"`

	// Header value of API
	Header map[string]string `yaml:"header,omitempty" json:"header"`

	// Target Request timeout
	Timeout *int `yaml:"timeout,omitempty" json:"timeout"`

	// Internal means whether or not to run within VPC
	Internal *bool `yaml:"internal" json:"internal"`

	// Regions means the list of regions to run bigshot check
	Regions []string `yaml:"regions,omitempty" json:"regions"`
}

// Region configuration
type Region struct {
	// ID of region
	Region         *string  `yaml:"region,omitempty" json:"region"`
	SecurityGroups []string `yaml:"security_groups" json:"security_groups"`
	Subnets        []string `yaml:"subnets,omitempty" json:"subnets"`
}
