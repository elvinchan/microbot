package microbot

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics?target=pprof&name=profile&seconds=5", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	ProfController().ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)
}

func TestMetricsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	MetricsController().ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)
}

func TestIPBlock(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	req.Header.Add("x-real-ip", "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	MetricsController().ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusForbidden)
}
