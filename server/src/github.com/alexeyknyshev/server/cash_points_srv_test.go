package main

import (
	"github.com/mediocregopher/radix.v2/pool"
	"testing"
	"net/http"
	"net/http/httptest"
)

const REDIS_HOST = "localhost:6379"

type RequestBuilderFunc func(r *http.Request)

func sendRequest(requestType string, handler EndpointCallback, builder RequestBuilderFunc) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(requestType, "", nil)
	req.Header.Add("Id", "1")
	builder(req)
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestPing(t *testing.T) {
	p, err := pool.New("tcp", REDIS_HOST, 16)
	if err != nil {
		t.Errorf("cannot connect to Redis")
	}
	handler := handlerPing(p)
	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Add("Id", "1")
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("ping status failed")
	}
}

func TestUserCreate(t *testing.T) {
	/*response := sendRequest("POST", handlerUserCreate(), func(r *http.Request) {})
	if (response.Code != http.StatusOK) {
		t.Errorf("Expected http status: 200")
	}*/
}