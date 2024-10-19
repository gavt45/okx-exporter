package okx

import (
	"encoding/json"
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
	ChannelTickers     Channel = "tickers"
	ChannelInstruments Channel = "instruments"
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
}

type WSSubscriptionTopic struct {
	WSArgument
	InstId Instrument `json:"instId"`
}

// WSRequest request to okx wss API
type WSRequest struct {
	Op   Operation             `json:"op"`
	Args []WSSubscriptionTopic `json:"args"`
}

type TSms struct {
	time.Time
}

func (t *TSms) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// Fractional seconds are handled implicitly by Parse.
	i, err := strconv.ParseInt(string(data[1:len(data)-1]), 10, 64)
	if err != nil {
		panic(err)
	}

	update := time.UnixMilli(i)
	*t = TSms{update}
	return err
}

type WSDataTickers struct {
	InstType string     `json:"instType"`
	InstId   Instrument `json:"instId"`
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

// WSData a message from okx wss API
type WSData struct {
	Action `json:"action,omitempty"`
	Event  Operation         `json:"event,omitempty"`
	Arg    WSArgument        `json:"arg"`
	Data   []json.RawMessage `json:"data"`
}
