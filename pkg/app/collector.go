package app

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/gavt45/okx-exporter/pkg/core"
	"github.com/gavt45/okx-exporter/pkg/core/domain/okx"
	"github.com/gavt45/okx-exporter/pkg/dao"
	"github.com/gavt45/okx-exporter/pkg/log"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	OKX_PUBLIC_PATH string        = "/ws/v5/ipublic"
	READ_TIMEOUT    time.Duration = 15 * time.Second
	PING_INTERVAL   time.Duration = 10 * time.Second
)

type RecieverApp struct {
	msgs chan okx.WSData

	cfg  *core.OKXConfig
	conn *websocket.Conn
	repo dao.MetricsRepository

	svc *core.Service
}

func NewRecieverApp(cfg *core.OKXConfig) (*RecieverApp, error) {
	var err error

	app := &RecieverApp{
		cfg:  cfg,
		msgs: make(chan okx.WSData, 100),
		svc:  &core.Service{},
	}

	err = app.connect()
	if err != nil {
		return nil, err
	}

	return app, err
}

func (a *RecieverApp) WithMetricsRepo(repo dao.MetricsRepository) *RecieverApp {
	a.repo = repo
	return a
}

func (a *RecieverApp) connect() error {
	var err error

	u := url.URL{Scheme: "wss", Host: a.cfg.WSHost, Path: OKX_PUBLIC_PATH}

	log.Debug("Dialing ", u.String())
	a.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "can't dial websocket at "+u.String())
	}

	a.conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))
	a.conn.SetPongHandler(func(string) error {
		log.Debug("Pong")
		a.conn.SetReadDeadline(time.Now().Add(READ_TIMEOUT))
		return nil
	})

	log.Debug("Subscribing to updates")
	a.conn.SetWriteDeadline(time.Now().Add(READ_TIMEOUT))
	err = a.conn.WriteJSON(&okx.WSRequest{
		Op: okx.OperationSubscribe,
		Args: []okx.WSSubscriptionTopic{
			{
				WSArgument: okx.WSArgument{
					Channel: okx.ChannelTickers,
				},
				InstId: okx.InstrumentETHxUSDT,
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "can't send initial message to "+u.String())
	}

	log.Debug("Connected to ", u.String())

	return nil
}

func (a *RecieverApp) reader(ctx context.Context) error {
	done := false
	for !done {
		msg := okx.WSData{}

		err := a.conn.ReadJSON(&msg)
		if err != nil {
			log.Debug("Got read error: ", err.Error())
			return err
		}

		a.msgs <- msg

		select {
		case <-ctx.Done():
			done = true
		default:
		}
	}

	log.Debug("Reader exiting")

	return nil
}

func (a *RecieverApp) pinger(ctx context.Context) error {
	ticker := time.NewTicker(PING_INTERVAL)

	for {
		select {
		case <-ticker.C:
			// Set write timeout before sending message
			a.conn.SetWriteDeadline(time.Now().Add(READ_TIMEOUT))
			log.Debug("Ping")
			if err := a.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Debug("Got write error: ", err.Error())
				ticker.Stop()
				return err
			}
		case <-ctx.Done():
			ticker.Stop()
			log.Debug("Pinger exiting")
			return nil
		}
	}
}

func (a *RecieverApp) processor(ctx context.Context) error {
	for {
		select {
		case msg := <-a.msgs:
			err := a.svc.ProcessMessage(msg)
			if err != nil {
				log.Debug("Got process error: ", err.Error())
				return err
			}
		case <-ctx.Done():
			log.Debug("Processor exiting")
			return nil
		}
	}
}

func (a *RecieverApp) startProcessing(ctx context.Context) error {
	g, connCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.reader(connCtx)
	})

	g.Go(func() error {
		return a.pinger(connCtx)
	})

	g.Go(func() error {
		return a.processor(connCtx)
	})

	return g.Wait()
}

func (a *RecieverApp) Start(ctx context.Context) error {
	errs := make(chan error)

	go func() {
		errs <- a.startProcessing(ctx)
	}()

	for {
		select {
		case err := <-errs:
			// Try to reconnect
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				log.Debug("Collector is handling timeout error: ", err.Error())

				err := a.connect()
				if err != nil {
					return errors.Wrap(err, "can't connect")
				}

				go func() {
					errs <- a.startProcessing(ctx)
				}()

				log.Info("Reconnected")
			} else {
				log.Debug("Collector got unknown error: ", err.Error())
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}
