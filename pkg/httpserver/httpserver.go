package httpserver

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/niksmo/runlytics/pkg/di"
)

type HTTPServer struct {
	http.Server
	logger di.Logger
}

func New(addr string, handler http.Handler, logger di.Logger) *HTTPServer {
	return &HTTPServer{
		http.Server{Addr: addr, Handler: handler},
		logger,
	}
}

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
