package agent

import (
	"log"
	"net/http"
	"net/url"
)

const contentType = "text/plain"
const updatePath = "update"

type HttpEmitter func(mType, name, value string)

func NewHttpEmitter(addr string) (HttpEmitter, error) {

	baseURL, err := url.ParseRequestURI(addr)
	if err != nil {
		return nil, err
	}

	emitter := func(mType, name, value string) {
		reqURL := baseURL.JoinPath(updatePath, mType, name, value).String()
		log.Println("POST", reqURL, "start")
		res, err := http.Post(reqURL, contentType, http.NoBody)
		if err != nil {
			log.Println("POST", reqURL, "error:", err)
			return
		}

		log.Println("POST", reqURL, "response status:", res.StatusCode)
	}

	return emitter, nil

}
