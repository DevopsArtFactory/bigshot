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

package generator

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

func TestGetTargetRegions(t *testing.T) {
	testData := []struct {
		Input  string
		Output []string
	}{
		{
			Input:  "ap-northeast-2",
			Output: []string{"ap-northeast-2"},
		},
		{
			Input:  "ap-northeast-2,us-east-1",
			Output: []string{"ap-northeast-2", "us-east-1"},
		},
		{
			Input:  constants.EmptyString,
			Output: constants.AllAWSRegions,
		},
		{
			Input:  "ap-northeast-2,us-east-1,ap-south-100",
			Output: nil,
		},
	}

	for _, td := range testData {
		r, _ := GetTargetRegions(td.Input, false, true)
		if strings.Join(r, ",") != strings.Join(td.Output, ",") {
			t.Errorf("expected: %v / output: %v", td.Output, r)
		}
	}
}

func TestCheckAvailableRegions(t *testing.T) {
	config := &schema.Template{
		Targets: []schema.Target{
			{
				URL: aws.String("https://www.google.com"),
				Regions: []string{
					"ap-northeast-2",
					"ap-northeast-1",
				},
			},
		},
		Regions: []schema.Region{
			{
				Region: aws.String("ap-northeast-1"),
			},
			{
				Region: aws.String("ap-northeast-2"),
			},
			{
				Region: aws.String("us-east-1"),
			},
		},
	}

	regions, err := GetRegionsFromTemplate(config, false, "ap-northeast-2")
	if err != nil {
		t.Errorf(err.Error())
	}

	regionIDs := getRegionIDs(regions)

	if err := CheckAvailableRegions(regionIDs, config.Targets); err != nil {
		t.Errorf(err.Error())
	}

	config.Targets[0].Regions = append(config.Targets[0].Regions, "us-west-1")

	if err := CheckAvailableRegions(regionIDs, config.Targets); err == nil {
		t.Errorf("Error occurred on region validation check")
	}
}
