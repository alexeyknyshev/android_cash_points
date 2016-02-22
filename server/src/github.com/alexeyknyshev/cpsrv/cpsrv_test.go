package main

import (
	"bytes"
	"errors"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"github.com/tarantool/go-tarantool"
	"testing"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type TestRequest struct {
	RequestType string
	EndpointUrl string
	HandlerUrl string
	Data string
}

type TestResponse struct {
	Code int
	Data []byte
}

func readResponse(w *httptest.ResponseRecorder) (TestResponse, error) {
	response := TestResponse{}

	response.Code = w.Code

	data, err := ioutil.ReadAll(w.Body)
	if err != nil {
		err = errors.New("Cannot read response body: " + err.Error())
	} else {
		response.Data = data
	}
	return response, err
}

func diff(expected, received []byte) (string, error) {
	differ := gojsondiff.New()
	d, err := differ.Compare(expected, received)
	if err != nil {
		return "", errors.New("Failed to compare json pair: " + err.Error())
	}

	if !d.Modified() {
		return "", nil
	}

	var expectedJson map[string]interface{}
	json.Unmarshal(expected, &expectedJson)
	formatter := formatter.NewAsciiFormatter(expectedJson)
	formatter.ShowArrayIndex = true
	diffString, err := formatter.Format(d)
	if err != nil {
		// No error can occur
	}

	return diffString, nil
}

func tarantoolConnect() (*tarantool.Connection, error) {
	tntUrl := "localhost:3301"
	tntOpts := tarantool.Opts{
		User: "admin",
		Pass: "admin",
	}

	return tarantool.Connect(tntUrl, tntOpts)
}

func testRequest(request TestRequest, handler EndpointCallback) *httptest.ResponseRecorder {
	var req *http.Request = nil

        if request.Data != "" {
		req, _ = http.NewRequest(request.RequestType, request.EndpointUrl, bytes.NewBufferString(request.Data))
	} else {
		req, _ = http.NewRequest(request.RequestType, request.EndpointUrl, nil)
	}

	req.Header.Add("Id", "1")

	w := httptest.NewRecorder()
	m := mux.NewRouter()
	if request.HandlerUrl == "" {
		request.HandlerUrl = request.EndpointUrl
	}
	m.HandleFunc(request.HandlerUrl, handler)
	m.ServeHTTP(w, req)

	return w
}

// ======================================================================

func TestPing(t *testing.T) {
        tntUrl := "localhost:3301"
	tntOpts := tarantool.Opts{
		User: "admin",
		Pass: "admin",
	}

	tnt, err := tarantool.Connect(tntUrl, tntOpts)
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerPing(tnt)
	request := TestRequest{ RequestType: "GET", EndpointUrl: url }
	response, err := readResponse(testRequest(request, handler))

	if response.Code != http.StatusOK {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	expected := Message{ Text: "pong" }
	expectedJson, _ := json.Marshal(expected)

	diffStr, err := diff(expectedJson, response.Data)
	if err != nil {
		t.Errorf("Failed to compare json pair: %v", err)
	}
	if diffStr != "" {
		t.Errorf("\n%s", diffStr)
	}
}

// ======================================================================

type Town struct {
	Id             uint32  `json:"id"`
	Name           string  `json:"name"`
	NameTr         string  `json:"name_tr"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	RegionId       uint32  `json:"region_id"`
	RegionalCenter bool    `json:"regional_center"`
	Big            bool    `json:"big"`
	Zoom           uint32  `json:"zoom"`
}

func TestTown(t *testing.T) {
	tnt, err := tarantoolConnect()
	if err != nil {
		t.Errorf("Connection to tarantool failed: %v", err)
	}
	defer tnt.Close()

	url, handler := handlerTown(tnt)
	request := TestRequest{ RequestType: "GET", EndpointUrl: "/town/4", HandlerUrl: url }
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}
	if response.Code != http.StatusOK {
		t.Errorf("Expected 200 OK but got %d", response.Code)
	}

	expected := Town{
		Id: 4,
		Name: "Москва",
		NameTr: "Moskva",
		Longitude: 37.61775970459,
		Latitude: 55.755771636963,
		RegionId: 3,
		RegionalCenter: true,
		Big: true,
		Zoom: 10,
	}
	expectedJson, _ := json.Marshal(expected)

	diffStr, err := diff(expectedJson, response.Data)
	if err != nil {
		t.Errorf("Failed to compare json pair: %v", err)
	}
	if diffStr != "" {
		t.Errorf("\n%s", diffStr)
	}
}
