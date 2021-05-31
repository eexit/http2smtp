package smtp

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/ctx"
	"github.com/eexit/http2smtp/pkg/testutil"
	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	t.Run("constructor fails to dial to SMTP server", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected to panic")
			}
		}()

		New("::", zerolog.New(ioutil.Discard))
	})

	t.Run("new client dials ok", func(t *testing.T) {
		ln := newLocalListener(t)
		defer ln.Close()

		go New(ln.Addr().String(), zerolog.New(ioutil.Discard))

		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("failed to accept connection: %v", err)
		}
		defer conn.Close()

		send := smtpSender{conn}.send
		send("220 127.0.0.1 ESMTP service ready")
	})

	t.Run("SMTP server returns a bad handshake", func(t *testing.T) {
		ln := newLocalListener(t)
		defer ln.Close()

		errc := make(chan error)

		go func() {
			defer func() {
				if r := recover(); r == nil {
					errc <- errors.New("expected to panic")
					return
				}
				errc <- nil
			}()
			New(ln.Addr().String(), zerolog.New(ioutil.Discard))
		}()

		conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("failed to accept connection: %v", err)
		}
		defer conn.Close()

		send := smtpSender{conn}.send
		send("502 127.0.0.1 ESMTP service ready")

		if err := <-errc; err != nil {
			t.Error(err)
		}
	})
}

func TestSMTP_Send(t *testing.T) {
	type args struct {
		ctx context.Context
		msg *converter.Message
	}
	tests := []struct {
		name       string
		smtpClient goSMTP
		args       args
		accepted   int
		wantErr    bool
	}{
		{
			name:       "message is nil",
			smtpClient: &fakeSMTP{},
			args: args{
				ctx: context.Background(),
				msg: nil,
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name:       "error when reading raw message",
			smtpClient: &fakeSMTP{},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("", nil, nil, nil, &testutil.FailingReader{}),
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name:       "message has no recipients",
			smtpClient: &fakeSMTP{},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("", nil, nil, nil, strings.NewReader("")),
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name:       "context is done",
			smtpClient: &fakeSMTP{},
			args: args{
				// Expires the given context
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					return ctx
				}(),
				msg: converter.NewMessage("from@example.com", []string{"to@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 0,
			wantErr:  false,
		},
		{
			name: "mail command failed",
			smtpClient: &fakeSMTP{
				mail: strCmdKO,
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name: "rcpt command failed",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdKO,
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name: "data command failed",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: func() (io.WriteCloser, error) {
					return nil, errors.New("internal error")
				},
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name: "data write command failed",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: func() (io.WriteCloser, error) {
					return &testutil.StubWriteCloser{WriteErr: errors.New("failed to write")}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 0,
			wantErr:  true,
		},
		{
			name: "transaction succeed",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: dataOK,
			},
			args: args{
				ctx: func() context.Context { return ctx.WithTraceID(context.Background(), "trace_id") }(),
				msg: converter.NewMessage("from@example.com", []string{"to@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 1,
			wantErr:  false,
		},
		{
			name: "transaction succeed with multiple tos",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: dataOK,
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to1@example.com", "John Doe <to2@example.com>", "to3@example.com"}, nil, nil, strings.NewReader("")),
			},
			accepted: 3,
			wantErr:  false,
		},
		{
			name: "transaction succeed with to and cc",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: dataOK,
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to1@example.com", "to2@example.com"}, []string{"John Doe <cc@example.com>"}, nil, strings.NewReader("")),
			},
			accepted: 3,
			wantErr:  false,
		},
		{
			name: "transactions succeed with to, cc, and bcc",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: dataOK,
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", []string{"to1@example.com", "to2@example.com"}, []string{"John Doe <cc@example.com>"}, []string{"bcc@example.com"}, strings.NewReader("")),
			},
			accepted: 4,
			wantErr:  false,
		},
		{
			name: "transactions succeed with bcc only",
			smtpClient: &fakeSMTP{
				mail: strCmdOK,
				rcpt: strCmdOK,
				data: dataOK,
			},
			args: args{
				ctx: context.Background(),
				msg: converter.NewMessage("from@example.com", nil, nil, []string{"bcc1@example.com", "bcc2@example.com", "bcc3@example.com"}, strings.NewReader("")),
			},
			accepted: 3,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &smtpClient{
				client: tt.smtpClient,
				logger: zerolog.Nop(),
			}
			got, err := s.Send(tt.args.ctx, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SMTP.Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.accepted {
				t.Errorf("SMTP.Send() = %v, want %v", got, tt.accepted)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("SMTP close ok", func(t *testing.T) {
		s := &smtpClient{
			client: &fakeSMTP{
				close: func() error {
					return nil
				},
			},
		}
		if err := s.Close(); err != nil {
			t.Errorf("SMTP.Close() = %v, want nil", err)
		}
	})

	t.Run("SMTP close error", func(t *testing.T) {
		wantErr := errors.New("closing error")
		s := &smtpClient{
			client: &fakeSMTP{
				close: func() error {
					return wantErr
				},
			},
		}
		if err := s.Close(); err == nil {
			t.Errorf("SMTP.Close() = nil, want %v", wantErr)
		}
	})
}

func Test_buildRcptLists(t *testing.T) {
	tests := []struct {
		name string
		msg  *converter.Message
		want [][]string
	}{
		{
			name: "no recipient",
			msg:  nil,
			want: nil,
		},
		{
			name: "single to recipient",
			msg:  converter.NewMessage("", []string{"to@example.com"}, nil, nil, nil),
			want: [][]string{{"to@example.com"}},
		},
		{
			name: "multiple to recipients",
			msg:  converter.NewMessage("", []string{"to@example.com", "to2@example.com"}, nil, nil, nil),
			want: [][]string{{"to@example.com", "to2@example.com"}},
		},
		{
			name: "single cc recipient",
			msg:  converter.NewMessage("", nil, []string{"cc@example.com"}, nil, nil),
			want: [][]string{{"cc@example.com"}},
		},
		{
			name: "multiple cc recipients",
			msg:  converter.NewMessage("", nil, []string{"cc@example.com", "cc2@example.com"}, nil, nil),
			want: [][]string{{"cc@example.com", "cc2@example.com"}},
		},
		{
			name: "to and cc recipients",
			msg:  converter.NewMessage("", []string{"to@example.com"}, []string{"cc@example.com"}, nil, nil),
			want: [][]string{{"to@example.com", "cc@example.com"}},
		},
		{
			name: "to, cc, and bcc recipients",
			msg:  converter.NewMessage("", []string{"to@example.com"}, []string{"cc@example.com"}, []string{"bcc@example.com"}, nil),
			want: [][]string{{"to@example.com", "cc@example.com"}, {"bcc@example.com"}},
		},
		{
			name: "to, cc, and multiple bcc recipients",
			msg:  converter.NewMessage("", []string{"to@example.com"}, []string{"cc@example.com"}, []string{"bcc@example.com", "bcc2@example.com"}, nil),
			want: [][]string{{"to@example.com", "cc@example.com"}, {"bcc@example.com"}, {"bcc2@example.com"}},
		},
		{
			name: "bcc only recipients",
			msg:  converter.NewMessage("", nil, nil, []string{"bcc@example.com", "bcc2@example.com"}, nil),
			want: [][]string{{"bcc@example.com"}, {"bcc2@example.com"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRcptLists(tt.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildRcptLists() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

// Copied from smtp/smtp_test.go
func newLocalListener(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

// Copied from smtp/smtp_test.go
type smtpSender struct {
	w io.Writer
}

// Copied from smtp/smtp_test.go
func (s smtpSender) send(f string) {
	(s.w.Write([]byte(f + "\r\n")))
}

var (
	strCmdOK = func(s string) error { return nil }
	strCmdKO = func(s string) error { return errors.New("cmd failed") }
	dataOK   = func() (io.WriteCloser, error) { return &testutil.StubWriteCloser{}, nil }
)

type fakeSMTP struct {
	mail  func(string) error
	rcpt  func(string) error
	data  func() (io.WriteCloser, error)
	close func() error
}

func (f *fakeSMTP) Mail(s string) error {
	if f.mail == nil {
		panic("not implemented")
	}
	return f.mail(s)
}

func (f *fakeSMTP) Rcpt(s string) error {
	if f.rcpt == nil {
		panic("not implemented")
	}
	return f.rcpt(s)
}

func (f *fakeSMTP) Data() (io.WriteCloser, error) {
	if f.data == nil {
		panic("not implemented")
	}
	return f.data()
}

func (f *fakeSMTP) Close() error {
	if f.close == nil {
		panic("not implemented")
	}
	return f.close()
}
