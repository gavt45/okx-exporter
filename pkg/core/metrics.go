package core

import (
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

	mLastTS = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_candle_ts",
			Help: "Candle timestamp",
		},
		[]string{"instrument", "candle"},
	)

	mLastOpen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_open",
			Help: "Open price got from candleXX message i.e candle1H",
		},
		[]string{"instrument", "candle"},
	)

	mLastHigh = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_high",
			Help: "High price got from candleXX message i.e candle1H",
		},
		[]string{"instrument", "candle"},
	)

	mLastLow = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_low",
			Help: "Low price got from candleXX message i.e candle1H",
		},
		[]string{"instrument", "candle"},
	)

	mLastClose = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_close",
			Help: "Close price got from candleXX message i.e candle1H",
		},
		[]string{"instrument", "candle"},
	)

	mLastVolume = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "okx_volume",
			Help: "Volume got from candleXX message i.e candle1H",
		},
		[]string{"instrument", "candle"},
	)

	mLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "okx_latency",
		},
		[]string{"instrument"},
	)

	mTradeSizeHist = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "okx_trade_size",
			Buckets: []float64{0.001, 0.01, 0.1, 1, 5, 10, 100, 1000, 10000},
		},
		[]string{"instrument"},
	)
)
