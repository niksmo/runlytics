// Package httpserver provides http.Server wrapper object and usefull constants.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HTTPServer wrap the [http.Server] and provide Run method.
type HTTPServer struct {
	http.Server
}

// New construncts the http.Server with passed parameters and returns HTTPServer pointer.
func New(addr string, handler http.Handler) *HTTPServer {
	return &HTTPServer{
		http.Server{Addr: addr, Handler: handler},
	}
}

// Run start the http.Server and then waiting Done signal from context for gracefully shutdown.
func (s *HTTPServer) Run(stopCtx context.Context, wg *sync.WaitGroup) {
	fmt.Println("Listen", "host", s.Addr)
	wg.Add(1)
	go waitForShutdown(stopCtx, wg, s)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed with errors", "error", err)
	}
}

func waitForShutdown(stopCtx context.Context, wg *sync.WaitGroup, s *HTTPServer) {
	defer wg.Done()
	<-stopCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), 3*time.Second,
	)
	defer cancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Server shutdown", "error", err)
	}
}
