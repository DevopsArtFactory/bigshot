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

package slacker

import (
	"bytes"
	"encoding/json"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type Slack struct {
	Client *slack.Client
	Color  string
}

type Body struct {
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Block struct {
	Type string `json:"type"`
	Text *Text  `json:"text,omitempty"`
}

type Text struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type Attachment struct {
	Text   string  `json:"text"`
	Color  string  `json:"color"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackClient creates new slack client
func NewSlackClient() Slack {
	return Slack{
		Client: slack.New(constants.EmptyString),
	}
}

// SendMessageWithWebhook is for WebhookURL
func (s *Slack) SendMessageWithWebHook(attachments []Attachment, blocks []Block, url string) error {
	logrus.Infof("slack target is set: %s", url)
	slackBody, _ := json.Marshal(Body{
		Attachments: attachments,
		Blocks:      blocks,
	})

	return sendSlackRequest(slackBody, url)
}

// sendSlackRequest sends request for slack message
func sendSlackRequest(slackBody []byte, url string) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("non-ok response returned from Slack")
	}
	resp.Body.Close()

	return nil
}
