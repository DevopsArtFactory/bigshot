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
package schema

import "time"

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
	URL                  string
	ConnectAddr          string
	DNSStart             time.Time
	DNSDone              time.Time
	TLSHandshakeStart    time.Time
	TLSHandshakeDone     time.Time
	ConnectionStart      time.Time
	ConnectionDone       time.Time
	GotConn              time.Time
	GetFirstResponseBtye time.Time
	FinishRequest        time.Time

	// Stat
	DNSLookup        time.Duration
	TCPConnection    time.Duration
	TLSHandShaking  time.Duration
	ServerProcessing time.Duration
	ContentTransfer  time.Duration
	Total            time.Duration

	// Stat as String
	DNSLookupStr        string
	TCPConnectionStr    string
	TLSHandShakingStr   string
	ServerProcessingStr string
	ContentTransferStr  string
	TotalStr            string
}
