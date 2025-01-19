package router

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

const text = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>runlytics</title>
</head>
<body>
  <h1>Metrics list</h1>
  {{if .}}
  <ul>
  {{range .}}
    <li>{{.}}</li>
  {{end}}
  </ul>
  {{else}}
  <p>metrics data is empty</p>
  {{end}}
</body>
</html>
`

func SetMainRoute(r *chi.Mux, repo server.RepoRead) {
	h := &mainHandler{repo}
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.getHandleFunc())
		logRegister("/")
	})
}

type mainHandler struct {
	repo server.RepoRead
}

func (h *mainHandler) getHandleFunc() http.HandlerFunc {
	t := template.Must(template.New("main").Parse(text))

	return func(w http.ResponseWriter, r *http.Request) {
		gauge, counter := h.repo.GetData()
		render := make([]string, 0, len(gauge)+len(counter))

		for k, v := range gauge {
			vs := strconv.FormatFloat(v, 'f', -1, 64)
			render = append(render, fmt.Sprintf("%s: %s", k, vs))
		}

		for k, v := range counter {
			vs := strconv.FormatInt(v, 10)
			render = append(render, fmt.Sprintf("%s: %s", k, vs))
		}

		slices.Sort(render)

		var b bytes.Buffer
		if err := t.Execute(&b, render); err != nil {
			logger.Log.Panic("Render template error", zap.Error(err))
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		b.WriteTo(w)
	}
}
