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

package builder

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Builder struct {
	Config *schema.Config
	Flags  Flags
}

type Flags struct {
	Region    string `json:"region"`
	Config    string `json:"config"`
	AllRegion bool   `json:"all"`
	ZipFile   string `json:"zip_file"`
}

// ValidateFlags checks validation of flags
func ValidateFlags(flags Flags) error {
	if len(flags.Region) > 0 && !tools.IsStringInArray(flags.Region, constants.AllAWSRegions) {
		return fmt.Errorf("region is not correct: %s", flags.Region)
	}

	if len(flags.ZipFile) > 0 && !tools.FileExists(flags.ZipFile) {
		return fmt.Errorf("file does not exist: %s", flags.ZipFile)
	}

	return nil
}

// CreateNewBuilder creates new builder
func CreateNewBuilder(flags Flags) (*Builder, error) {
	var config schema.Config

	if len(flags.Config) == 0 {
		logrus.Debug("You have no config file")
		return New(nil, flags), nil
	}

	if !tools.FileExists(flags.Config) {
		return nil, fmt.Errorf("configuration file does not exist: %s", flags.Config)
	}

	file, err := ioutil.ReadFile(flags.Config)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return New(&config, flags), nil
}

// Create new builder
func New(config *schema.Config, flags Flags) *Builder {
	return &Builder{
		Config: config,
		Flags:  flags,
	}
}

// GetFlags makes flags from command
func GetFlags() (Flags, error) {
	keys := viper.AllKeys()
	flags := Flags{}

	val := reflect.ValueOf(&flags).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		key := strings.ReplaceAll(typeField.Tag.Get("json"), "_", "-")
		if tools.IsStringInArray(key, keys) {
			t := val.FieldByName(typeField.Name)
			if t.CanSet() {
				switch t.Kind() {
				case reflect.String:
					t.SetString(viper.GetString(key))
				case reflect.Int:
					t.SetInt(viper.GetInt64(key))
				case reflect.Bool:
					t.SetBool(viper.GetBool(key))
				}
			}
		}
	}

	return flags, nil
}

// GetDefaultRegion gets default region with env or configuration file
func GetDefaultRegion(profile string) (string, error) {
	if len(os.Getenv(constants.DefaultRegionVariable)) > 0 {
		return os.Getenv(constants.DefaultRegionVariable), nil
	}

	functions := []func() (*ini.File, error){
		ReadAWSCredentials,
		ReadAWSConfig,
	}

	for _, f := range functions {
		cfg, err := f()
		if err != nil {
			return constants.EmptyString, err
		}

		section, err := cfg.GetSection(profile)
		if err != nil {
			return constants.EmptyString, err
		}

		if _, err := section.GetKey("region"); err == nil && len(section.Key("region").String()) > 0 {
			return section.Key("region").String(), nil
		}
	}
	return constants.EmptyString, errors.New("no aws region configuration exists")
}

// ReadAWSCredentials parse an aws credentials
func ReadAWSCredentials() (*ini.File, error) {
	if !tools.FileExists(constants.AWSCredentialsPath) {
		return ReadAWSConfig()
	}

	cfg, err := ini.Load(constants.AWSCredentialsPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// ReadAWSConfig parse an aws configuration
func ReadAWSConfig() (*ini.File, error) {
	if !tools.FileExists(constants.AWSConfigPath) {
		return nil, fmt.Errorf("no aws configuration file exists in $HOME/%s", constants.AWSConfigPath)
	}

	cfg, err := ini.Load(constants.AWSConfigPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}