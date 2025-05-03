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
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

// HTMLService working with prepared html template and repository
type HTMLService struct {
	template   *template.Template
	repository di.ReadListRepository
}

// NewHTMLService returns HTMLService pointer
func NewHTMLService(repository di.ReadListRepository) *HTMLService {
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

// RenderMetricsList writes HTML page to buffer or returns error if occured.
func (s *HTMLService) RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error {
	counter, gauge, err := s.readMetrics(ctx)
	if err != nil {
		return err
	}
	renderList := s.makeRenderList(counter, gauge)
	return s.renderTemplate(renderList, buf)
}

func (s *HTMLService) readMetrics(
	ctx context.Context,
) (
	counter map[string]int64,
	gauge map[string]float64,
	readErr error,
) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		g, err := s.repository.ReadGauge(ctx)
		if err != nil {
			readErr = err
			logger.Log.Error("Read all gauge metrics", zap.Error(err))
			return
		}
		gauge = g
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := s.repository.ReadCounter(ctx)
		if err != nil {
			readErr = err
			logger.Log.Error("Read all counter metrics", zap.Error(err))
			return
		}
		counter = c
	}()
	wg.Wait()
	return
}

func (s *HTMLService) makeRenderList(
	counter map[string]int64, gauge map[string]float64,
) []string {
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
	return render
}

func (s *HTMLService) renderTemplate(
	renderList []string, buf *bytes.Buffer,
) error {
	if err := s.template.Execute(buf, renderList); err != nil {
		logger.Log.Error(
			"Error metrics list template executing", zap.Error(err),
		)
		return err
	}
	return nil
}
