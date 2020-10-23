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
package tools

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
)

// ClearOsEnv removes all environment variables about AWS
func ClearOsEnv() error {
	logrus.Debugf("remove environment variable")
	if err := os.Unsetenv("AWS_ACCESS_KEY_ID"); err != nil {
		return err
	}
	if err := os.Unsetenv("AWS_SECRET_ACCESS_KEY"); err != nil {
		return err
	}

	if err := os.Unsetenv("AWS_SESSION_TOKEN"); err != nil {
		return err
	}

	return nil
}

// Check if file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//Figure out if string is in array
func IsStringInArray(s string, arr []string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}

	return false
}

// IsExpired compares current time with (targetDate + timeAdded)
func IsExpired(targetDate time.Time, timeAdded time.Duration) bool {
	return time.Since(targetDate.Add(timeAdded)) > 0
}

// SetUpLogs set logrus log format
func SetUpLogs(stdErr io.Writer, level string) error {
	logrus.SetOutput(stdErr)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("parsing log level: %w", err)
	}
	logrus.SetLevel(lvl)
	return nil
}

// Formatting removes nil value
func Formatting(i interface{}) interface{} {
	switch reflect.TypeOf(i).String() {
	case "*time.Time":
		if i.(*time.Time) == nil {
			return constants.EmptyString
		}
	case "*string":
		if i.(*string) == nil {
			return constants.EmptyString
		}

	case "[]string":
		arr := i.([]string)
		switch len(arr) {
		case 0:
			return constants.EmptyString
		case 1:
			return arr[0]
		default:
			return strings.Join(arr, ", ")
		}
	}

	return i
}

// GenerateNewWorkerName generates a new name for lambda function
func GenerateNewWorkerName(region, mode string) string {
	return fmt.Sprintf("%s-%s-%s", constants.CommonNamePrefix, mode, region)
}

// GenerateNewLambdaRoleName generates a new role name for lambda function
func GenerateNewLambdaRoleName(region string) string {
	return fmt.Sprintf("%s-%s", constants.CommonNamePrefix, region)
}

// GenerateDescription generates a description for lambda function
func GenerateDescription(region string) string {
	return fmt.Sprintf("Bigshot lambda in %s", region)
}

// GenerateNewTableName generates a new name for lambda function
func GenerateNewTableName() string {
	return fmt.Sprintf("%s-metadata", constants.ControllerNamePrefix)
}

// Wait runs empty timer
func Wait(timer int64, formatMsg string) {
	logrus.Infof(formatMsg, timer)
	for timer > 0 {
		fmt.Print(".")
		time.Sleep(1 * time.Second)
		timer--
	}
	fmt.Println("done")
}

// ReadZipFile reads zip binary file from local disk
func ReadZipFile(path string) ([]byte, error) {
	if !FileExists(path) {
		return nil, fmt.Errorf("file does not exist: %s", path)
	}
	logrus.Infof("zip file is detected: %s", path)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

// RoundTime creates rounded time
func RoundTime(d time.Duration) string {
	var r float64
	var suffix string
	switch {
	case d > time.Minute:
		r = d.Minutes()
		suffix = "m"
	case d > time.Second:
		r = d.Seconds()
		suffix = "s"
	default:
		r = float64(d.Milliseconds())
		suffix = "ms"
	}

	return fmt.Sprintf("%.2f%s", r, suffix)
}
