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

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
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
