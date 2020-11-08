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
	TLSHandShacking  time.Duration
	ServerProcessing time.Duration
	ContentTransfer  time.Duration
	Total            time.Duration
}
