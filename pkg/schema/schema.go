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
type Config struct {
	// Bigshot template name. This will be key of dynamodb table
	Name string `yaml:"name"`

	// Bigshot template name. This will be key of dynamodb table
	SlackURL string `yaml:"slack_url"`

	// List of targets for api check
	Targets []Target `yaml:"targets,omitempty"`

	// List of regions.
	Regions []Region `yaml:"regions,omitempty"`
}

// Target configuration
type Target struct {
	// Target URL of API
	URL string `json:"url,omitempty"`
}

// Region configuration
type Region struct {
	// ID of region
	Region string `json:"region,omitempty"`
}