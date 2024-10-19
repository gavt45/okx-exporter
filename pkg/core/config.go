package core

type OKXConfig struct {
	WSHost string `json:"ws_host" yaml:"ws_host" config:"ws_host"`
}

type ServiceConfig struct {
	Host string    `json:"host" yaml:"host" config:"host" validate:"required"`
	Port int       `json:"port" yaml:"port" config:"port" validate:"required"`
	OKX  OKXConfig `json:"okx" yaml:"okx" config:"okx"`
}
