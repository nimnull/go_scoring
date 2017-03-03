package scoring

import (
	"net"
	"net/http"
	"time"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout: 5 * time.Second,
		KeepAlive: 15 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 5 * time.Second,
}


func GetHttpClient() *http.Client {
	transport.DisableKeepAlives = true
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
}

func All(batches map[int]string, f func(string) bool) bool {
	for _, val := range batches {
		if !f(val) {
			return false
		}
	}
	return true
}
