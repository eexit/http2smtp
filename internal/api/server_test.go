package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/env"
	"github.com/eexit/http2smtp/internal/smtp"
	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	got := New(
		env.Bag{
			SMTPAddr: "smtp:25",
			LogLevel: "info",
		},
		zerolog.New(io.Discard),
		&smtp.Stub{},
		converter.NewProvider(),
	)
	want := &API{
		env: env.Bag{
			SMTPAddr: "smtp:25",
			LogLevel: "info",
		},
		logger:            zerolog.New(io.Discard),
		smtpClient:        &smtp.Stub{},
		converterProvider: converter.NewProvider(),
		svr:               &serverWrapper{&http.Server{Addr: ":80"}},
	}

	if !reflect.DeepEqual(got.logger, want.logger) {
		t.Errorf("logger = %#v, want %#v", got.logger, want.logger)
	}

	if !reflect.DeepEqual(got.env, want.env) {
		t.Errorf("env = %#v, want %#v", got.env, want.env)
	}

	if !reflect.DeepEqual(got.smtpClient, want.smtpClient) {
		t.Errorf("smtpClient = %#v, want %#v", got.smtpClient, want.smtpClient)
	}

	if !reflect.DeepEqual(got.converterProvider, want.converterProvider) {
		t.Errorf("converterProvider = %#v, want %#v", got.converterProvider, want.converterProvider)
	}

	if got.shutdownCtx != got.svr.BaseContext() {
		t.Errorf("shutdownCtx should be equal to server base context")
	}
}

func TestAPI_Serve(t *testing.T) {
	type fields struct {
		svr        goServer
		smtpClient smtp.Client
	}
	tests := []struct {
		name    string
		fields  fields
		sigint  bool
		wantErr bool
	}{
		{
			name: "serve returns an error",
			fields: fields{
				svr: &stubServer{serveErr: errors.New("serving error")},
			},
			sigint:  false,
			wantErr: true,
		},
		{
			name: "serve no error",
			fields: fields{
				svr: &stubServer{serveTimeoutAfter: 100 * time.Millisecond},
			},
			sigint:  false,
			wantErr: false,
		},
		{
			name: "serve shutdown no error",
			fields: fields{
				svr:        &stubServer{serveTimeoutAfter: 100 * time.Millisecond},
				smtpClient: &smtp.Stub{},
			},
			sigint:  true,
			wantErr: false,
		},
		{
			name: "serve shutdown no error",
			fields: fields{
				svr:        &stubServer{shutdownErr: errors.New("shutdown error")},
				smtpClient: &smtp.Stub{},
			},
			sigint:  true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdownCtx, cancelFunc := context.WithCancel(context.Background())

			a := &API{
				svr:         tt.fields.svr,
				env:         env.Bag{ServerShutdownTimeout: 0},
				shutdownCtx: shutdownCtx,
				logger:      zerolog.New(io.Discard), // replace by os.Stdout for debudding
				cancelFunc:  cancelFunc,
				smtpClient:  tt.fields.smtpClient,
				sigint:      make(chan os.Signal, 1),
			}

			signal.Notify(a.sigint, syscall.SIGUSR1)

			done := make(chan bool)

			go func() {
				if err := a.Serve(); (err != nil) != tt.wantErr {
					t.Errorf("Server.Serve() error = %v, wantErr %v", err, tt.wantErr)
				}
				done <- true
			}()

			if tt.sigint {
				a.sigint <- syscall.SIGUSR1
			}

			<-done
		})
	}
}

type stubServer struct {
	serveTimeoutAfter     time.Duration
	serveErr, shutdownErr error
}

func (s *stubServer) ListenAndServe() error {
	if s.serveErr != nil {
		return s.serveErr
	}
	time.Sleep(s.serveTimeoutAfter)
	return http.ErrServerClosed
}

func (s *stubServer) Shutdown(context.Context) error {
	return s.shutdownErr
}

func (s *stubServer) BaseContext() context.Context {
	return context.Background()
}
