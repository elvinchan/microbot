package microbot

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics?t=pprof&name=profile&seconds=5", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	ProfController().ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)
}
