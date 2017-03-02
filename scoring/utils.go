package scoring

import (
	"net"
	"net/http"
	"time"
)

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

func GetHttpClient() *http.Client {

	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	return netClient
}
