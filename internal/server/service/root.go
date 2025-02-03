package service

import (
	"bytes"
	"fmt"
	"html/template"
	"slices"
	"strconv"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type ReadRepository interface {
	ReadCounter() (map[string]int64, error)
	ReadGauge() (map[string]float64, error)
}

type HTMLService struct {
	template   *template.Template
	repository ReadRepository
}

func NewHTMLService(repository ReadRepository) *HTMLService {
	text := `<!DOCTYPE html>
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
	return &HTMLService{
		template:   template.Must(template.New("main").Parse(text)),
		repository: repository,
	}
}

func (s *HTMLService) RenderMetricsList(buf *bytes.Buffer) error {
	gauge := s.repository.ReadGauge()
	counter := s.repository.ReadCounter()
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

	if err := s.template.Execute(buf, render); err != nil {
		logger.Log.Error("Error metrics list template executing", zap.Error(err))
		return err
	}
	return nil
}
