package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/smtp"

	"github.com/eexit/http2smtp/internal/converter"
	ictx "github.com/eexit/http2smtp/internal/ctx"
	"github.com/rs/zerolog"
)

// goSMTP exposes the Go SMTP client methods used by this package so
// it could be easier tested
type goSMTP interface {
	Mail(string) error
	Rcpt(string) error
	Data() (io.WriteCloser, error)
	Close() error
}

// Client exposes the methods of the SMTP client
type Client interface {
	Send(ctx context.Context, msg *converter.Message) (int, error)
	Close() error
}

// smtpClient wraps smtpClient email sending
type smtpClient struct {
	addr   string
	client goSMTP
	logger zerolog.Logger
}

// New creates a new Go native SMTP client
func New(addr string, logger zerolog.Logger) Client {
	logger = logger.With().Dict(
		"smtp", zerolog.Dict().
			Fields(map[string]interface{}{
				"id":   "go:net/smtp",
				"addr": addr,
			})).Logger()

	logger.Info().Msg("dialing to smtp server")

	client, err := smtp.Dial(addr)
	if err != nil {
		logger.Panic().Err(err).Msg("could not dial to smtp server")
	}

	return &smtpClient{
		addr:   addr,
		client: client,
		logger: logger,
	}
}

// Send sends given messsage and returns the number accepted recipients by the server.
// One transaction is executed for the combination of To+Cc while it will create
// one extra transaction for reach Bcc recipient.
func (s *smtpClient) Send(ctx context.Context, msg *converter.Message) (int, error) {
	if msg == nil {
		return 0, errors.New("failed to process nil message")
	}

	logger := s.logger

	// Contextualize the logger by adding the context trace ID
	if traceID := ictx.TraceID(ctx); traceID != "" {
		logger = logger.With().Str("trace_id", traceID).Logger()
	}

	raw, err := msg.Raw()
	if err != nil {
		logger.Error().Err(err).Msg("failed to read email data")
		return 0, err
	}

	if !msg.HasRecipients() {
		return 0, errors.New("message has no recipient")
	}

	logger.Info().Msg("sending message")

	accepted := 0
	// Loops over all recipients lists and execute one email transaction per list
	for _, tos := range buildRcptLists(msg) {
		select {
		case <-ctx.Done():
			logger.Warn().Msgf("process aborted: %s", ctx.Err())
			return accepted, nil
		default:
			logger.Debug().Strs("tos", tos).Msg("executing transaction")
			if err := s.execTransaction(logger, msg.From(), tos, raw); err != nil {
				return accepted, fmt.Errorf("an error occurred while sending emails: %w", err)
			}
			accepted += len(tos)
			logger.Debug().Strs("tos", tos).Msg("transaction executed")
		}
	}

	logger.Info().Int("accepted", accepted).Msg("message sent")

	return accepted, nil
}

// Close terminates the SMTP connection
func (s *smtpClient) Close() error {
	s.logger.Info().Msg("closing smtp server connection")
	return s.client.Close()
}

func (s *smtpClient) execTransaction(logger zerolog.Logger, from string, tos []string, raw []byte) error {
	logger.Debug().Str("from", from).Msg("sending MAIL FROM cmd")
	if err := s.client.Mail(from); err != nil {
		logger.Error().Err(err).Msg("failed to issue MAIL FROM cmd")
		return err
	}

	for _, to := range tos {
		logger.Debug().Str("to", to).Msg("sending RCPT cmd")
		if err := s.client.Rcpt(to); err != nil {
			logger.Error().Err(err).Msg("failed to issue RCPT cmd")
			return err
		}
	}

	logger.Debug().Msg("sending DATA cmd")
	w, err := s.client.Data()
	if err != nil {
		logger.Error().Err(err).Msg("failed to issue DATA cmd")
		return err
	}
	defer w.Close()

	logger.Debug().Bytes("data", raw).Msg("writing data")
	if _, err := w.Write(raw); err != nil {
		logger.Error().Err(err).Msg("failed to write DATA")
		return err
	}
	return nil
}

// buildRcptLists builds a list of recipients. Each list will translate into
// a single transaction. This function merges To and Cc into a single list while
// it creates a single list per Bcc recipient as the Section 7.2 if RFC 5321 suggests:
// "[...] sending SMTP systems that are aware of "bcc" use MAY find it helpful
// to send each blind copy as a separate message transaction containing only
// a single RCPT command."
func buildRcptLists(msg *converter.Message) [][]string {
	var (
		rcpts [][]string
		tos   []string
	)

	if msg == nil {
		return rcpts
	}

	for _, provider := range []converter.RecipientProvider{(*converter.Message).To, (*converter.Message).Cc} {
		if to := provider(msg); len(to) > 0 {
			tos = append(tos, to...)
		}
	}

	if len(tos) > 0 {
		rcpts = append(rcpts, tos)
	}

	// This loop convert each Bcc recipient into a one-entry string array with the
	// the recipient
	for _, bcc := range msg.Bcc() {
		if len(bcc) > 0 {
			rcpts = append(rcpts, []string{bcc})
		}
	}

	return rcpts
}
