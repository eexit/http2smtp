package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/env"
	"github.com/eexit/http2smtp/internal/server"
	"github.com/eexit/http2smtp/internal/smtp"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var e env.Bag
	envconfig.MustProcess("", &e)

	level, err := zerolog.ParseLevel(e.LogLevel)
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Str("version", server.Version).
		Logger()

	smtpClient := smtp.New(e.SMTPAddr, logger)

	converterProvider := converter.NewProvider(
		converter.NewRFC5322(),
		converter.NewSparkPost(),
	)

	server := server.New(e, logger, smtpClient, converterProvider)
	if err := server.Serve(); err != nil {
		panic(err)
	}
	os.Exit(0)
}
