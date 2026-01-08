package wordstat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecovery_JSONIncludesRequestID(t *testing.T) {
	panicH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// RequestID снаружи, чтобы Recovery мог взять id из контекста
	h := RequestID(Recovery(panicH))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/wordstat", strings.NewReader("a"))
	req.Header.Set("X-Request-Id", "test-xyz")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status=%d, want %d, body=%q", res.StatusCode, http.StatusInternalServerError, rr.Body.String())
	}

	if got := res.Header.Get("X-Request-Id"); got != "test-xyz" {
		t.Fatalf("X-Request-Id=%q, want=%q", got, "test-xyz")
	}

	ct := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type=%q, want application/json", ct)
	}

	var payload struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if payload.Error == "" {
		t.Fatalf("error field is empty: %q", rr.Body.String())
	}
	if payload.RequestID != "test-xyz" {
		t.Fatalf("request_id=%q, want %q", payload.RequestID, "test-xyz")
	}
}

func TestRecovery_CantChangeStatusAfterWriteHeader(t *testing.T) {
	panicAfterWrite := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		panic("boom")
	})

	h := RequestID(Recovery(panicAfterWrite))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/wordstat", strings.NewReader("a"))
	req.Header.Set("X-Request-Id", "id-1")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want %d", res.StatusCode, http.StatusOK)
	}
}
