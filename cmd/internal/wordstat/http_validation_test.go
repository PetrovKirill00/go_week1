package wordstat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type jsonErr struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id"`
}

func requireReqIDHeader(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := rr.Result().Header.Get("X-Request-Id"); got != want {
		t.Fatalf("X-Request-Id=%q want=%q body=%q", got, want, rr.Body.String())
	}
}

func requireJSONError(t *testing.T, rr *httptest.ResponseRecorder, wantReqID string) jsonErr {
	t.Helper()

	ct := rr.Result().Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type=%q want application/json body=%q", ct, rr.Body.String())
	}

	var e jsonErr
	if err := json.Unmarshal(rr.Body.Bytes(), &e); err != nil {
		t.Fatalf("json unmarshal: %v body=%q", err, rr.Body.String())
	}

	if e.Error == "" {
		t.Fatalf("empty error field: body=%q", rr.Body.String())
	}
	if wantReqID != "" && e.RequestID != wantReqID {
		t.Fatalf("request_id=%q want=%q body=%q", e.RequestID, wantReqID, rr.Body.String())
	}
	return e
}

func TestHTTPValidation_BadSort_JSON(t *testing.T) {
	h := NewHTTPMux()
	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=oops&format=json", strings.NewReader("a a"))
	req.Header.Set("X-Request-Id", "rid-badsort")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	requireReqIDHeader(t, rr, "rid-badsort")
	_ = requireJSONError(t, rr, "rid-badsort")
}

func TestHTTPValidation_BadFormat_JSON(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=count&format=xml", strings.NewReader("a a"))
	req.Header.Set("X-Request-Id", "rid-badfmt")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	requireReqIDHeader(t, rr, "rid-badfmt")
	_ = requireJSONError(t, rr, "rid-badfmt")
}

func TestHTTPValidation_BadMin(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=count&format=json&min=abc", strings.NewReader("a a"))
	req.Header.Set("X-Request-Id", "rid-badmin")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	requireReqIDHeader(t, rr, "rid-badmin")
	_ = requireJSONError(t, rr, "rid-badmin")
}

func TestHTTPValidation_BadK(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=count&format=json&k=zz", strings.NewReader("a a"))
	req.Header.Set("X-Request-Id", "rid-badk")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	requireReqIDHeader(t, rr, "rid-badk")
	_ = requireJSONError(t, rr, "rid-badk")
}

func TestHTTPValidation_TooLargeBody(t *testing.T) {
	h := NewHTTPMux()

	// > 1 MB
	body := bytes.Repeat([]byte("a "), 700_000) // ~1.4 MB
	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=count&format=json", bytes.NewReader(body))
	req.Header.Set("X-Request-Id", "rid-toolarge")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusRequestEntityTooLarge, rr.Body.String())
	}
	requireReqIDHeader(t, rr, "rid-toolarge")
	_ = requireJSONError(t, rr, "rid-toolarge")
}

func TestHTTPValidation_MethodNotAllowed_AllowHeader(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodGet, "/wordstat?format=json", nil)
	req.Header.Set("X-Request-Id", "rid-405")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusMethodNotAllowed, rr.Body.String())
	}
	allow := rr.Result().Header.Get("Allow")
	if !strings.Contains(allow, http.MethodPost) {
		t.Fatalf("Allow=%q want to contain %q", allow, http.MethodPost)
	}

	requireReqIDHeader(t, rr, "rid-405")
	_ = requireJSONError(t, rr, "rid-405")
}

func TestHTTPValidation_Valid_Request_JSON(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(
		http.MethodPost,
		"/wordstat?sort=count&format=json",
		strings.NewReader("b a a b c"),
	)
	req.Header.Set("X-Request-Id", "rid-ok")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}

	requireReqIDHeader(t, rr, "rid-ok")

	ct := rr.Result().Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type=%q want application/json body=%q", ct, rr.Body.String())
	}

	got := extractCountsFromJSON(t, rr.Body.Bytes())

	want := map[string]int{"a": 2, "b": 2, "c": 1}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d != len(want)=%d", len(got), len(want))
	}

	for k, v := range want {
		if got[k] != v {
			t.Fatalf("counts[%q]=%d want=%d; full=%v; body=%q", k, got[k], v, got, rr.Body.String())
		}
	}
}

func extractCountsFromJSON(t *testing.T, b []byte) map[string]int {
	t.Helper()

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("json unmarshal: %v body=%q", err, string(b))
	}

	out := map[string]int{}

	switch vv := v.(type) {
	case map[string]any:
		// формат {"a":2,"b":2}
		for k, raw := range vv {
			if f, ok := raw.(float64); ok {
				out[k] = int(f)
			}
		}
		return out

	case []any:
		for _, el := range vv {
			m, ok := el.(map[string]any)
			if !ok {
				continue
			}
			w, _ := m["word"].(string)
			if w == "" {
				w, _ = m["Word"].(string)
			}
			var c int
			if f, ok := m["count"].(float64); ok {
				c = int(f)
			} else if f, ok := m["Count"].(float64); ok {
				c = int(f)
			}
			if w != "" {
				out[w] = c
			}
		}
		return out
	}

	t.Fatalf("unexpected json shape: %T body=%q", v, string(b))
	return nil
}
