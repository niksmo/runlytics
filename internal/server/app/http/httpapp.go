package httpapp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api/httpapi"
	"github.com/niksmo/runlytics/internal/server/app/http/middleware"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

type App struct {
	mux  *chi.Mux
	addr *net.TCPAddr
	s    *http.Server
}

type AppParams struct {
	HTMLService        di.HTMLService
	UpdateService      di.UpdateService
	ReadService        di.ReadService
	HealthCheckService di.HealthCheckService
	BatchUpdateService di.BatchUpdateService
	Addr               *net.TCPAddr
	HashKey            string
	Decrypter          di.Decrypter
	TrustedNet         *net.IPNet
}

func New(p AppParams) *App {
	mux := chi.NewRouter()

	mux.Use(middleware.Logger)
	mux.Use(middleware.Decrypt(p.Decrypter))
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)

	if p.HashKey != "" {
		mux.Use(
			middleware.VerifyAndWriteSHA256(p.HashKey, http.MethodPost),
		)
	}

	if p.TrustedNet != nil {
		mux.Use(middleware.TrustedNet(p.TrustedNet))
	}

	httpapi.Register(
		mux,
		httpapi.RegisterServices{
			HTMLService:        p.HTMLService,
			UpdateService:      p.UpdateService,
			BatchUpdateService: p.BatchUpdateService,
			ReadService:        p.ReadService,
			HealthCheckService: p.HealthCheckService,
		},
	)

	return &App{mux: mux, addr: p.Addr}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	log := logger.Log.With(
		zap.String("op", op),
		zap.String("addr", a.addr.String()),
	)

	a.s = &http.Server{Addr: a.addr.String(), Handler: a.mux}

	log.Info("http server started")

	err := a.s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		log.Warn("http server listen error", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const (
		op      = "httpapp.Stop"
		timeout = 30 * time.Second
	)

	log := logger.Log.With(
		zap.String("op", op),
		zap.String("addr", a.addr.String()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info("http server stopping gracefully")
	err := a.s.Shutdown(ctx)
	if err != nil {
		log.Error("http server shutdown error", zap.Error(err))
	}
	log.Info("http server stopped")
}
