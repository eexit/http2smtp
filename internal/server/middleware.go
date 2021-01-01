package server

import (
	"net/http"
	"time"

	"github.com/eexit/httpsmtp/internal/ctx"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) wrap(fn http.Handler) http.Handler {
	return alice.New().
		Append(
			hlog.NewHandler(s.logger),
			traceIDHeaderHandler(s.env.HTTPTraceHeader),
			hlog.MethodHandler("verb"),
			hlog.RemoteAddrHandler("ip"),
			hlog.UserAgentHandler("user_agent"),
			hlog.URLHandler("url"),
			responseHeaderHandler("content-type", "application/json"),
			hlog.AccessHandler(func(r *http.Request, code, size int, duration time.Duration) {
				var level zerolog.Level
				switch {
				case code < http.StatusMultipleChoices:
					level = zerolog.InfoLevel
				case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
					level = zerolog.WarnLevel
				case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
					level = zerolog.ErrorLevel
				case code >= http.StatusInternalServerError:
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

func responseHeaderHandler(key, value string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" && value != "" {
				w.Header().Set(key, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Adapted copy of hlog.CustomHeaderHandler
func traceIDHeaderHandler(header string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if val := r.Header.Get(header); val != "" {
				log := zerolog.Ctx(r.Context())
				log.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("trace_id", val)
				})
				r = r.WithContext(ctx.WithTraceID(r.Context(), val))
			}
			next.ServeHTTP(w, r)
		})
	}
}
