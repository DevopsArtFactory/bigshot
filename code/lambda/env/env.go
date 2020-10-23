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

package env

import (
	"os"
	"reflect"
	"strings"

	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Env struct {
	Region   string `json:"region"`
	RunType  string `json:"type"`
	URL      string `json:"url"`
	Mode     string `json:"mode"`
	Template string `json:"template"`
}

// GetEnvs returns all environment variables
func GetEnvs() Env {
	var keys []string
	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		keys = append(keys, split[0])
	}
	envs := Env{}

	val := reflect.ValueOf(&envs).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		key := strings.ReplaceAll(typeField.Tag.Get("json"), "_", "-")
		if tools.IsStringInArray(key, keys) {
			t := val.FieldByName(typeField.Name)
			t.SetString(os.Getenv(key))
		}
	}

	return envs
}
