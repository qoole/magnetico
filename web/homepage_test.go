package web

import (
	"bytes"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomepage(t *testing.T) {
	t.Parallel()

	inputs := []uint{0, math.MaxUint, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	inputs = append(inputs, uint(rand.Int64N(math.MaxInt64)))

	var buffer bytes.Buffer
	for _, tc := range inputs {
		t.Run(fmt.Sprintf("TestHomepage%d", tc), func(t *testing.T) {
			if err := homepage(tc).Render(&buffer); err != nil {
				t.Errorf("homepage render: %v", err)
			}
		})
	}
}

func TestRootHandler(t *testing.T) {
	t.Parallel()

	initDb()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rootHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "0 torrents available") {
		t.Error("handler returned unexpected body: did not contain 0 torrents available")
	}
}