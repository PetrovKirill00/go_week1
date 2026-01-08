package wordstat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"expvar"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

type ctxKey int

const ctxKeyReqID ctxKey = iota

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}

		// ответ всегда содержит request id
		w.Header().Set("X-Request-Id", id)

		// пробрасываем вниз по цепочке
		ctx := context.WithValue(r.Context(), ctxKeyReqID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	// fallback
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}

func GetRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyReqID).(string)
	return id, ok
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(p []byte) (int, error) {
	// если handler не вызвал WriteHeader - по умолчанию это 200
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

// ---- metrics ----

var (
	httpRequestsTotal   = expvar.NewInt("http_requests_total")
	httpErrorsTotal     = expvar.NewInt("http.errors_total")
	httpStatusTotal     = expvar.NewMap("http_status_total")
	httpDurationUsTotal = expvar.NewInt("http_duration_us_total")
	httpRespBytesTotal  = expvar.NewInt("http_response_bytes_total")
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(sw, r)

		lat := time.Since(start)

		// Бывает, что ничего не записали (например, паника/обрыв до ответа)
		status := sw.status
		if status == 0 {
			status = http.StatusOK
		}

		// metrics
		httpRequestsTotal.Add(1)
		httpStatusTotal.Add(strconv.Itoa(status), 1)
		httpDurationUsTotal.Add(lat.Microseconds())
		httpRespBytesTotal.Add(int64(sw.bytes))
		if status >= 500 {
			httpErrorsTotal.Add(1)
		}

		reqID, _ := GetRequestID(r.Context())

		log.Printf("req_id=%s %s %s status=%d bytes=%d dur=%s",
			reqID, r.Method, r.URL.RequestURI(), status, sw.bytes, lat)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID, _ := GetRequestID(r.Context())
				log.Printf("panic: req_id=%s rec=%v\n%s", reqID, rec, debug.Stack())

				// Пытаемся дать нормальный ответ клиенту
				writeError(w, r, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
