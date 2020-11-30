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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"

	"github.com/DevopsArtFactory/bigshot/pkg/client"
	"github.com/DevopsArtFactory/bigshot/pkg/color"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/slacker"
	"github.com/DevopsArtFactory/bigshot/pkg/templates"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type Tracer struct {
	Name     string
	Attacker *http.Client
	Rate     int
	Duration time.Duration
	Target   string
	Method   string
	Body     map[string]string
	Header   map[string]string
	Protocol string
	Region   string
	SlackURL []string
	Result   schema.Result
	LogLevel string
	Timeout  int
}

// SetRate sets rate of request
func (t *Tracer) SetRate(i int) {
	t.Rate = i
}

// SetTimeout sets request timeout
func (t *Tracer) SetTimeout(i int) {
	if i == 0 {
		i = 3
	}
	t.Attacker.Timeout = time.Duration(i) * time.Second
	logrus.Infof("Timout: %d", i)
}

// SetTarget sets the target for the request
func (t *Tracer) SetTarget(s string) {
	logrus.Infof("Target: %s", s)
	t.Target = s
	if strings.HasPrefix(s, "http://") {
		t.Protocol = constants.HTTP
	}

	if strings.HasPrefix(s, "https://") {
		t.Protocol = constants.HTTPS
	}
}

// SetLogLevel sets loglevel
func (t *Tracer) SetLogLevel(logLevel string) {
	logrus.Infof("LogLevel: %s", logLevel)
	t.LogLevel = logLevel
}

// Trace starts tracing
func (t *Tracer) Trace() error {
	var bodyJSON string
	var err error
	if t.Body != nil {
		b, err := json.Marshal(t.Body)
		if err != nil {
			return err
		}

		bodyJSON = string(b)
	}

	var req *http.Request
	if len(bodyJSON) > 0 {
		req, err = http.NewRequest(t.Method, t.Target, bytes.NewBuffer([]byte(bodyJSON)))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest(t.Method, t.Target, nil)
		if err != nil {
			return err
		}
	}

	if t.Header != nil {
		header := http.Header{}
		for k, v := range t.Header {
			header.Set(k, v)
		}
		req.Header = header
	}

	td := schema.TracingData{
		URL: t.Target,
	}

	trace := &httptrace.ClientTrace{
		DNSStart:          func(dsi httptrace.DNSStartInfo) { td.DNSStart = time.Now() },
		DNSDone:           func(ddi httptrace.DNSDoneInfo) { td.DNSDone = time.Now() },
		TLSHandshakeStart: func() { td.TLSHandshakeStart = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if err != nil {
				panic(err)
			}
			td.TLSHandshakeDone = time.Now()
		},
		ConnectStart: func(network, addr string) {
			if td.DNSDone.IsZero() {
				td.ConnectionStart = time.Now()
			} else {
				td.ConnectionStart = td.DNSDone
			}
		},
		ConnectDone: func(network, addr string, err error) {
			td.ConnectAddr = addr
			td.ConnectionDone = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			td.GotConn = time.Now()
		},

		GotFirstResponseByte: func() {
			td.GetFirstResponseBtye = time.Now()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	if err := t.SetupTransport(); err != nil {
		return err
	}

	resp, err := t.Attacker.Do(req)
	if err != nil {
		if sendErr := t.SendErrorAlarm(err.Error()); sendErr != nil {
			logrus.Errorln(sendErr)
		}
		return err
	}

	td.FinishRequest = time.Now()

	if err := resp.Body.Close(); err != nil {
		return err
	}

	if err := t.SetResult(td, resp); err != nil {
		return err
	}

	return nil
}

// Run starts to trace request
func (t *Tracer) Run() error {
	if err := t.Trace(); err != nil {
		return err
	}

	if t.LogLevel == "debug" {
		if err := t.PrintResult(); err != nil {
			return err
		}
	}

	if len(t.SlackURL) > 0 && t.Result.Response.StatusCode != 200 {
		if err := t.SendAlarm(); err != nil {
			return err
		}
	}

	if err := t.SaveData(); err != nil {
		return err
	}

	return nil
}

// RunWithResult runs tracing and returns result
func (t *Tracer) RunWithResult() (*schema.Result, error) {
	if err := t.Trace(); err != nil {
		return nil, err
	}

	return &t.Result, nil
}

// PrintResult prints result
func (t *Tracer) PrintResult() error {
	var scanData = struct {
		Summary schema.Result
	}{
		Summary: t.Result,
	}

	funcMap := template.FuncMap{
		"decorate": color.DecorateAttr,
		"format":   tools.Formatting,
	}

	// Template for scan result
	w := tabwriter.NewWriter(os.Stdout, 0, 5, 3, ' ', tabwriter.TabIndent)
	tt := template.Must(template.New("Result").Funcs(funcMap).Parse(templates.TracingTemplate))

	err := tt.Execute(w, scanData)
	if err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if err := t.DrawResultTable(); err != nil {
		return err
	}

	return nil
}

// SendAlarm sends slack alarm
func (t *Tracer) SendAlarm() error {
	var attachments []slacker.Attachment
	var blocks []slacker.Block
	slack := slacker.NewSlackClient()

	// TODO: create chart and upload it to S3
	//filePath := "output.png"
	//if err := t.CreateChart(filePath); err != nil {
	//	return err
	//}

	// title
	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*%s*", t.Name),
		},
	})

	// divider
	blocks = append(blocks, slacker.Block{
		Type: "divider",
	})

	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Domain*: `%s`", t.Target),
		},
	})

	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Connect IP*: `%s`", t.Result.TracingData.ConnectAddr),
		},
	})

	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Status Code*: %d", t.Result.Response.StatusCode),
		},
	})

	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Status Message*: %s", t.Result.Response.StatusMsg),
		},
	})

	// divider
	blocks = append(blocks, slacker.Block{
		Type: "divider",
	})

	// Tracing Data
	index := 1
	var fields []slacker.Field
	fields = append(fields, slacker.Field{
		Title: fmt.Sprintf("[%d]DNS Lookup", index),
		Value: t.Result.TracingData.DNSLookup.String(),
		Short: true,
	})
	index++

	fields = append(fields, slacker.Field{
		Title: fmt.Sprintf("[%d]TCP Connection", index),
		Value: t.Result.TracingData.TCPConnection.String(),
		Short: true,
	})
	index++

	if t.Protocol == constants.HTTPS {
		fields = append(fields, slacker.Field{
			Title: fmt.Sprintf("[%d]TLS Handshake", index),
			Value: t.Result.TracingData.TLSHandShacking.String(),
			Short: true,
		})
		index++
	}

	fields = append(fields, slacker.Field{
		Title: fmt.Sprintf("[%d]Server Processing", index),
		Value: t.Result.TracingData.ServerProcessing.String(),
		Short: true,
	})
	index++

	fields = append(fields, slacker.Field{
		Title: fmt.Sprintf("[%d]Content Transfer", index),
		Value: t.Result.TracingData.ContentTransfer.String(),
		Short: true,
	})

	attachments = append(attachments, slacker.Attachment{
		Color:  constants.ErrorColor,
		Text:   fmt.Sprintf("*Request Tracing result* - Total Time: %s", t.Result.TracingData.FinishRequest.String()),
		Fields: fields,
	})

	for _, URL := range t.SlackURL {
		if err := slack.SendMessageWithWebHook(attachments, blocks, URL); err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	return nil
}

// SendErrorAlarm sends error alarm
func (t *Tracer) SendErrorAlarm(errorMsg string) error {
	var attachments []slacker.Attachment
	var blocks []slacker.Block
	slack := slacker.NewSlackClient()

	// title
	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("Error occurred: `%s`", t.Target),
		},
	})

	// divider
	blocks = append(blocks, slacker.Block{
		Type: "divider",
	})

	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*%s*", t.Region),
		},
	})

	blocks = append(blocks, slacker.Block{
		Type: "section",
		Text: &slacker.Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*%s*", errorMsg),
		},
	})

	for _, URL := range t.SlackURL {
		if err := slack.SendMessageWithWebHook(attachments, blocks, URL); err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	return nil
}

// SetResult sets the result to Tracker.Result
func (t *Tracer) SetResult(td schema.TracingData, response *http.Response) error {
	res := schema.Response{}
	var err error

	// Parse status code
	res.StatusCode, res.StatusMsg, err = ParseStatus(response.Status)
	if err != nil {
		return err
	}

	header := map[string][]string{}
	for k, v := range response.Header {
		if v != nil {
			header[k] = v
		}
	}

	res.Header = header

	td = Calculated(td, t.Protocol == constants.HTTPS)

	t.Result = schema.Result{
		TracingData: td,
		Response:    res,
	}

	return nil
}

// DrawResultTable draws result table of request
func (t *Tracer) DrawResultTable() error {
	var data [][]string
	//tableString := &strings.Builder{}
	table := tablewriter.NewWriter(os.Stdout)
	if t.Protocol == constants.HTTPS {
		data = [][]string{
			{
				t.Result.TracingData.DNSLookup.String(),
				t.Result.TracingData.TCPConnection.String(),
				t.Result.TracingData.TLSHandShacking.String(),
				t.Result.TracingData.ServerProcessing.String(),
				t.Result.TracingData.ContentTransfer.String(),
			},
		}
		table.SetHeader([]string{"DNS Lookup", "TCP Connection", "TLS Handshake", "Server Processing", "Content Transfer"})
	} else {
		data = [][]string{
			{
				t.Result.TracingData.DNSLookup.String(),
				t.Result.TracingData.TCPConnection.String(),
				t.Result.TracingData.ServerProcessing.String(),
				t.Result.TracingData.ContentTransfer.String(),
			},
		}
		table.SetHeader([]string{"DNS Lookup", "TCP Connection", "Server Processing", "Content Transfer"})
	}

	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetRowLine(true)
	table.Render()

	return nil
}

// Calculated calculates the durations of each step
func Calculated(td schema.TracingData, tls bool) schema.TracingData {
	td.DNSLookup = td.DNSDone.Sub(td.DNSStart)
	td.TCPConnection = td.ConnectionDone.Sub(td.ConnectionStart)
	if tls {
		td.TLSHandShacking = td.TLSHandshakeDone.Sub(td.TLSHandshakeStart)
	}
	td.ServerProcessing = td.GetFirstResponseBtye.Sub(td.GotConn)
	td.ContentTransfer = td.FinishRequest.Sub(td.GetFirstResponseBtye)

	td.Total = td.FinishRequest.Sub(td.DNSStart)

	return td
}

// ParseStatus parses status code and msg
func ParseStatus(status string) (int, string, error) {
	split := strings.Split(status, " ")
	if len(split) < 2 {
		return 0, constants.EmptyString, fmt.Errorf("response status is not correct: %s", status)
	}

	statusCode, err := strconv.Atoi(split[0])
	if err != nil {
		return 0, constants.EmptyString, err
	}

	statusString := strings.Join(split[1:], " ")

	return statusCode, statusString, nil
}

// SetupTransport sets transport configuration of request
func (t *Tracer) SetupTransport() error {
	parsed, err := url.Parse(t.Target)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if t.Protocol == constants.HTTPS {
		host, _, err := net.SplitHostPort(parsed.Host)
		if err != nil {
			host = parsed.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: false,
			Certificates:       nil,
		}

		err = http2.ConfigureTransport(tr)
		if err != nil {
			return err
		}
	}

	t.Attacker.Transport = tr
	return nil
}

// SaveData saves data to time-series database
func (t *Tracer) SaveData() error {
	writer := client.NewTimeStreamClient(constants.DefaultRegion)
	if err := writer.WriteData("bigshot", "synthetics", t.Region, t.Protocol, t.Result); err != nil {
		return err
	}

	return nil
}

// SetSlackURL set slack URL for notification
func (t *Tracer) SetSlackURL(s []string) {
	t.SlackURL = s
	logrus.Infof("SlackUrl: %s", strings.Join(s, ","))
}

// SetMethod sets method of API
func (t *Tracer) SetMethod(s string) {
	t.Method = s
	logrus.Infof("Method: %s", s)
}

// SetBody sets body data
func (t *Tracer) SetBody(m map[string]string) {
	t.Body = m
}

// SetHeader sets body data
func (t *Tracer) SetHeader(m map[string]string) {
	t.Header = m
}

// NewTracer creates tracer test
func NewTracer(region string) Shooter {
	return &Tracer{
		Name: fmt.Sprintf("Request from %s", region),
		Attacker: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Duration: constants.DefaultWorkerDuration,
		Region:   region,
		Target:   constants.EmptyString,
	}
}
