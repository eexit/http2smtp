package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/eexit/httpsmtp/internal/env"
	"github.com/eexit/httpsmtp/internal/server/handler"
	"github.com/eexit/httpsmtp/internal/smtp"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Version receives its from compile time
var Version string

// Server is the app entry point: it contains the HTTP server, config and services
type Server struct {
	svr         *http.Server
	logger      zerolog.Logger
	shutdownCtx context.Context
	cancelFunc  context.CancelFunc
	smtpClient  *smtp.SMTP
	env         env.Bag
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

	smtpClient := smtp.NewSMTP(e.SMTPHost, e.SMTPPort, logger)

	svr := &Server{
		cancelFunc:  cancel,
		logger:      logger,
		shutdownCtx: ctx,
		smtpClient:  smtpClient,
		env:         e,
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

// routeHandler returns the app routes
func (s *Server) routeHandler() http.Handler {
	r := mux.NewRouter()

	r.Handle("/healthcheck", handler.Healthcheck()).
		Methods(http.MethodHead, http.MethodGet)

	r.Handle("/sparkpost/api/v1/transmissions", handler.SparkPost(s.smtpClient)).
		Methods(http.MethodPost)

	return r
}

func (s *Server) wrap(fn http.Handler) http.Handler {
	return alice.New().
		Append(
			hlog.NewHandler(s.logger),
			hlog.CustomHeaderHandler("trace_id", s.env.HTTPTraceHeader),
			hlog.MethodHandler("verb"),
			hlog.RemoteAddrHandler("ip"),
			hlog.UserAgentHandler("user_agent"),
			hlog.URLHandler("url"),
			contentTypeResponseHandler("application/json"),
			hlog.AccessHandler(func(r *http.Request, code, size int, duration time.Duration) {
				var level zerolog.Level
				switch {
				case code < 300:
					level = zerolog.InfoLevel
				case code >= 300 && code < 400:
					level = zerolog.WarnLevel
				case code >= 400 && code < 500:
					level = zerolog.ErrorLevel
				case code > 500:
					level = zerolog.FatalLevel
				}
				hlog.FromRequest(r).
					WithLevel(level).
					Int("code", code).
					Int("size", size).
					Dur("duration", duration).
					Msg("served request")
			}),
		).Then(fn)
}

func contentTypeResponseHandler(contentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}
