package wordstat

import (
	"context"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const defaultMaxBodyBytes = 1 << 20

type apiError struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func writeError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	reqID, _ := GetRequestID(r.Context())

	w.Header().Set("content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiError{
		Error:     msg,
		RequestID: reqID,
	})
}

type HTTPConfig struct {
	MaxBodyBytes int64
}

var DefaultHTTPConfig = HTTPConfig{
	MaxBodyBytes: defaultMaxBodyBytes,
}

func NewHTTPMux() http.Handler {
	return NewHTTPMuxWithConfig(DefaultHTTPConfig)
}

func NewHTTPMuxWithConfig(cfg HTTPConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("/wordstat", func(w http.ResponseWriter, r *http.Request) {
		// panic("boom")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeError(w, r, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		opts, err := optionsFromQuery(r)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		// ограничение размер входа 1 MB
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)

		if opts.Format == "json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		err = RunCtx(r.Context(), r.Body, w, opts)
		if err == nil {
			return
		}
		// Если контекст уже отменён/истёк, то классифицируем
		if cerr := r.Context().Err(); cerr != nil {
			if errors.Is(cerr, context.Canceled) {
				// Клиент ушел или сервер закрыл соединение - писать ответ бессмысленно.
				reqID, _ := GetRequestID(r.Context())
				log.Printf("request canceled: req_id=%s err=%v", reqID, cerr)
				return
			}
			if errors.Is(cerr, context.DeadlineExceeded) {
				writeError(w, r, http.StatusRequestTimeout, "request timeout")
				return
			}
		}
		// Превышен MaxBytesReader -> 413
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			writeError(w, r, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		// Остальное пока считаем ошибкой запроса
		writeError(w, r, http.StatusBadRequest, err.Error())
	})

	mux.Handle("/debug/vars", expvar.Handler())

	return RequestID(Logging(Recovery(mux)))
}

func optionsFromQuery(r *http.Request) (Options, error) {
	q := r.URL.Query()

	opts := Options{
		SortBy: q.Get("sort"),
		Format: q.Get("format"),
		Min:    1,
		K:      0,
	}

	if opts.SortBy == "" {
		opts.SortBy = "word"
	}
	if opts.Format == "" {
		opts.Format = "text"
	}

	switch opts.SortBy {
	case "word", "count":
	default:
		return Options{}, fmt.Errorf("bad sort=%q", opts.SortBy)
	}
	switch opts.Format {
	case "text", "json":
	default:
		return Options{}, fmt.Errorf("bad format=%q", opts.Format)
	}
	if v := q.Get("min"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Options{}, fmt.Errorf("bad min=%q", v)
		}
		if n <= 0 {
			return Options{}, fmt.Errorf("bad min=%q (must be > 0", v)
		}
		opts.Min = n
	}

	if v := q.Get("k"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Options{}, fmt.Errorf("bad k=%q", v)
		}
		if n < 0 {
			return Options{}, fmt.Errorf("bad k=%q (must be >= 0)", v)
		}
		opts.K = n
	}
	return opts, nil
}
