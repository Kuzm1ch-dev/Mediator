package client

import (
	"bytes"
	"crypto/tls"
	"log"
	"net/http"
)

type Http struct {
	Client *http.Client
}

func InitHttp() *Http {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	return &Http{httpClient}
}

func (h *Http) Do(url string, data []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	log.Println(req.URL)
	return h.Client.Do(req)
}

func (h *Http) DoContext(req *http.Request) (*http.Response, error) {
	return h.Client.Do(req)
}
