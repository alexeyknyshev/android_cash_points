package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexeyknyshev/gojsondiff"
	"github.com/alexeyknyshev/gojsondiff/formatter"
	"github.com/gorilla/mux"
	//"github.com/tarantool/go-tarantool"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type TestRequest struct {
	RequestType string
	EndpointUrl string
	HandlerUrl  string
	Data        string
}

type TestResponse struct {
	Code int
	Data []byte
}

func getServerConfig() *ServerConfig {
	servConf := new(ServerConfig)
	servConf.TntUrl = "localhost:3301"
	servConf.TntUser = "admin"
	servConf.TntPass = "admin"
	return servConf
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

var DEFAULT_COMPARE_CONF gojsondiff.CompareConfig = gojsondiff.CompareConfig{FloatEpsilon: 0.0001}

func diff(expected, received []byte, conf *gojsondiff.CompareConfig) (string, error) {
	//isObject := true
	if conf == nil {
		conf = &DEFAULT_COMPARE_CONF
	}

	differ := gojsondiff.New()

	// try to compare as objects
	d, err := differ.Compare(expected, received, conf)
	if err != nil {
		// try to compare as arrays
		var nextErr error
		d, nextErr = differ.CompareArrays(expected, received, conf)

		// return first error on second failure
		if nextErr != nil {
			return "", errors.New("Failed to compare json pair: " + err.Error())
		}
		//isObject = false
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
	m.HandleFunc(request.HandlerUrl, handler).Methods(request.RequestType)
	m.ServeHTTP(w, req)

	return w
}

func checkHttpCode(t *testing.T, got, expected int) bool {
	if got != expected {
		t.Errorf("Expected %d %s but got %d", expected, http.StatusText(expected), got)
		return false
	}
	return true
}

func checkJsonResponse(t *testing.T, got, expected []byte) bool {
	diffStr, err := diff(expected, got, nil)
	if err != nil {
		t.Errorf("Failed to compare json pair: %v", err)
		return false
	}
	if diffStr != "" {
		t.Errorf("\n%s", diffStr)
		return false
	}
	return true
}

// ======================================================================

func getSpaceMetrics(hCtx HandlerContext) ([]byte, error) {
	url, handler := handlerSpaceMetrics(hCtx)
	request := TestRequest{RequestType: "GET", EndpointUrl: url}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

type SpaceMetricsGetter func() ([]byte, error)

func checkSpaceMetrics(t *testing.T, getMetrics SpaceMetricsGetter, expected []byte) bool {
	got, err := getMetrics()
	if err != nil {
		t.Errorf("Failed to get space metric on defer: %v", err)
		return false
	}

	return checkJsonResponse(t, got, expected)
}

// ======================================================================

func getQuadTreeBranch(t *testing.T, hCtx HandlerContext, longitude, latitude float64) ([]byte, error) {
	// get quadkey for coorditate
	quadKeyReq := QuadKeyRequest{
		Longitude: longitude,
		Latitude:  latitude,
	}
	quadkeyReqJson, _ := json.Marshal(quadKeyReq)
	url, handlerQuadKey := handlerCoordToQuadKey(hCtx)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: url,
		Data:        string(quadkeyReqJson),
	}

	response, err := readResponse(testRequest(request, handlerQuadKey))
	if err != nil {
		return nil, err
	}
	if !checkHttpCode(t, response.Code, http.StatusOK) {
		return nil, fmt.Errorf("cannot get quadkey for coordinate: (%f, %f)",
			longitude, latitude)
	}

	quadKeyResponse := QuadKeyResponse{}
	err = json.Unmarshal(response.Data, &quadKeyResponse)
	if err != nil {
		return nil, fmt.Errorf("cannot unpack quadkey response: %v => %s",
			err, string(response.Data))
	}

	if quadKeyResponse.QuadKey == "" {
		return nil, fmt.Errorf("received empty quadkey => %s", string(response.Data))
	}

	// save quadtree branch state (before adding cashpoint)
	url, handlerTreeBranch := handlerQuadTreeBranch(hCtx)
	requestTreeBranch := TestRequest{
		RequestType: "GET",
		EndpointUrl: "/quadtree/branch/" + quadKeyResponse.QuadKey,
		HandlerUrl:  url,
	}

	response, err = readResponse(testRequest(requestTreeBranch, handlerTreeBranch))
	if err != nil {
		return nil, err
	}
	if !checkHttpCode(t, response.Code, http.StatusOK) {
		return nil, fmt.Errorf("cannot get quad tree branch for quadkey: %s", quadKeyResponse.QuadKey)
	}

	return response.Data, nil
}

type QuadTreeBranchGetter func() ([]byte, error)

func checkQuadTreeBranch(t *testing.T, getBranch QuadTreeBranchGetter, expected []byte) bool {
	got, err := getBranch()
	if err != nil {
		t.Errorf("Failed to get quad tree branch on defer: %v", err)
		return false
	}

	var clusters, clustersNew ClusterArray
	err = json.Unmarshal(expected, &clusters)
	if err != nil {
		t.Errorf("Cannot unpack expected quad tree branch response: %v", err)
	}

	err = json.Unmarshal(got, &clustersNew)
	if err != nil {
		t.Errorf("Cannot unpack defer quad tree branch response: %v", err)
	}

	if same, diffText := clusters.Compare(clustersNew); !same {
		t.Fatalf("%s\n\n%s\n%s\n%s",
			diffText,
			"ALERT! Quad tree branches before and after create + delete are different.",
			"Looks like quad tree is broken after test and data in tarantool is corrupted.",
			"Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script.")
		return false
	}
	return true
}

// ======================================================================

func TestPing(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	url, handler := handlerPing(hCtx)
	request := TestRequest{RequestType: "GET", EndpointUrl: url}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)

	expected := Message{Text: "pong"}
	expectedJson, _ := json.Marshal(expected)

	checkJsonResponse(t, response.Data, expectedJson)
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
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	url, handler := handlerTown(hCtx)
	request := TestRequest{RequestType: "GET", EndpointUrl: "/town/4", HandlerUrl: url}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	expected := Town{
		Id:             4,
		Name:           "Москва",
		NameTr:         "Moskva",
		Longitude:      37.61775970459,
		Latitude:       55.755771636963,
		RegionId:       3,
		RegionalCenter: true,
		Big:            true,
		Zoom:           10,
	}
	expectedJson, _ := json.Marshal(expected)

	checkJsonResponse(t, response.Data, expectedJson)
}

// ======================================================================

type CashpointShort struct {
	Id             uint32  `json:"id,omitempty"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	Type           string  `json:"type"`
	BankId         uint32  `json:"bank_id"`
	TownId         uint32  `json:"town_id"`
	Address        string  `json:"address"`
	AddressComment string  `json:"address_comment"`
	MetroName      string  `json:"metro_name"`
	FreeAccess     bool    `json:"free_access"`
	MainOffice     bool    `json:"main_office"`
	WithoutWeekend bool    `json:"without_weekend"`
	RoundTheClock  bool    `json:"round_the_clock"`
	WorksAsShop    bool    `json:"works_as_shop"`
	Schedule       string  `json:"schedule"`
	Tel            string  `json:"tel"`
	Additional     string  `json:"additional"`
	Rub            bool    `json:"rub"`
	Usd            bool    `json:"usd"`
	Eur            bool    `json:"eur"`
	CashIn         bool    `json:"cash_in"`
}

type CashpointFull struct {
	CashpointShort
	Version uint32 `json:"version"`
	//	Timestamp      uint64  `json:"timestamp"` // TODO: timestamp on server
	Approved bool `json:"approved"`
}

func TestCashpoint(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())

	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	url, handler := handlerCashpoint(hCtx)
	var id uint32 = 7138832
	request := TestRequest{
		RequestType: "GET",
		EndpointUrl: "/cashpoint/" + strconv.FormatUint(uint64(id), 10),
		HandlerUrl:  url,
	}
	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)

	cpShort := CashpointShort{
		Id:             id,
		Longitude:      37.562019348145,
		Latitude:       55.6633644104,
		Type:           "atm",
		BankId:         2764,
		TownId:         4,
		Address:        "г. Москва, ул. Новочеремушкинская, д. 69",
		AddressComment: "ОАО «Вниизарубежгеология»",
		MetroName:      "",
		FreeAccess:     true,
		MainOffice:     false,
		WithoutWeekend: false,
		RoundTheClock:  false,
		WorksAsShop:    true,
		Schedule:       "",
		Tel:            "",
		Additional:     "",
		Rub:            true,
		Usd:            false,
		Eur:            false,
		CashIn:         false,
	}

	cp := CashpointFull{
		CashpointShort: cpShort,
		Version:        0,
		//Timestamp: 0,
		Approved: true,
	}
	expectedJson, _ := json.Marshal(cp)
	checkJsonResponse(t, response.Data, expectedJson)
}

// ======================================================================

type QuadKeyRequest struct {
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Zoom      uint32  `json:"zoom,omitempty"`
}

type QuadKeyResponse struct {
	QuadKey string `json:"quadkey"`
}

func TestQuadKeyFromCoord(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())

	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	url, handler := handlerCoordToQuadKey(hCtx)

	// empty request
	quadKeyReq := QuadKeyRequest{}
	reqJson, _ := json.Marshal(quadKeyReq)

	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: url,
		Data:        string(reqJson),
	}

	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusBadRequest)

	// request with missing Latitude
	quadKeyReq.Longitude = 56.6
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusBadRequest)

	// normal request
	quadKeyReq.Latitude = 34.84
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := QuadKeyResponse{QuadKey: "3032100220113311"}
	expectedJson, _ := json.Marshal(expected)

	checkHttpCode(t, response.Code, http.StatusOK)
	checkJsonResponse(t, response.Data, expectedJson)

	// request with zoom
	quadKeyReq.Zoom = 16
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)
	checkJsonResponse(t, response.Data, expectedJson)

	// request with lower zoom
	quadKeyReq.Zoom = 12
	reqJson, _ = json.Marshal(quadKeyReq)

	request.Data = string(reqJson)

	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	expected.QuadKey = "303210022011"
	expectedJson, _ = json.Marshal(expected)

	checkHttpCode(t, response.Code, http.StatusOK)
	checkJsonResponse(t, response.Data, expectedJson)
}

// ======================================================================

type Cluster struct {
	Id        string   `json:"id"`
	Longitude float64  `json:"longitude"`
	Latitude  float64  `json:"latitude"`
	Members   []uint32 `json:"members"`
	Size      uint32   `json:"size"`
}

type ClusterArray []Cluster

func (c ClusterArray) Compare(other ClusterArray) (bool, string) {
	if len(c) != len(other) {
		return false, "different length of ClusterArrays"
	}

	for i := 0; i < len(c); i++ {
		expectedJson, _ := json.Marshal(c[i])
		responseJson, _ := json.Marshal(other[i])

		diffStr, err := diff(expectedJson, responseJson, nil)
		if err != nil {
			return false, "failed to compare Cluster json pair at index " +
				strconv.FormatInt(int64(i), 10) + ": " + err.Error()
		}
		if diffStr != "" {
			return false, "different Cluster json pair at index " +
				strconv.FormatInt(int64(i), 10) + ": " + diffStr
		}
	}

	// 	return false, "successfully compared " + strconv.FormatInt(int64(len(c)), 10) + " clusters"
	return true, ""
}

func TestQuadTreeBranch(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	url, handler := handlerQuadTreeBranch(hCtx)
	request := TestRequest{
		RequestType: "GET",
		EndpointUrl: "/quadtree/branch/3201323213002023",
		HandlerUrl:  url,
	}

	response, err := readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)

	// test short quadkey
	request.EndpointUrl = "/quadtree/branch/3201323213002"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusOK)

	// test empty quadkey
	request.EndpointUrl = "/quadtree/branch/"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusNotFound)

	// test too long quadkey
	request.EndpointUrl = "/quadtree/branch/320132321300211100"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusBadRequest)

	// test wrong quadkey
	request.EndpointUrl = "/quadtree/branch/3201323253002023"
	response, err = readResponse(testRequest(request, handler))
	if err != nil {
		t.Errorf("%v", err)
	}

	checkHttpCode(t, response.Code, http.StatusNotFound)
}

// ======================================================================

type CashpointCreateRequest struct {
	UserId uint32         `json:"user_id"`
	Data   CashpointShort `json:"data"`
}

func TestCashpointCreateSuccessful(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	longitude := 37.6247
	latitude := 55.7591

	// predict quadkey of the future cashpoint
	quadKeyReq := QuadKeyRequest{
		Longitude: longitude,
		Latitude:  latitude,
	}
	quadkeyReqJson, _ := json.Marshal(quadKeyReq)

	url, handlerQuadKey := handlerCoordToQuadKey(hCtx)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: url,
		Data:        string(quadkeyReqJson),
	}

	response, err := readResponse(testRequest(request, handlerQuadKey))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	quadKeyResponse := QuadKeyResponse{}
	err = json.Unmarshal(response.Data, &quadKeyResponse)
	if err != nil {
		t.Errorf("Cannot unpack quadkey response: %v => %s", err, string(response.Data))
	}

	if quadKeyResponse.QuadKey == "" {
		t.Errorf("Received empty quadkey => %s", string(response.Data))
	}

	// save quadtree branch state (before adding cashpoint)
	url, handlerTreeBranch := handlerQuadTreeBranch(hCtx)
	requestTreeBranch := TestRequest{
		RequestType: "GET",
		EndpointUrl: "/quadtree/branch/" + quadKeyResponse.QuadKey,
		HandlerUrl:  url,
	}

	response, err = readResponse(testRequest(requestTreeBranch, handlerTreeBranch))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	var clusters ClusterArray
	err = json.Unmarshal(response.Data, &clusters)
	if err != nil {
		t.Errorf("Cannot unpack quad tree branch response: %v", err)
	}

	// creating real cashpoint
	cp := CashpointShort{
		Longitude:      longitude,
		Latitude:       latitude,
		Type:           "atm",
		BankId:         322, // Sberbank
		TownId:         4,   // Moscow
		Address:        "",
		AddressComment: "",
		//		MetroName: "",
		FreeAccess:     true,
		MainOffice:     false,
		WithoutWeekend: true,
		RoundTheClock:  false,
		WorksAsShop:    false,
		Schedule:       "",
		Tel:            "",
		Additional:     "",
		Rub:            true,
		Usd:            false,
		Eur:            false,
		CashIn:         true,
	}

	reqData := CashpointCreateRequest{
		UserId: 0, // TODO: check against real user
		Data:   cp,
	}
	reqJson, _ := json.Marshal(reqData)

	url, handlerCreate := handlerCashpointCreate(hCtx)
	request = TestRequest{
		RequestType: "POST",
		EndpointUrl: "/cashpoint",
		HandlerUrl:  url,
		Data:        string(reqJson),
	}
	response, err = readResponse(testRequest(request, handlerCreate))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	var cashpointId uint64 = 0
	err = json.Unmarshal(response.Data, &cashpointId)
	if err != nil {
		t.Errorf("Cannot unpack cashpoint id response: %v => %s", err, string(response.Data))
	}

	cashpointIdStr := strconv.FormatUint(cashpointId, 10)
	t.Logf("created cashpoint with id: %s", cashpointIdStr)

	// TODO: check cashpoint data
	urlGet, handlerGet := handlerCashpoint(hCtx)
	request = TestRequest{
		RequestType: "GET",
		EndpointUrl: "/cashpoint/" + cashpointIdStr,
		HandlerUrl:  urlGet,
	}

	response, err = readResponse(testRequest(request, handlerGet))
	if err != nil {
		t.Errorf("%v", err)
	}
	if !checkHttpCode(t, response.Code, http.StatusOK) {
		t.Fatalf("Cannot get created cashpoint with id: %d", cashpointId)
	}

	// extend cashpoint short data with returned id
	cp.Id = uint32(cashpointId)
	cpFull := CashpointFull{
		CashpointShort: cp,
		Version:        0,
		Approved:       false,
	}
	expectedJson, _ := json.Marshal(cpFull)

	// repack response to remove timestamp field
	var tmpJson map[string]interface{}
	err = json.Unmarshal(response.Data, &tmpJson)
	if err != nil {
		t.Fatalf("Cannot unpack cashpoint data retrieved by id: %d", cashpointId)
	}
	delete(tmpJson, "timestamp")

	responseJson, _ := json.Marshal(tmpJson)

	checkJsonResponse(t, responseJson, expectedJson)

	// TODO: check cluster data

	// TODO: check nearby cashpoints

	// now delete created cashpoint
	url, handlerDelete := handlerCashpointDelete(hCtx)
	request = TestRequest{
		RequestType: "DELETE",
		EndpointUrl: "/cashpoint/" + cashpointIdStr,
		HandlerUrl:  url,
	}

	response, err = readResponse(testRequest(request, handlerDelete))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	// try to get deleted cashpoint
	request = TestRequest{
		RequestType: "GET",
		EndpointUrl: "/cashpoint/" + cashpointIdStr,
		HandlerUrl:  urlGet,
	}

	response, err = readResponse(testRequest(request, handlerGet))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusNotFound)

	// get new quadtree branch state (after adding cashpoint)
	// there is no expected changes
	response, err = readResponse(testRequest(requestTreeBranch, handlerTreeBranch))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusOK)

	var clustersNew ClusterArray
	err = json.Unmarshal(response.Data, &clustersNew)
	if err != nil {
		t.Errorf("Cannot unpack quad tree branch response: %v", err)
	}

	if same, diffText := clusters.Compare(clustersNew); !same {
		t.Fatalf("%s\n\n%s\n%s\n%s",
			diffText,
			"ALERT! Quad tree branches before and after create + delete are different.",
			"Looks like quad tree is broken after test and data in tarantool is corrupted.",
			"Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script.")
	}
}

func TestCashpointCreateWrongCoordinates(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	longitude := 203.456
	latitude := 55.7591

	quadKeyReq := QuadKeyRequest{
		Longitude: longitude,
		Latitude:  latitude,
	}
	quadkeyReqJson, _ := json.Marshal(quadKeyReq)

	url, handlerQuadKey := handlerCoordToQuadKey(hCtx)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: url,
		Data:        string(quadkeyReqJson),
	}

	response, err := readResponse(testRequest(request, handlerQuadKey))
	if err != nil {
		t.Errorf("%v", err)
	}
	checkHttpCode(t, response.Code, http.StatusBadRequest)

	// creating real cashpoint with wrong coordinates
	cp := CashpointShort{
		Longitude:      longitude,
		Latitude:       latitude,
		Type:           "atm",
		BankId:         322, // Sberbank
		TownId:         4,   // Moscow
		Address:        "",
		AddressComment: "",
		//		MetroName: "",
		FreeAccess:     true,
		MainOffice:     false,
		WithoutWeekend: true,
		RoundTheClock:  false,
		WorksAsShop:    false,
		Schedule:       "",
		Tel:            "",
		Additional:     "",
		Rub:            true,
		Usd:            false,
		Eur:            false,
		CashIn:         true,
	}

	reqData := CashpointCreateRequest{
		UserId: 0, // TODO: check against real user
		Data:   cp,
	}
	reqJson, _ := json.Marshal(reqData)

	url, handlerCreate := handlerCashpointCreate(hCtx)
	request = TestRequest{
		RequestType: "POST",
		EndpointUrl: "/cashpoint",
		HandlerUrl:  url,
		Data:        string(reqJson),
	}

	response, err = readResponse(testRequest(request, handlerCreate))
	if err != nil {
		t.Errorf("%v", err)
	}
	if !checkHttpCode(t, response.Code, http.StatusInternalServerError) {
		// cashpoint created for some reason
		if response.Code == http.StatusOK {
			var cashpointId uint64 = 0
			err = json.Unmarshal(response.Data, &cashpointId)
			if err != nil {
				t.Fatalf(`ALERT! Looks like cashpoint created but its id was not returned.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script.`)
			} else {
				t.Fatalf(`ALERT! Looks like cashpoint created with id '%d'.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script
					  or delete cashpoint and following data manually.`, cashpointId)
			}
		}
	}
}

func TestCashpointCreateMissingRequredFields(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	longitude := 38.2371
	latitude := 56.4631

	// check quad tree branches for coordinate before and after test
	quadTreeBranch, err := getQuadTreeBranch(t, hCtx, longitude, latitude)
	if err != nil {
		t.Errorf("Failed to cache quad tree branch: %v", err)
	}
	defer checkQuadTreeBranch(t, func() ([]byte, error) { return getQuadTreeBranch(t, hCtx, longitude, latitude) }, quadTreeBranch)

	// creating real cashpoint with missing required fields
	cp := CashpointShort{
		Longitude:      longitude,
		Latitude:       latitude,
		Type:           "atm",
		BankId:         322, // Sberbank
		TownId:         4,   // Moscow
		Address:        "",
		AddressComment: "",
		//		MetroName: "",
		FreeAccess:     true,
		MainOffice:     false,
		WithoutWeekend: true,
		// 		RoundTheClock: false, // WARNING: here is missing field
		WorksAsShop: false,
		Schedule:    "",
		Tel:         "",
		Additional:  "",
		Rub:         true,
		Usd:         false,
		Eur:         false,
		CashIn:      true,
	}

	reqData := CashpointCreateRequest{
		UserId: 0, // TODO: check against real user
		Data:   cp,
	}
	reqJson, _ := json.Marshal(reqData)

	var tmpJson map[string]interface{}
	err = json.Unmarshal(reqJson, &tmpJson)
	if err != nil {
		t.Errorf("Failed unmarshal tmp CashpointCreateRequest: %v", err)
	}

	// repack with missing field
	data := tmpJson["data"].(map[string]interface{})
	delete(data, "round_the_clock")
	reqJson, _ = json.Marshal(tmpJson)

	url, handlerCreate := handlerCashpointCreate(hCtx)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: "/cashpoint",
		HandlerUrl:  url,
		Data:        string(reqJson),
	}

	response, err := readResponse(testRequest(request, handlerCreate))
	if err != nil {
		t.Errorf("%v", err)
	}
	if !checkHttpCode(t, response.Code, http.StatusInternalServerError) {
		// cashpoint created for some reason
		if response.Code == http.StatusOK {
			var cashpointId uint64 = 0
			err = json.Unmarshal(response.Data, &cashpointId)
			if err != nil {
				t.Fatalf(`ALERT! Looks like cashpoint created but its id was not returned.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script.`)
			} else {
				t.Fatalf(`ALERT! Looks like cashpoint created with id '%d'.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script
					  or delete cashpoint and following data manually.`, cashpointId)
			}
		}
	}
}

func TestCashpointCreateApproveHack(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	// check metrics before and after test
	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	longitude := 38.2371
	latitude := 56.4631

	// check quad tree branches for coordinate before and after test
	quadTreeBranch, err := getQuadTreeBranch(t, hCtx, longitude, latitude)
	if err != nil {
		t.Errorf("Failed to cache quad tree branch: %v", err)
	}
	defer checkQuadTreeBranch(t, func() ([]byte, error) { return getQuadTreeBranch(t, hCtx, longitude, latitude) }, quadTreeBranch)

	// creating real cashpoint with missing required fields
	cp := CashpointShort{
		Longitude:      longitude,
		Latitude:       latitude,
		Type:           "atm",
		BankId:         322, // Sberbank
		TownId:         4,   // Moscow
		Address:        "",
		AddressComment: "",
		//MetroName: "",
		FreeAccess:     true,
		MainOffice:     false,
		WithoutWeekend: true,
		//RoundTheClock: false, // WARNING: here is missing field
		WorksAsShop: false,
		Schedule:    "",
		Tel:         "",
		Additional:  "",
		Rub:         true,
		Usd:         false,
		Eur:         false,
		CashIn:      true,
	}

	reqData := CashpointCreateRequest{
		UserId: 0, // TODO: check against real user
		Data:   cp,
	}
	reqJson, _ := json.Marshal(reqData)

	var tmpJson map[string]interface{}
	err = json.Unmarshal(reqJson, &tmpJson)
	if err != nil {
		t.Errorf("Failed unmarshal tmp CashpointCreateRequest: %v", err)
	}

	// repack with unexpected field approved
	data := tmpJson["data"].(map[string]interface{})
	data["approved"] = true
	reqJson, _ = json.Marshal(tmpJson)

	url, handlerCreate := handlerCashpointCreate(hCtx)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: "/cashpoint",
		HandlerUrl:  url,
		Data:        string(reqJson),
	}

	response, err := readResponse(testRequest(request, handlerCreate))
	if err != nil {
		t.Errorf("%v", err)
	}
	// expecting validation failure
	if !checkHttpCode(t, response.Code, http.StatusInternalServerError) {
		// cashpoint created for some reason
		if response.Code == http.StatusOK {
			var cashpointId uint64 = 0
			err = json.Unmarshal(response.Data, &cashpointId)
			if err != nil {
				t.Fatalf(`ALERT! Looks like cashpoint created but its id was not returned.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script.`)
			} else {
				t.Fatalf(`ALERT! Looks like cashpoint created with id '%d'.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script
					  or delete cashpoint and following data manually.`, cashpointId)
			}
		}
	}
}

type Coordinate struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type NearByRequestFilter struct {
	BankId []uint32 `json:"bank_id"`
}

type NearByRequest struct {
	BottomRight Coordinate          `json:"bottomRight"`
	TopLeft     Coordinate          `json:"topLeft"`
	Filter      NearByRequestFilter `json:"filter"`
}

func TestFilterBankIdCount(t *testing.T) {
	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.close()

	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	reqNearBy := NearByRequest{
		BottomRight: Coordinate{
			Longitude: 12.0,
			Latitude:  13.0,
		},
		TopLeft: Coordinate{
			Longitude: 12.01,
			Latitude:  13.01,
		},
		Filter: NearByRequestFilter{
			BankId: []uint32{322, 325, 326, 327, 328, 329, 330, 331, 332, 333, 334, 335, 336, 337, 357, 338, 339, 340},
		},
	}

	url, handlerCreate := handlerNearbyCashPoints(hCtx)

	reqJson, _ := json.Marshal(reqNearBy)
	request := TestRequest{
		RequestType: "POST",
		EndpointUrl: url,
		Data:        string(reqJson),
	}

	response, err := readResponse(testRequest(request, handlerCreate))
	if err != nil {
		t.Errorf("%v", err)
	}

	if !checkHttpCode(t, response.Code, http.StatusInternalServerError) {
		// cashpoint created for some reason
		if response.Code == http.StatusOK {
			var cashpointId uint64 = 0
			err = json.Unmarshal(response.Data, &cashpointId)
			if err != nil {
				t.Fatalf(`ALERT! Looks like cashpoint created but its id was not returned.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script.`)
			} else {
				t.Fatalf(`ALERT! Looks like cashpoint created with id '%d'.
					  Please, refill database with fresh testing data again by running 'build_db_tnt.sh' script
					  or delete cashpoint and following data manually.`, cashpointId)
			}
		}
	}
}
