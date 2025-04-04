package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"slices"
	"strconv"
	"sync"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

type HTMLService struct {
	template   *template.Template
	repository di.ReadRepository
}

func NewHTMLService(repository di.ReadRepository) *HTMLService {
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

func (s *HTMLService) RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error {
	var (
		wg       sync.WaitGroup
		gauge    map[string]float64
		counter  map[string]int64
		errSlice []error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		gauge, err = s.repository.ReadGauge(ctx)
		if err != nil {
			errSlice = append(errSlice, err)
			logger.Log.Error("Read all gauge metrics", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		counter, err = s.repository.ReadCounter(ctx)
		if err != nil {
			errSlice = append(errSlice, err)
			logger.Log.Error("Read all counter metrics", zap.Error(err))
		}
	}()

	wg.Wait()

	if len(errSlice) != 0 {
		return server.ErrInternal
	}

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
