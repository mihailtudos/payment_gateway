package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/adapters/bank/montebank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/config"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/httputil"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/pkg/tel"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
)

type API struct {
	router        *chi.Mux
	paymentsRepo  *repository.PaymentsRepository
	acquiringBank bank.Adapter
	telemetry     *tel.Telemetry
}

func New(acquiringBankURL string, telemetry *tel.Telemetry) *API {
	a := &API{
		telemetry: telemetry,
	}
	a.paymentsRepo = repository.NewPaymentsRepository()
	a.acquiringBank = montebank.NewHTTPBankAdapter(montebank.DefaultConfig(acquiringBankURL))

	a.setupRouter()
	return a
}

func (a *API) Handler() http.Handler { return a.router }

func (a *API) Run(ctx context.Context, cfg config.HTTP) error {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           a.router,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:       time.Duration(cfg.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		slog.InfoContext(ctx, "shutting down HTTP server")
		return httpServer.Shutdown(ctx)
	})

	g.Go(func() error {
		slog.InfoContext(ctx, "starting HTTP server", slog.String("addr", httpServer.Addr))
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (a *API) setupRouter() {
	a.router = chi.NewRouter()

	a.router.Use(middleware.RealIP)
	a.router.Use(httputil.RequestID)
	a.router.Use(httputil.RequestLogger)
	a.router.Use(middleware.Recoverer)
	a.router.Use(middleware.Timeout(30 * time.Second))
	a.router.Use(middleware.StripSlashes)

	a.router.Get("/ping", a.PingHandler())
	a.router.Get("/swagger/*", a.SwaggerHandler())

	a.router.Route("/api", func(r chi.Router) {
		// Payment routes
		r.Get("/payments/{id}", a.GetPaymentHandler())
		r.Post("/payments", a.PostPaymentHandler())
	})

}
