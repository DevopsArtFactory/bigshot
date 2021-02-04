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
	"strconv"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

// ParseTemplates parses items to list of template
func ParseTemplates(items []map[string]*dynamodb.AttributeValue) ([]schema.Template, error) {
	var configs []schema.Template

	for _, item := range items {
		config, err := ChangeItemToConfig(item)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *config)
	}

	return configs, nil
}

// ChangeItemToConfig changes item value from dynamoDB to schema.Template
func ChangeItemToConfig(item map[string]*dynamodb.AttributeValue) (*schema.Template, error) {
	config := &schema.Template{
		Name: item["name"].S,
	}

	interval, err := strconv.Atoi(*item["interval"].N)
	if err != nil {
		return nil, err
	}
	config.Interval = &interval

	timeout, err := strconv.Atoi(*item["timeout"].N)
	if err != nil {
		return nil, err
	}
	config.Timeout = &timeout

	regions := []schema.Region{}
	for _, region := range item["regions"].L {
		regions = append(regions, schema.Region{Region: region.M["region"].S})
	}
	config.Regions = regions

	slackURLs := []string{}
	for _, url := range item["slack_urls"].SS {
		slackURLs = append(slackURLs, *url)
	}
	config.SlackURLs = slackURLs

	targets := []schema.Target{}
	for _, target := range item["targets"].L {
		t := schema.Target{
			URL:    target.M["url"].S,
			Method: target.M["method"].S,
		}

		if _, ok := target.M["header"]; ok {
			headers := map[string]string{}
			for key, val := range target.M["header"].M {
				headers[key] = *val.S
			}
			t.Header = headers
		}

		if _, ok := target.M["body"]; ok {
			bodies := map[string]string{}
			for key, val := range target.M["body"].M {
				bodies[key] = *val.S
			}
			t.Body = bodies
		}

		targets = append(targets, t)
	}
	config.Targets = targets

	return config, nil
}
