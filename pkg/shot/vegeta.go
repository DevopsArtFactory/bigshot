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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

type Vegeta struct {
	Name     string
	Attacker *vegeta.Attacker
	Rate     *vegeta.Rate
	Duration time.Duration
	SlackURL []string
	Targeter vegeta.Targeter
}

func (v *Vegeta) SetTimeout(i int) {
	panic("implement me")
}

func (v *Vegeta) SetLogLevel(s string) {
	panic("implement me")
}

func (v *Vegeta) RunWithResult() (*schema.Result, error) {
	panic("implement me")
}

func (v *Vegeta) SetMethod(s string) {
	panic("implement me")
}

func (v *Vegeta) SetBody(m map[string]string) {
	panic("implement me")
}

func (v *Vegeta) SetHeader(m map[string]string) {
	panic("implement me")
}

// NewVegeta creates a new vegeta client
func NewVegeta(region string) Shooter {
	return &Vegeta{
		Name:     fmt.Sprintf("Request from %s", region),
		Attacker: vegeta.NewAttacker(),
		Duration: constants.DefaultWorkerDuration,
		Targeter: nil,
		Rate: &vegeta.Rate{
			Freq: 0,
			Per:  time.Second,
		},
	}
}

// SetSlackURL set slack URL for notification
func (v *Vegeta) SetSlackURL(s []string) {
	v.SlackURL = s
}

// SetRate sets rate for api call
func (v *Vegeta) SetRate(freq int) {
	v.Rate.Freq = freq
}

// SetTarget sets target for api call
func (v *Vegeta) SetTarget(url, port string) {
	v.Targeter = vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    url,
	})

	logrus.Infof("target is registered: %s", url)
}

// Shoot runs api request
func (v *Vegeta) Run() error {
	var metrics vegeta.Metrics
	for res := range v.Attacker.Attack(v.Targeter, v.Rate, v.Duration, v.Name) {
		metrics.Add(res)
	}
	metrics.Close()

	logrus.Infof("%+v", metrics)

	return nil
}
