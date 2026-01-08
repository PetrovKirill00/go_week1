package wordstat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHTTPWordstat_Text(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=count&format=text", strings.NewReader("b a a b c"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rr.Code, rr.Body.String())
	}

	want := "a 2\nb 2\nc 1\n"
	if rr.Body.String() != want {
		t.Fatalf("got=%q want=%q", rr.Body.String(), want)
	}
}

func TestHTTPWordstat_JSON(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodPost, "/wordstat?sort=count&format=json", strings.NewReader("b a a b c"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rr.Code, rr.Body.String())
	}

	var got []Entry
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal error=%v body=%q", err, rr.Body.String())
	}

	want := []Entry{
		{Word: "a", Count: 2},
		{Word: "b", Count: 2},
		{Word: "c", Count: 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want%v", got, want)
	}
}

func TestHTTPWordstat_BadMethod(t *testing.T) {
	h := NewHTTPMux()

	req := httptest.NewRequest(http.MethodGet, "/wordstat", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d body=%q", rr.Code, rr.Body.String())
	}
}
