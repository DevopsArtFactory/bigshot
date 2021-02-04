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

package shot

import (
	"errors"

	"github.com/DevopsArtFactory/bigshot/pkg/builder"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

type Shooter interface {
	SetRate(int)
	SetTimeout(int)
	SetLogLevel(string)
	SetTarget(string, string)
	SetMethod(string)
	SetBody(map[string]string)
	SetHeader(map[string]string)
	SetSlackURL([]string)
	Run() error
	RunWithResult() (*schema.Result, error)
}

// NewShooter returns new shooter
func NewShooter(t, region string) Shooter {
	switch t {
	case "trace":
		return NewTracer(region)
	case "ping":
		return NewPing(region)
	case "vegeta":
		return NewVegeta(region)
	}

	return nil
}

// Shoot tries shooting target checking
func Shoot(target schema.Target, resultNeeded bool) (*schema.Result, error) {
	defaultRegion, err := builder.GetDefaultRegion(constants.DefaultProfile)
	if err != nil {
		return nil, err
	}

	shooter := NewShooter(constants.DefaultShooter, defaultRegion)
	if shooter == nil {
		return nil, errors.New("cannot create new finder")
	}
	shooter.SetTarget(*target.URL, *target.Port)
	shooter.SetMethod(*target.Method)
	if target.Body != nil {
		shooter.SetBody(target.Body)
	}
	if target.Header != nil {
		shooter.SetHeader(target.Header)
	}
	shooter.SetRate(1)
	if resultNeeded {
		result, err := shooter.RunWithResult()
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, shooter.Run()
}
