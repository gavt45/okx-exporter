package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	receiver *RecieverApp
	cfg      core.ServiceConfig
}

func New(cfg core.ServiceConfig) (App, error) {
	var err error

	app := &MetricsApp{cfg: cfg}

	app.receiver, err = NewRecieverApp(&cfg.OKX)
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
		return a.receiver.Start(ctx)
	})

	grp.Go(func() error {
		// Capacity 1, so read is not blocked when channel is empty
		errs := make(chan error, 1)
		srv := &http.Server{
			Addr:              net.JoinHostPort(a.cfg.Host, strconv.Itoa(a.cfg.Port)),
			ReadTimeout:       10 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		}

		http.Handle("/metrics", promhttp.Handler())

		go func() {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				errs <- err
			}
		}()

		select {
		case <-ctx.Done():
			// parent context is closed, so we create a new one here
			// ignore cancel here as it is not needed
			ctxTimeout, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint:govet
			return srv.Shutdown(ctxTimeout)
		case err := <-errs:
			return err
		}
	})

	return grp.Wait()
}
