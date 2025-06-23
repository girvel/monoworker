package monoworker

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingRoute(t *testing.T) {
    worker := NewWorker(func(_ string) string { return "" }, Config{})
    router := BuildAPI(worker)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/ping", nil)
    router.ServeHTTP(w, req)

    if w.Code != 200 {
        t.Errorf("/ping errored out with code %d", w.Code)
    }
}

// TODO other endpoints
