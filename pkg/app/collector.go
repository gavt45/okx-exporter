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

	return nil
}

func (a *RecieverApp) reader() error {
	for {
		msg := okx.WSData{}

		err := a.conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		// Try to write and return if channel is closed
		select {
		case a.msgs <- msg:
		default:
			return nil
		}
	}
}

func (a *RecieverApp) Start(ctx context.Context) error {
	errs := make(chan error)
	ticker := time.NewTicker(PING_INTERVAL)

	readerThread := func() { errs <- a.reader() }

	go readerThread()

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Debug("Ping")
				if err := a.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-a.msgs:
				err := a.svc.ProcessMessage(msg)
				if err != nil {
					errs <- err
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case err := <-errs:
		// Try to reconnect
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			err = a.connect()
			if err != nil {
				return errors.Wrap(err, "can't reconnect")
			}

			go readerThread()

			log.Info("Reconnected")
		}
		return err
	case <-ctx.Done():
		close(a.msgs)
		return nil
	}
}
