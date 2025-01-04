package agent

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
)

const contentType = "text/plain"
const updatePath = "update"

var (
	ErrParse  = errors.New("used unnecessary url symbols")
	ErrScheme = errors.New("supports only http or https scheme")
	ErrHost   = errors.New("pass absolute url with host, e.g. http://one.two.com/path1/id")
)

type HttpClient interface {
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
}

type EmittingFunc func(metricType, name, value string)

func HttpEmittingFunc(addr string, client HttpClient) (EmittingFunc, error) {

	baseURL, err := url.ParseRequestURI(addr)
	if err != nil {
		return nil, errors.Join(ErrParse, err)
	}

	if !(baseURL.Scheme == "http" || baseURL.Scheme == "https") {
		return nil, ErrScheme
	}

	if baseURL.Host == "" {
		return nil, ErrHost
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
