package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gavt45/okx-exporter/pkg/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

type App interface {
	Start(ctx context.Context) error
}

type MetricsApp struct {
	reciever *RecieverApp
	cfg      core.ServiceConfig
}

func New(cfg core.ServiceConfig) (App, error) {
	var err error

	app := &MetricsApp{cfg: cfg}

	app.reciever, err = NewRecieverApp(&cfg.OKX)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (a *MetricsApp) Start(_ctx context.Context) error {
	grp, ctx := errgroup.WithContext(_ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	grp.Go(func() error {
		select {
		case sig := <-sigs:
			return fmt.Errorf("stopping due to signal %d", sig)
		case <-ctx.Done():
			return nil
		}
	})

	grp.Go(func() error {
		return a.reciever.Start(ctx)
	})

	grp.Go(func() error {
		errs := make(chan error)
		srv := &http.Server{Addr: fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)}

		http.Handle("/metrics", promhttp.Handler())

		go func() {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				errs <- err
			}
		}()

		select {
		case <-ctx.Done():
			// parent context is closed, so we create a new one here
			ctxTimeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
			return srv.Shutdown(ctxTimeout)
		case err := <-errs:
			return err
		}
	})

	return grp.Wait()
}
