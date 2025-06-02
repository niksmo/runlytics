package httpapi_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/server/api/httpapi"

	"github.com/stretchr/testify/mock"
)

// ExampleHTMLService represents mocked HTMLService project component.
type ExampleHTMLService struct {
	mock.Mock
}

// RenderMetricsList writes html page to buffer.
func (s *ExampleHTMLService) RenderMetricsList(ctx context.Context, buf *bytes.Buffer) error {
	getArgs := s.Called(context.Background(), buf)
	buf.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Runlytics</title>
</head>
<body>
    <p>Alloc: 123.45</p>
    <!-- other metrics -->
</body>
</html>`)
	return getArgs.Error(0)
}

func ExampleSetHTMLHandler() {
	buf := new(bytes.Buffer)
	HTMLService := new(ExampleHTMLService)
	HTMLService.On("RenderMetricsList", context.Background(), buf).Return(nil)

	mux := chi.NewRouter()
	httpapi.SetHTMLHandler(mux, HTMLService)
	s := httptest.NewServer(mux)
	defer s.Close()

	req, err := http.NewRequest(http.MethodGet, s.URL+"/", http.NoBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := s.Client().Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)
	// Output:
	// 200
	// <!DOCTYPE html>
	// <html lang="en">
	// <head>
	//     <meta charset="UTF-8">
	//     <meta name="viewport" content="width=device-width, initial-scale=1.0">
	//     <title>Runlytics</title>
	// </head>
	// <body>
	//     <p>Alloc: 123.45</p>
	//     <!-- other metrics -->
	// </body>
	// </html>
}
