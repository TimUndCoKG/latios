package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/timsalokat/latios_proxy/db"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func AnalyticsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		// Remove internal logs
		if strings.HasPrefix(r.URL.Path, "/latios-api") ||
			strings.HasPrefix(r.URL.Path, "/latios") {
			return
		}

		logEntry := db.RequestLog{
			Timestamp:  start,
			Method:     r.Method,
			Host:       r.Host,
			Path:       r.URL.Path,
			StatusCode: rw.statusCode,
			LatencyMs:  time.Since(start).Milliseconds(),
			RemoteAddr: r.RemoteAddr,
		}

		go db.Client.Create(&logEntry)
	})
}

// func LoggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.URL.Path != "/latios-api/health" || r.Host == "localhost" {
// 			log.Printf("[LOG-MW REQUEST] Method=%s Path=%s RemoteAddr=%s Host=%s Headers=%v",
// 				r.Method, r.URL.Path, r.RemoteAddr, r.Host, r.Header)
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
