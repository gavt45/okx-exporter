package okx

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

type Operation string

const (
	OperationLogin       Operation = "login"
	OperationSubscribe   Operation = "subscribe"
	OperationUnsubscribe Operation = "unsubscribe"
	OperationError       Operation = "error"
	OperationEmpty       Operation = ""
)

type Channel string

const (
	ChannelTickers          Channel = "tickers"
	ChannelInstruments      Channel = "instruments"
	ChannelCandle1H         Channel = "candle1H"
	ChannelAggregatedTrades Channel = "aggregated-trades"
)

// Used in okx_volume metric candle parameter
const (
	Candle1H string = "1H"
)

type Instrument string

const (
	InstrumentETHxUSDT Instrument = "ETH-USDT"
)

type Action string

const (
	ActionUpdate Action = "update"
)

type WSArgument struct {
	Channel `json:"channel"`
	InstID  Instrument `json:"instId,omitempty"`
}

type WSSubscriptionTopic struct {
	WSArgument
	InstID Instrument `json:"instId"`
}

// WSRequest request to okx wss API
type WSRequest struct {
	Op   Operation             `json:"op"`
	Args []WSSubscriptionTopic `json:"args"`
}

// TSms JSON unmarshallable timestamp in milliseconds
type TSms struct {
	time.Time
}

func TSmsFromString(str string) (TSms, error) {
	// Fractional seconds are handled implicitly by Parse.
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return TSms{}, err
	}

	update := time.UnixMilli(i)

	return TSms{update}, nil
}

func (t *TSms) UnmarshalJSON(data []byte) error {
	var err error

	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	*t, err = TSmsFromString(string(data[1 : len(data)-1]))

	return err
}

type WSDataTickers struct {
	InstType string     `json:"instType"`
	InstID   Instrument `json:"instId"`
	Last     string     `json:"last"`
	High24h  string     `json:"high24h"`
	Low24h   string     `json:"low24h"`
	TS       TSms       `json:"ts"`
}

func (t WSDataTickers) LastFloat() float64 {
	f, err := strconv.ParseFloat(t.Last, 64)
	if err != nil {
		panic(err)
	}

	return f
}

// Aggregated trades related structures
type Side string

const (
	SideSell Side = "sell"
	SideBuy  Side = "buy"
)

type WSDataTrade struct {
	FId    string     `json:"fId"`
	LId    string     `json:"lId"`
	InstID Instrument `json:"instId,omitempty"`
	PX     string     `json:"px"`
	Side   `json:"side"`
	SZ     string `json:"sz"`
	TS     TSms   `json:"ts"`
}

func (t WSDataTrade) SZFloat() float64 {
	f, err := strconv.ParseFloat(t.SZ, 64)
	if err != nil {
		panic(err)
	}

	return f
}

type WSDataCandle struct {
	TS     TSms    `json:"candleTimestamp"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

var ErrShortCandleDataArray error = errors.New("candle data array must have at least 6 values")

func (c *WSDataCandle) UnmarshalJSON(data []byte) error {
	var arr []string

	/*
		Example:
		[
			"1739685600000",
			"2709.13",
			"2720.65",
			"2706.99",
			"2718.45",
			"1407.871749",
			"3820532.8437713",
			"3820532.8437713",
			"0"
		]
	*/

	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	if len(arr) < 6 {
		return ErrShortCandleDataArray
	}

	c.TS, err = TSmsFromString(arr[0])
	if err != nil {
		return err
	}

	c.Open, err = strconv.ParseFloat(arr[1], 64)
	if err != nil {
		return err
	}

	c.High, err = strconv.ParseFloat(arr[2], 64)
	if err != nil {
		return err
	}

	c.Low, err = strconv.ParseFloat(arr[3], 64)
	if err != nil {
		return err
	}

	c.Close, err = strconv.ParseFloat(arr[4], 64)
	if err != nil {
		return err
	}

	c.Volume, err = strconv.ParseFloat(arr[5], 64)

	return err
}

// WSData a message from okx wss API
type WSData struct {
	Action `json:"action,omitempty"`
	Event  Operation         `json:"event,omitempty"`
	Arg    WSArgument        `json:"arg"`
	Data   []json.RawMessage `json:"data"`
}
