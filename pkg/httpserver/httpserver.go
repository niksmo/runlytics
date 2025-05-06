// Package httpserver provides http.Server wrapper object and usefull constants.
package httpserver

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/niksmo/runlytics/pkg/di"
)

// HTTPServer wrap the [http.Server] and provide Run method.
type HTTPServer struct {
	logger di.Logger
	http.Server
}

// New construncts the http.Server with passed parameters and returns HTTPServer pointer.
func New(addr string, handler http.Handler, logger di.Logger) *HTTPServer {
	return &HTTPServer{
		logger,
		http.Server{Addr: addr, Handler: handler},
	}
}

// Run start the http.Server and then waiting Done signal from context for gracefully shutdown.
func (server *HTTPServer) Run(stopCtx context.Context, wg *sync.WaitGroup) {
	server.logger.Infow("Listen", "host", server.Addr)
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-stopCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(), 3*time.Second,
		)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.logger.Infow("Server shutdown", "error", err)
		}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		server.logger.Errorw("Server closed with errors", "error", err)
	}
}
