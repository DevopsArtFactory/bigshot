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
	"net/url"
	"time"

	"github.com/go-ping/ping"
	"github.com/sirupsen/logrus"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

type Ping struct {
	Name     string
	Attacker *ping.Pinger
	Rate     int
	Duration time.Duration
	Target   string
	SlackURL []string
}

func (p *Ping) SetTimeout(i int) {
	panic("implement me")
}

func (p *Ping) SetLogLevel(s string) {
	panic("implement me")
}

func (p *Ping) RunWithResult() (*schema.Result, error) {
	panic("implement me")
}

func (p *Ping) SetMethod(s string) {
	panic("implement me")
}

func (p *Ping) SetBody(m map[string]string) {
	panic("implement me")
}

func (p *Ping) SetHeader(m map[string]string) {
	panic("implement me")
}

// NewPing creates ping test
func NewPing(region string) Shooter {
	return &Ping{
		Name:     fmt.Sprintf("Request from %s", region),
		Attacker: nil,
		Duration: constants.DefaultWorkerDuration,
		Target:   constants.EmptyString,
	}
}

// SetSlackURL set slack URL for notification
func (p *Ping) SetSlackURL(s []string) {
	p.SlackURL = s
}

// SetRate sets rate for api call
func (p *Ping) SetRate(freq int) {
	p.Attacker.Count = freq
}

// SetTarget sets target for api call
func (p *Ping) SetTarget(url, port string) {
	p.Target = url
	parsed, err := parseURL(url)
	if err != nil {
		logrus.Errorf(err.Error())
	}
	p.Attacker = ping.New(parsed)
}

// Run starts ping test
func (p *Ping) Run() error {
	p.Attacker.SetPrivileged(true)

	if err := p.Attacker.Run(); err != nil {
		return err
	}

	stats := p.Attacker.Statistics()

	logrus.Infof("%+v", stats)

	return nil
}

// parseURL retrieves host from url
func parseURL(ori string) (string, error) {
	u, err := url.Parse(ori)
	if err != nil {
		return ori, err
	}

	return u.Host, nil
}
