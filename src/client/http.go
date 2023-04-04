package client

import (
	"bytes"
	"crypto/tls"
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

func (h *Http) Do(url string, data []byte, requiredHeaders map[string]string, additionalHeaders map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if len(requiredHeaders) > 0 {
		for key, value := range requiredHeaders {
			req.Header.Add(key, value)
		}
	}
	if len(additionalHeaders) > 0 {
		for key, value := range additionalHeaders {
			req.Header.Add(key, value)
		}
	}
	return h.Client.Do(req)
}

func (h *Http) DoContext(req *http.Request) (*http.Response, error) {
	return h.Client.Do(req)
}
