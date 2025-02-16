package app

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gavt45/okx-exporter/pkg/core"
	"github.com/gavt45/okx-exporter/pkg/core/domain/okx"
	"github.com/gavt45/okx-exporter/pkg/log"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	OKXPublicPath string        = "/ws/v5/ipublic"
	ReadTimeout   time.Duration = 15 * time.Second
	PingInterval  time.Duration = 10 * time.Second
)

// Codes we consider irrecoverable, so we will crash when receiving them
var irrecoverableCodes = map[int]bool{
	websocket.CloseProtocolError:   true,
	websocket.CloseUnsupportedData: true,
	websocket.CloseMessageTooBig:   true,
	websocket.ClosePolicyViolation: true,
}

type RecieverApp struct {
	msgs chan okx.WSData

	cfg  *core.OKXConfig
	conn *websocket.Conn

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

func (a *RecieverApp) subscribeToChannel(instrument okx.Instrument, channel okx.Channel) error {
	err := a.conn.WriteJSON(&okx.WSRequest{
		Op: okx.OperationSubscribe,
		Args: []okx.WSSubscriptionTopic{
			{
				WSArgument: okx.WSArgument{
					Channel: channel,
				},
				InstID: instrument,
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "can't subscribe to "+string(channel))
	}

	return nil
}

// subscribeToRequiredChannels subscribes to all channels, required by core app
func (a *RecieverApp) subscribeToRequiredChannels() error {
	for _, channel := range a.svc.RequiredChannels() {
		err := a.subscribeToChannel(okx.InstrumentETHxUSDT, channel)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *RecieverApp) connect() error {
	var err error

	u := url.URL{Scheme: "wss", Host: a.cfg.WSHost, Path: OKXPublicPath}

	log.Debug("Dialing ", u.String())

	var resp *http.Response

	if a.conn, resp, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		return errors.Wrap(err, "can't dial websocket at "+u.String())
	}

	err = resp.Body.Close()
	if err != nil {
		return errors.Wrap(err, "can't close websocket response body")
	}

	err = a.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	if err != nil {
		return errors.Wrap(err, "can't set read deadline")
	}

	a.conn.SetPongHandler(func(string) error {
		log.Debug("Pong")
		return a.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	})

	log.Debug("Subscribing to updates")

	err = a.conn.SetWriteDeadline(time.Now().Add(ReadTimeout))
	if err != nil {
		return errors.Wrap(err, "can't set write deadline")
	}

	err = a.subscribeToRequiredChannels()
	if err != nil {
		return errors.Wrap(err, "can't subscribe to required channels on connect")
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
	ticker := time.NewTicker(PingInterval)

	for {
		select {
		case <-ticker.C:
			// Set write timeout before sending message
			if err := a.conn.SetWriteDeadline(time.Now().Add(ReadTimeout)); err != nil {
				return errors.Wrap(err, "can't set write deadline when pinging")
			}

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
	errs := make(chan error, 1)

	go func() {
		errs <- a.startProcessing(ctx)
	}()

	for {
		select {
		case err := <-errs:
			// Try to reconnect
			var netErr net.Error

			closeError := &websocket.CloseError{}

			if (errors.As(err, &netErr) && netErr.Timeout()) ||
				(errors.As(err, &closeError) && !irrecoverableCodes[closeError.Code]) {
				log.Debug("Collector is handling recoverable error: ", err.Error())

				if cerr := a.connect(); cerr != nil {
					return errors.Wrap(cerr, "can't connect")
				}

				go func() {
					errs <- a.startProcessing(ctx)
				}()

				log.Info("Reconnected")
			} else {
				log.Debug("Collector got irrecoveralbe error: ", err.Error())
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}
