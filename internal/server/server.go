package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/env"
	"github.com/eexit/http2smtp/internal/smtp"
	"github.com/rs/zerolog"
)

// Version receives its value at compile time
var Version string

// Server is the app entry point: it contains the HTTP server, config and services
type Server struct {
	svr               *http.Server
	logger            zerolog.Logger
	shutdownCtx       context.Context
	cancelFunc        context.CancelFunc
	smtpClient        smtp.Client
	converterProvider converter.Provider
	env               env.Bag
}

// New returns a new http server for the API
func New(e env.Bag) *Server {
	// This context will be used as a base context for all incoming
	// request. It is cancellable so when the server is shutting down,
	// we can propagate the cancellation signal to handlers and services
	ctx, cancel := context.WithCancel(context.Background())

	level, err := zerolog.ParseLevel(e.LogLevel)
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().
		Str("version", Version).
		Logger()

	logger.Info().Msg("app is starting")

	smtpClient := smtp.New(e.SMTPAddr, logger)

	converterProvider := converter.NewProvider(
		converter.NewRFC5322(),
		converter.NewSparkPost(),
	)

	svr := &Server{
		cancelFunc:        cancel,
		logger:            logger,
		shutdownCtx:       ctx,
		smtpClient:        smtpClient,
		converterProvider: converterProvider,
		env:               e,
	}

	svr.svr = &http.Server{
		Addr:         e.ServerHost + ":" + e.ServerPort,
		Handler:      svr.wrap(svr.routeHandler()),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	return svr
}

// Serve listens and serves for incoming HTTP request. It also handles
// graceful shutdown logic
func (s *Server) Serve() error {
	go func() {
		s.logger.Info().Msgf("listening on http%s", s.svr.Addr)
		if err := s.svr.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting or closing listener:
			s.logger.Panic().Err(err).Msgf("server shutdown error: %s", err)
		}
	}()

	// registers SIGINT channel
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint
	s.logger.Info().Msg("shutting down")

	s.cancelFunc()
	s.smtpClient.Close() // closes SMTP connection

	// server shutdown context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(s.env.ServerShutdownTimeout)*time.Second)
	defer shutdownCancel()

	// We received an interrupt signal, shut down
	if err := s.svr.Shutdown(shutdownCtx); err != nil {
		return err
	}
	s.logger.Info().Msg("server stopped")
	return nil
}
