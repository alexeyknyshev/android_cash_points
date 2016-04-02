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
	//"strconv"
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

func getPatchExampleExistCP() *PatchRequest {
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

func TestPatchCreating(t *testing.T) {

	type Req struct {
		Data   PatchRequest `json:"data"`
		UserId uint         `json:"user_id"`
	}
	patchReq := getPatchExampleNewCP()
	fmt.Println("PatchReq.Eur = ", *(patchReq.Eur))
	request := Req{
		Data:   *patchReq,
		UserId: 0,
	}
	requestJson, err := json.Marshal(request)
	fmt.Print("\n\n", string(requestJson), "\n\n")
	if err != nil {
		t.Fatalf("Json Marshal error %v", err)
	}

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
		t.Errorf("Failed to create patch: %v", err)
	}
	CpId := TaranResp.Data[0].([]interface{})[0]

	fmt.Print("\nCreate new patch, CpId = ", CpId)

	resp, err := hCtx.Tnt().Call("getCashpointPatches", []interface{}{CpId})
	if err != nil {
		t.Errorf("Failed to unmarshal %v", err)
	}
	fmt.Print("\nCheck new patch, CpId :\n ", resp)

	resp, err = hCtx.Tnt().Call("deleteCashpointById", []interface{}{CpId})

}

func func_name() {

}
