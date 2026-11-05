package web

import (
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("Terjadi panic pada handler", "err", rec, "stack", string(debug.Stack()))
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Reswap", "none")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`<div id="toast" hx-swap-oob="innerHTML">
						<div class="rounded-2xl p-5 mb-4 shadow-lg border-2 bg-rose-50 border-rose-500 text-rose-950 font-bold">
							Ada gangguan sistem. Coba lagi atau hubungi IT.
						</div>
					</div>`))
				} else {
					http.Error(w, "Ada gangguan sistem. Coba lagi atau hubungi IT.", http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		isHX := r.Header.Get("HX-Request") == "true"
		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"hx", isHX,
		)
	})
}

func CrossOriginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Check Sec-Fetch-Site if available
			fetchSite := r.Header.Get("Sec-Fetch-Site")
			if fetchSite != "" && fetchSite != "same-origin" && fetchSite != "none" {
				slog.Warn("Menolak POST lintas-origin", "Sec-Fetch-Site", fetchSite, "origin", r.Header.Get("Origin"))
				http.Error(w, "Permintaan lintas origin ditolak.", http.StatusForbidden)
				return
			}

			// Check Origin (bandingkan host secara utuh, bukan substring)
			origin := r.Header.Get("Origin")
			if origin != "" {
				parsed, err := url.Parse(origin)
				if err != nil || parsed.Host == "" || !strings.EqualFold(parsed.Host, r.Host) {
					slog.Warn("Origin tidak cocok dengan Host", "origin", origin, "host", r.Host)
					http.Error(w, "Origin tidak diizinkan.", http.StatusForbidden)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		next.ServeHTTP(w, r)
	})
}
