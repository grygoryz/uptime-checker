package logger

import (
	"bytes"
	"context"
	"fmt"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var skipHeaders = []string{"Set-Cookie", "Cookie"}

type Bodies struct {
	ResBody []byte
	ReqBody []byte
}

func Logger() func(next http.Handler) http.Handler {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}

	var f chiMiddleware.LogFormatter = &requestLogger{log.With().Str("hostname", hostname).Logger()}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := f.NewLogEntry(r)

			ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			resBuf := bytes.NewBuffer(make([]byte, 0))
			ww.Tee(resBuf)

			reqBuf := bytes.NewBuffer(make([]byte, 0))
			r.Body = ioutil.NopCloser(io.TeeReader(r.Body, reqBuf))

			t := time.Now()
			defer func() {
				extra := Bodies{
					ResBody: nil,
					ReqBody: nil,
				}
				if ww.Status() >= http.StatusBadRequest {
					extra.ResBody = resBuf.Bytes()
					extra.ReqBody = reqBuf.Bytes()
				}
				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t), extra)
			}()

			next.ServeHTTP(ww, chiMiddleware.WithLogEntry(r, entry))
		}
		return http.HandlerFunc(fn)
	}
}

type requestLogger struct {
	Logger zerolog.Logger
}

func (l *requestLogger) NewLogEntry(r *http.Request) chiMiddleware.LogEntry {
	entry := &RequestLoggerEntry{}
	entry.Logger = getRequestChildLogger(r, l.Logger)
	entry.Logger.Info().Msg("request started")

	return entry
}

type RequestLoggerEntry struct {
	Logger zerolog.Logger
}

func (l *RequestLoggerEntry) Write(status int, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	dict := zerolog.Dict().
		Int("status", status).
		Int("bytes", bytes).
		Int64("elapsed", elapsed.Milliseconds())

	if len(header) > 0 {
		dict.Dict("headers", getHeaderLogDict(header))
	}

	lg := l.Logger.WithLevel(statusLevel(status))

	bodies, _ := extra.(Bodies)
	if status >= http.StatusBadRequest {
		dict.Str("body", string(bodies.ResBody))
		lg.Str("reqBody", string(bodies.ReqBody))
	}

	lg.Dict("response", dict).Msg("request completed")
}

func (l *RequestLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger.Error().
		Str("stack", string(stack)).
		Str("panic", fmt.Sprintf("%+v", v)).
		Msg("panic")
}

func getRequestChildLogger(r *http.Request, l zerolog.Logger) zerolog.Logger {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	requestURL := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	dict := zerolog.Dict().Str("url", requestURL).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remoteIp", r.RemoteAddr).
		Str("proto", r.Proto)

	if reqId := chiMiddleware.GetReqID(r.Context()); reqId != "" {
		dict.Str("reqId", reqId)
	}

	if len(r.Header) > 0 {
		dict.Dict("headers", getHeaderLogDict(r.Header))
	}

	return l.With().Dict("request", dict).Logger()
}

func getHeaderLogDict(header http.Header) *zerolog.Event {
	dict := zerolog.Dict()
outer:
	for k, v := range header {
		for _, h := range skipHeaders {
			if k == h {
				dict.Str(k, "[redacted]")
				continue outer
			}
		}

		switch {
		case len(v) == 0:
			continue
		case len(v) == 1:
			dict.Str(k, v[0])
		default:
			dict.Strs(k, v)
		}
	}

	return dict
}

func statusLevel(status int) zerolog.Level {
	switch {
	case status < http.StatusBadRequest:
		return zerolog.InfoLevel
	case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
		return zerolog.WarnLevel
	case status >= http.StatusInternalServerError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func LogEntry(ctx context.Context) zerolog.Logger {
	entry, ok := ctx.Value(chiMiddleware.LogEntryCtxKey).(*RequestLoggerEntry)
	if !ok || entry == nil {
		return zerolog.Nop()
	}
	return entry.Logger
}
