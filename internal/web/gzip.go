package web

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// gzipMiddleware compresses responses when the client advertises gzip support.
// The live feed is server-sent events and is passed through untouched, as are
// range requests and HEAD requests.
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead ||
			r.URL.Path == "/api/live" ||
			r.Header.Get("Range") != "" ||
			!acceptsGzip(r) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Accept-Encoding")
		writer := gzip.NewWriter(w)
		gzw := &gzipResponseWriter{ResponseWriter: w, writer: writer}
		defer func() {
			if gzw.compressing {
				_ = writer.Close()
			}
		}()

		next.ServeHTTP(gzw, r)
	})
}

func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		encoding := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if strings.EqualFold(encoding, "gzip") {
			return true
		}
	}
	return false
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	compressing bool
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	okStatus := status >= 200 && status < 300 &&
		status != http.StatusNoContent &&
		status != http.StatusResetContent &&
		status != http.StatusPartialContent
	w.compressing = (okStatus || status >= 400) && w.Header().Get("Content-Encoding") == ""

	if w.compressing {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		if len(p) > 0 && w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(p))
		}
		w.WriteHeader(http.StatusOK)
	}
	if w.compressing {
		return w.writer.Write(p)
	}
	return w.ResponseWriter.Write(p)
}
