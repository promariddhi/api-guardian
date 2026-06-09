package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"api_guardian/internal/metrics"

	"github.com/google/uuid"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

var traceIdKey contextKey = "trace_id"

func colorForStatus(status int) string {
	switch {
	case status >= 500:
		return red
	case status >= 400:
		return yellow
	case status >= 300:
		return blue
	default:
		return green
	}
}

type StatusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *StatusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StatusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &StatusWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		metrics.ActiveRequests.Inc()
		defer metrics.ActiveRequests.Dec()

		start := time.Now()
		next.ServeHTTP(sw, r)

		metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.String(), strconv.Itoa(sw.status)).Inc()

		duration := time.Since(start)
		metrics.RequestDuration.WithLabelValues(r.Method, r.URL.String()).Observe(duration.Seconds())

		metrics.ResponseBytes.WithLabelValues(r.URL.Path).Add(float64(sw.bytes))

		traceId := r.Context().Value(traceIdKey).(string)
		log.Printf("trace=%s %s %s %d %v", traceId, r.Method, r.URL.Path, sw.status, duration)
	})
}

func Tracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := r.Header.Get("X-Request-ID")
		if traceId == "" {
			traceId = uuid.NewString()
		}

		w.Header().Set("X-Request-ID", traceId)

		ctx := context.WithValue(r.Context(), traceIdKey, traceId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
