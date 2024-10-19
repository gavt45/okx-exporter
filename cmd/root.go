package main

import (
	"context"

	"github.com/gavt45/okx-exporter/pkg/app"
	"github.com/gavt45/okx-exporter/pkg/core"
	"github.com/gavt45/okx-exporter/pkg/log"
	validator "github.com/go-playground/validator/v10"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"
)

func main() {
	log.Info("Starting okx exporter")

	loader := confita.NewLoader(
		env.NewBackend(),
		file.NewOptionalBackend("config.yaml"),
		file.NewOptionalBackend("config/config.yaml"),
		flags.NewBackend(),
	)

	cfg := core.ServiceConfig{}

	err := loader.Load(context.Background(), &cfg)
	if err != nil {
		log.Fatal("Can't load config: ", err.Error())
		return
	}

	v := validator.New()
	err = v.Struct(cfg)
	if err != nil {
		log.Fatal("Can't load config: ", err.Error())
		return
	}

	app, err := app.New(cfg)
	if err != nil {
		log.Fatal("Can't create app: ", err.Error())
		return
	}

	err = app.Start(context.Background())
	if err != nil {
		log.Fatal("App error: ", err.Error())
		return
	}

	log.Info("Done!")
}
