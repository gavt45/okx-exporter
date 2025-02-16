package core

import (
	"encoding/json"
	"time"

	"github.com/gavt45/okx-exporter/pkg/core/domain/okx"
	"github.com/gavt45/okx-exporter/pkg/log"
	"github.com/pkg/errors"
)

// List of channels required by core service
var RequiredChannels = []okx.Channel{
	okx.ChannelTickers,
	okx.ChannelCandle1H,
	okx.ChannelAggregatedTrades,
}

type Service struct{}

func (s Service) RequiredChannels() []okx.Channel {
	return RequiredChannels
}

func (s *Service) ProcessMessage(data okx.WSData) error {
	if data.Event != okx.OperationEmpty {
		return nil // don't process callbacks
	}

	now := time.Now()

	switch data.Arg.Channel { //nolint:exhaustive // instruments are not implemented yet, so we don't subscribe to them
	case okx.ChannelTickers:
		tickers := okx.WSDataTickers{}

		err := json.Unmarshal(data.Data[0], &tickers)
		if err != nil {
			return errors.Wrap(err, "can't parse data as data for tickers")
		}

		latency := now.Sub(tickers.TS.Time).Seconds()

		log.Info("Got tickers data: ", tickers)

		mLastPrice.WithLabelValues(string(tickers.InstID)).Set(tickers.LastFloat())
		mLatency.WithLabelValues(string(tickers.InstID)).Observe(float64(latency))
	case okx.ChannelCandle1H:
		candle1H := okx.WSDataCandle{}

		err := json.Unmarshal(data.Data[0], &candle1H)
		if err != nil {
			return errors.Wrap(err, "can't parse data as data for candle")
		}

		log.Info("Got 1H candle data: ", candle1H)

		mLastTS.WithLabelValues(string(data.Arg.InstID), okx.Candle1H).Set(float64(candle1H.TS.UnixMilli()))
		mLastOpen.WithLabelValues(string(data.Arg.InstID), okx.Candle1H).Set(candle1H.Open)
		mLastHigh.WithLabelValues(string(data.Arg.InstID), okx.Candle1H).Set(candle1H.High)
		mLastLow.WithLabelValues(string(data.Arg.InstID), okx.Candle1H).Set(candle1H.Low)
		mLastClose.WithLabelValues(string(data.Arg.InstID), okx.Candle1H).Set(candle1H.Close)
		mLastVolume.WithLabelValues(string(data.Arg.InstID), okx.Candle1H).Set(candle1H.Volume)
	case okx.ChannelAggregatedTrades:
		for _, tradeData := range data.Data {
			trade := okx.WSDataTrade{}

			if err := json.Unmarshal(tradeData, &trade); err != nil {
				return errors.Wrap(err, "can't parse data as data for trade")
			}

			log.Info("Got trade data: ", trade)

			mTradeSizeHist.WithLabelValues(string(data.Arg.InstID)).Observe(trade.SZFloat())
		}
	default:
		log.Warn("Unknown channel: " + data.Arg.Channel)
	}

	return nil
}
