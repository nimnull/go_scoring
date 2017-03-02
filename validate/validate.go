package validate

import (
	"net/url"
	"os"
	"log"
)

func ValidateHost(host string) *url.URL {
	hostUrl, err := url.Parse(host)
	if err != nil {
		log.Fatal(err)
	}

	if hostUrl.Scheme == "" || hostUrl.Host == "" {
		log.Fatalf("`%s` is not valid host URL\n", host)
	} else {
		//hostUrl.Path = "/api/v1"
		hostUrl.Path = "/v1.0"
	}
	return hostUrl
}

func ValidateFile(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if f, err := os.Create(filePath); err != nil {
			log.Fatal(err)
		} else {
			defer f.Close()
		}
	}
}
