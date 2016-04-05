package main

import (
	//"bytes"
	"encoding/json"
	//"errors"
	"fmt"
	//"github.com/alexeyknyshev/gojsondiff"
	//"github.com/alexeyknyshev/gojsondiff/formatter"
	//"github.com/gorilla/mux"
	//"io/ioutil"
	//"log"
	//"net/http"
	//"net/http/httptest"
	//"sort"
	"reflect"
	"strconv"
	"testing"
)

type PatchRequest struct {
	Id             uint32   `json:"id,omitempty"`
	Longitude      float64  `json:"longitude,omitempty"`
	Latitude       float64  `json:"latitude,omitempty"`
	Type           string   `json:"type,omitempty"`
	BankId         uint32   `json:"bank_id,omitempty"`
	TownId         uint32   `json:"town_id,omitempty"`
	Address        string   `json:"address,omitempty"`
	AddressComment string   `json:"address_comment,omitempty"`
	MetroName      string   `json:"metro_name,omitempty"`
	FreeAccess     *bool    `json:"free_access,omitempty"`
	MainOffice     *bool    `json:"main_office,omitempty"`
	WithoutWeekend *bool    `json:"without_weekend,omitempty"`
	RoundTheClock  *bool    `json:"round_the_clock,omitempty"`
	WorksAsShop    *bool    `json:"works_as_shop,omitempty"`
	Schedule       Schedule `json:"schedule,omitempty"`
	Tel            string   `json:"tel,omitempty"`
	Additional     string   `json:"additional,omitempty"`
	Rub            *bool    `json:"rub,omitempty"`
	Usd            *bool    `json:"usd,omitempty"`
	Eur            *bool    `json:"eur,omitempty"`
	CashIn         *bool    `json:"cash_in,omitempty"`
}

func getPatchExampleNewCP() *PatchRequest {
	False := false
	True := true
	patchReq := PatchRequest{
		Longitude:      37.6878262,
		Latitude:       55.6946643,
		Type:           "atm",
		BankId:         322,
		TownId:         4,
		Address:        "г. Москва, Район Моей Мечты",
		AddressComment: "ОАО UnderButtom",
		MetroName:      "",
		FreeAccess:     &True,
		MainOffice:     &False,
		WithoutWeekend: &False,
		RoundTheClock:  &False,
		WorksAsShop:    &True,
		Schedule:       Schedule{},
		Tel:            "",
		Additional:     "",
		Rub:            &True,
		Usd:            &False,
		Eur:            &False,
		CashIn:         &False,
	}
	return &patchReq

}

func getPatchExampleExistCP() (*PatchRequest, string) {
	patchReq := PatchRequest{
		Id:     58552,
		BankId: 2764,
	}
	exampleJson := "{\"schedule\":{},\"bank_id\":" + strconv.FormatUint(uint64(patchReq.BankId), 10) + "}"
	return &patchReq, exampleJson

}

func searchLastPatch(t *testing.T, resJsonPatches []byte) uint32 {
	var CPPatches map[string]interface{}
	err := json.Unmarshal(resJsonPatches, &CPPatches)
	if err != nil {
		t.Errorf("Unmarshal err %v", err)
	}
	last_key := uint64(0)
	//Search last patch number
	for key := range CPPatches {
		int_key, _ := strconv.ParseUint(key, 10, 64)
		if int_key > last_key {
			last_key = int_key
		}
	}
	return uint32(last_key)
}

func comparePatches(t *testing.T, resPatch []interface{}, expectedPatch []interface{}) {
	if len(resPatch) != len(expectedPatch) {
		t.Error("response and expected patches have different field amount")
		return
	}

	fields := []string{"patch id", "cashpoint id", "user_id", "data", "timestamp"}
	for i, vol := range expectedPatch {
		if i == 3 { //PATCH_DATA
			checkJsonResponse(t, []byte(vol.(string)), []byte(resPatch[i].(string)))
		} else if i != 4 { //don't check timestamp
			if resPatch[i].(uint64) != vol.(uint64) {
				t.Error("comparePatches: fields", fields[i], "don't match")
			}
		}
	}
}

func TestPatchCreateNewCP(t *testing.T) {

	type Req struct {
		Data   PatchRequest `json:"data"`
		UserId uint         `json:"user_id"`
	}
	patchReq := getPatchExampleNewCP()
	request := Req{
		Data:   *patchReq,
		UserId: 0,
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Json Marshal error %v", err)
	}
	fmt.Println("Request:\n", string(requestJson), "\n")

	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.Close()

	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	TaranResp, err := hCtx.Tnt().Call("cashpointProposePatch", []interface{}{requestJson})
	if err != nil {
		t.Errorf("Tnt cashpointProposePatch call err: %v", err)
	}
	CpId := TaranResp.Data[0].([]interface{})[0]

	fmt.Println("\nCreate new patch, CpId = ", CpId)

	resp, err := hCtx.Tnt().Call("getCashpointPatches", []interface{}{CpId})
	if err != nil {
		t.Errorf("Tnt getCashpointPatches call err: %v", err)
	}

	//Json response parse and compare
	byteResp := []byte(resp.Data[0].([]interface{})[0].(string))

	lastPatch := searchLastPatch(t, byteResp)
	resp, err = hCtx.Tnt().Call("getCashpointPatchByPatchId", []interface{}{lastPatch})
	resPatch := resp.Data[0].([]interface{})
	expPatchData := "{\"id\":" + strconv.FormatInt(int64(CpId.(uint64)), 10) + "}"
	expectedPatch := []interface{}{uint64(lastPatch), CpId.(uint64), uint64(request.UserId), expPatchData, uint64(0)}
	comparePatches(t, resPatch, expectedPatch)
	fmt.Println("Delete patch ", lastPatch)
	resp, err = hCtx.Tnt().Call("deleteCashpointById", []interface{}{CpId})
	if err != nil {
		t.Errorf("Tnt call _deleteCashpointPatchById err: %v", err)
	}
}

func TestPatchChangeExistCP(t *testing.T) {

	type Req struct {
		Data   PatchRequest `json:"data"`
		UserId uint         `json:"user_id"`
	}
	patchReq, expPatchData := getPatchExampleExistCP()
	request := Req{
		Data:   *patchReq,
		UserId: 1,
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Json Marshal error %v", err)
	}
	fmt.Println("Request:\n", string(requestJson), "\n")

	hCtx, err := makeHandlerContext(getServerConfig())
	if err != nil {
		t.Fatalf("Connection to tarantool failed: %v", err)
	}
	defer hCtx.Close()

	metrics, err := getSpaceMetrics(hCtx)
	if err != nil {
		t.Errorf("Failed to get space metric on start: %v", err)
	}
	defer checkSpaceMetrics(t, func() ([]byte, error) { return getSpaceMetrics(hCtx) }, metrics)

	TaranResp, err := hCtx.Tnt().Call("cashpointProposePatch", []interface{}{requestJson})
	if err != nil {
		t.Errorf("Tnt cashpointProposePatch call err: %v", err)
		return
	}
	CpId := TaranResp.Data[0].([]interface{})[0]
	fmt.Print(reflect.TypeOf(CpId), " Volume = ", CpId)
	if CpId == uint64(0) {
		t.Error("Failed to create patch, CpId == 0")
		return
	} else {
		fmt.Println("\nCreate new patch, CpId = ", CpId)
	}

	CPPatchesTaranResp, err := hCtx.Tnt().Call("getCashpointPatches", []interface{}{CpId})
	if err != nil {
		t.Errorf("Tnt getCashpointPatches call err: %v", err)
	}
	CPPatchesJson := CPPatchesTaranResp.Data[0].([]interface{})[0].(string)

	fmt.Println("\nPatches of cashpoint №", CpId, ":\n", CPPatchesJson, "\n")
	CPPatchesJsonByte := []byte(CPPatchesJson)
	var CPPatches map[string]interface{}
	err = json.Unmarshal(CPPatchesJsonByte, &CPPatches)
	if err != nil {
		t.Errorf("Patches unmarshal err: %v", err)
	}

	lastPatch := searchLastPatch(t, CPPatchesJsonByte)
	resp, err := hCtx.Tnt().Call("getCashpointPatchByPatchId", []interface{}{lastPatch})
	resPatch := resp.Data[0].([]interface{})
	expectedPatch := []interface{}{uint64(lastPatch), CpId.(uint64), uint64(request.UserId), expPatchData, uint64(0)}
	comparePatches(t, resPatch, expectedPatch)

	fmt.Println("response patch:\n", resPatch)

	fmt.Println("Delete patch ", lastPatch)
	resp, err = hCtx.Tnt().Call("_deleteCashpointPatchById", []interface{}{lastPatch})
	if err != nil {
		t.Errorf("Tnt _deleteCashpointPatchById call err: %v", err)
	}
}
