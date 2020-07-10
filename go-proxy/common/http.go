package common

import (
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() *HTTPClient {

	path := getEnvWithDefault("PROXY", "socks5://127.0.0.1:9050")
	proxy, err := url.Parse(path)
	if err != nil {
		log.Fatalf("failed to parse proxy URL '%s' with '%v'", path, err)
	}

	return &HTTPClient{client: &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxy)},
		Timeout:   time.Second * 18}}
}

func (c HTTPClient) Get(url string) (*http.Response, error) {

	return c.client.Get(url)
}

func getEnvWithDefault(variable, defaultValue string) string {

	ret := os.Getenv(variable)
	if ret == "" {
		ret = defaultValue
	}
	log.Infof("'%s': '%s'", variable, ret)

	return ret
}
