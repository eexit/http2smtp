package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/eexit/http2smtp/internal/api"
	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/env"
	"github.com/eexit/http2smtp/internal/smtp"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

var adapter *httpadapter.HandlerAdapter

func init() {
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
		Str("version", api.Version).
		Logger()

	smtpClient := smtp.New(e.SMTPAddr, logger)

	converterProvider := converter.NewProvider(
		converter.NewRFC5322(),
		converter.NewSparkPost(),
	)

	app := api.New(e, logger, smtpClient, converterProvider)
	adapter = httpadapter.New(app.Wrap(app.Mux()))
}

// Handler handles the function invocation
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
