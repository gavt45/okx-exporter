package core

import (
	"encoding/json"

	"github.com/gavt45/okx-exporter/pkg/core/domain/okx"
	"github.com/gavt45/okx-exporter/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	mLastPrice = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_price",
		},
		[]string{"instrument"},
	)
)

type Service struct{}

func (s *Service) ProcessMessage(data okx.WSData) error {
	if data.Event != okx.OperationEmpty {
		return nil // don't process callbacks
	}

	switch data.Arg.Channel {
	case okx.ChannelTickers:
		tickers := okx.WSDataTickers{}
		err := json.Unmarshal(data.Data[0], &tickers)
		if err != nil {
			return errors.Wrap(err, "can't parse data as data for tickers")
		}

		log.Info("Got tickers data: ", tickers)

		mLastPrice.WithLabelValues(string(tickers.InstId)).Set(tickers.LastFloat())
	default:
		log.Warn("Unknown channel: " + data.Arg.Channel)
	}

	return nil
}
