package agent

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

const contentType = "text/plain"
const updatePath = "update"

type HttpClient interface {
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type EmittingFunc func(metricType, name, value string)

func HttpEmittingFunc(addr string, client HttpClient) (EmittingFunc, error) {

	baseURL, err := url.ParseRequestURI(addr)
	if err != nil {
		return nil, err
	}

	emitter := func(metricType, name, value string) {
		reqURL := baseURL.JoinPath(updatePath, metricType, name, value).String()
		log.Println("POST", reqURL, "start")
		res, err := client.Post(reqURL, contentType, http.NoBody)
		if err != nil {
			log.Println("POST", reqURL, "error:", err)
			return
		}

		log.Println("POST", reqURL, "response status:", res.StatusCode)
	}

	return emitter, nil

}
