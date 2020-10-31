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
	"golang.org/x/net/http2"

	"github.com/DevopsArtFactory/bigshot/pkg/color"
	"github.com/DevopsArtFactory/bigshot/pkg/constants"
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
	Result   Result
}

type Result struct {
	TracingData TracingData
	Response    Response
}

type Response struct {
	StatusCode int
	StatusMsg  string
	Header     map[string][]string
}

type TracingData struct {
	// Real time
	DNSStart             time.Time
	DNSDone              time.Time
	TLSHandshakeStart    time.Time
	TLSHandshakeDone     time.Time
	ConnectionStart      time.Time
	ConnectionDone       time.Time
	GotConn              time.Time
	GetFirstResponseBtye time.Time
	FinishRequest        time.Time
	ConnectAddr          string

	// Stat
	DNSLookup        time.Duration
	TCPConnection    time.Duration
	TLSHandShacking  time.Duration
	ServerProcessing time.Duration
	ContentTransfer  time.Duration
	Total            time.Duration
}

// SetRate sets rate of request
func (t *Tracer) SetRate(i int) {
	t.Rate = i
}

// SetTarget sets the target for the request
func (t *Tracer) SetTarget(s string) {
	t.Target = s
	if strings.HasPrefix(s, "http://") {
		t.Protocol = constants.HTTP
	}

	if strings.HasPrefix(s, "https://") {
		t.Protocol = constants.HTTPS
	}
}

// Run starts to trace request
func (t *Tracer) Run() error {
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

	td := TracingData{}
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
		return err
	}

	td.FinishRequest = time.Now()

	if err := resp.Body.Close(); err != nil {
		return err
	}

	if err := t.SetResult(td, resp); err != nil {
		return err
	}

	if err := t.PrintResult(); err != nil {
		return err
	}

	if len(t.SlackURL) > 0 {
		if err := t.SendAlarm(); err != nil {
			return err
		}
	}

	return nil
}

// PrintResult prints result
func (t *Tracer) PrintResult() error {
	var scanData = struct {
		Summary Result
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
	slack := slacker.NewSlackClient()

	for _, URL := range t.SlackURL {
		if err := slack.SendMessageWithWebHook(fmt.Sprintf("tracer success: %s, target: `%s`", t.Region, t.Target), URL); err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	return nil
}

// SetResult sets the result to Tracker.Result
func (t *Tracer) SetResult(td TracingData, response *http.Response) error {
	res := Response{}
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

	t.Result = Result{
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
func Calculated(td TracingData, tls bool) TracingData {
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

// SetSlackURL set slack URL for notification
func (t *Tracer) SetSlackURL(s []string) {
	t.SlackURL = s
}

// SetMethod sets method of API
func (t *Tracer) SetMethod(s string) {
	t.Method = s
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
