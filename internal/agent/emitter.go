package agent

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

const contentType = "text/plain"
const updatePath = "update"

type HTTPClient interface {
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type EmittingFunc func(metricType, name, value string)

func HTTPEmittingFunc(baseURL *url.URL, client HTTPClient) (EmittingFunc, error) {
	log.Print("Ready for emitting on host ", baseURL)
	emitter := func(metricType, name, value string) {
		reqURL := baseURL.JoinPath(updatePath, metricType, name, value).String()
		log.Println("POST", reqURL, "start")
		res, err := client.Post(reqURL, contentType, http.NoBody)
		if err != nil {
			log.Println("POST", reqURL, "error:", err)
			return
		}
		defer res.Body.Close()

		log.Println("POST", reqURL, "response status:", res.StatusCode)
	}

	return emitter, nil
}
