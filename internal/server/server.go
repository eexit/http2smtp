package server

import (
	"context"
	"fmt"
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

// goServer exposes the Go Server methods used by this package so
// it could be easier tested
type goServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
	BaseContext() context.Context
}

// serverWrapper is *http.Server wrapper that implements goServer.
// This is needed only to test the *http.Server.BaseContext() method.
type serverWrapper struct {
	*http.Server
}

func (sw *serverWrapper) BaseContext() context.Context {
	return sw.Server.BaseContext(nil)
}

// Server is the app entry point: it contains the HTTP server, config and services
type Server struct {
	svr               goServer
	logger            zerolog.Logger
	shutdownCtx       context.Context
	cancelFunc        context.CancelFunc
	smtpClient        smtp.Client
	converterProvider converter.Provider
	env               env.Bag
	sigint            chan os.Signal
}

// New returns a new http server for the API
func New(
	e env.Bag,
	logger zerolog.Logger,
	smtpClient smtp.Client,
	converterProvider converter.Provider,
) *Server {
	// This context will be used as a base context for all incoming
	// request. It is cancellable so when the server is shutting down,
	// we can propagate the cancellation signal to handlers and services
	ctx, cancel := context.WithCancel(context.Background())

	logger.Info().Msg("app is starting")

	svr := &Server{
		env:               e,
		cancelFunc:        cancel,
		logger:            logger,
		shutdownCtx:       ctx,
		smtpClient:        smtpClient,
		converterProvider: converterProvider,
		sigint:            make(chan os.Signal, 1),
	}

	// registers SIGINT channel
	signal.Notify(svr.sigint, os.Interrupt)

	svr.svr = &serverWrapper{
		&http.Server{
			Addr:         e.ServerHost + ":" + e.ServerPort,
			Handler:      svr.wrap(svr.routeHandler()),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		},
	}

	return svr
}

// Serve listens and serves for incoming HTTP request. It also handles
// graceful shutdown logic
func (s *Server) Serve() error {
	errch := make(chan error)

	go func(errch chan error) {
		s.logger.Info().Msgf("listening on %s:%s", s.env.ServerHost, s.env.ServerPort)
		if err := s.svr.ListenAndServe(); err != http.ErrServerClosed {
			// Error starting the listener:
			s.logger.Err(err).Msgf("server listening error: %s", err)
			errch <- err
			return
		}
		errch <- nil
	}(errch)

	for {
		select {
		// This catches server start error
		case err := <-errch:
			return err
		// This catches a OS signal (SIGINT)
		case <-s.sigint:
			s.logger.Info().Msg("closing server")

			s.cancelFunc()
			(s.smtpClient.Close()) // closes SMTP connection

			// server shutdown context
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(s.env.ServerShutdownTimeout)*time.Second)
			defer shutdownCancel()

			// We received an interrupt signal, shut down
			if err := s.svr.Shutdown(shutdownCtx); err != nil {
				// Error shutting down
				err = fmt.Errorf("server close error: %w", err)
				s.logger.Error().Msg(err.Error())
				return err
			}
			s.logger.Info().Msg("server closed")
			return nil
		}
	}
}
