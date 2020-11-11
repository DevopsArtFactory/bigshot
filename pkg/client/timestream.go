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

package client

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"

	"github.com/DevopsArtFactory/bigshot/pkg/constants"
	"github.com/DevopsArtFactory/bigshot/pkg/schema"
	"github.com/DevopsArtFactory/bigshot/pkg/tools"
)

type TimeStream struct {
	WriteClient *timestreamwrite.TimestreamWrite
}

// NewTimeStreamClient creates TimeStream client
func NewTimeStreamClient(region string) *TimeStream {
	tr := &http.Transport{
		ResponseHeaderTimeout: 20 * time.Second,
		// Using DefaultTransport values for other parameters: https://golang.org/pkg/net/http/#RoundTripper
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// So client makes HTTP/2 requests
	http2.ConfigureTransport(tr)

	sess, _ := session.NewSession(&aws.Config{Region: aws.String(region), MaxRetries: aws.Int(10), HTTPClient: &http.Client{Transport: tr}})

	return &TimeStream{
		WriteClient: GetTimeStreamWriteClientFn(sess, region, nil),
	}
}

// GetTimeStreamWriteClientFn creates a new AWS cloudwatch client
func GetTimeStreamWriteClientFn(sess client.ConfigProvider, region string, creds *credentials.Credentials) *timestreamwrite.TimestreamWrite {
	if creds == nil {
		return timestreamwrite.New(sess, &aws.Config{Region: aws.String(region)})
	}
	return timestreamwrite.New(sess, &aws.Config{Region: aws.String(region), Credentials: creds})
}

// WriteData writes data to time series database
func (t *TimeStream) WriteData(databaseName, tableName, region, protocol string, result schema.Result) error {
	now := time.Now()
	currentTimeInSeconds := now.Unix()
	dimensions := []*timestreamwrite.Dimension{
		{
			Name:  aws.String("target"),
			Value: aws.String(result.TracingData.URL),
		},
		{
			Name:  aws.String("region"),
			Value: aws.String(region),
		},
	}
	inputTime := aws.String(strconv.FormatInt(currentTimeInSeconds, 10))
	timeUnit := aws.String("SECONDS")

	input := &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
		Records: []*timestreamwrite.Record{
			{
				Dimensions:       dimensions,
				MeasureName:      aws.String("status_code"),
				MeasureValue:     aws.String(tools.IntToString(result.Response.StatusCode)),
				MeasureValueType: aws.String("BIGINT"),
				Time:             inputTime,
				TimeUnit:         timeUnit,
			},
			{
				Dimensions:       dimensions,
				MeasureName:      aws.String("dns_lookup"),
				MeasureValue:     aws.String(tools.Int64ToString(result.TracingData.DNSLookup.Milliseconds())),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             inputTime,
				TimeUnit:         timeUnit,
			},
			{
				Dimensions:       dimensions,
				MeasureName:      aws.String("tcp_connection"),
				MeasureValue:     aws.String(tools.Int64ToString(result.TracingData.TCPConnection.Milliseconds())),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             inputTime,
				TimeUnit:         timeUnit,
			},
			{
				Dimensions:       dimensions,
				MeasureName:      aws.String("server_processing"),
				MeasureValue:     aws.String(tools.Int64ToString(result.TracingData.ServerProcessing.Milliseconds())),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             inputTime,
				TimeUnit:         timeUnit,
			},
			{
				Dimensions:       dimensions,
				MeasureName:      aws.String("content_transfer"),
				MeasureValue:     aws.String(tools.Int64ToString(result.TracingData.ContentTransfer.Milliseconds())),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             inputTime,
				TimeUnit:         timeUnit,
			},
			{
				Dimensions:       dimensions,
				MeasureName:      aws.String("total"),
				MeasureValue:     aws.String(tools.Int64ToString(result.TracingData.Total.Milliseconds())),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             inputTime,
				TimeUnit:         timeUnit,
			},
		},
	}

	if protocol == constants.HTTPS {
		input.Records = append(input.Records, &timestreamwrite.Record{
			Dimensions:       dimensions,
			MeasureName:      aws.String("tls_handshaking"),
			MeasureValue:     aws.String(tools.Int64ToString(result.TracingData.TLSHandShacking.Milliseconds())),
			MeasureValueType: aws.String("DOUBLE"),
			Time:             inputTime,
			TimeUnit:         timeUnit,
		})
	}

	_, err := t.WriteClient.WriteRecords(input)

	if err != nil {
		return err
	}

	logrus.Infof("[%s] data is successfully saved to %s:%s", now.Format(time.RFC3339), databaseName, tableName)

	return nil
}
