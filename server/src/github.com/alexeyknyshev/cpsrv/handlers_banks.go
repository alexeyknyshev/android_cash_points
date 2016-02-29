package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"io/ioutil"
	"log"
	"os"
	"net/http"
	"path"
	"strconv"
)

func handlerBank(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/bank/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}
		go logRequest(w, r, requestId, "")

		params := mux.Vars(r)
		bankIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBank", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"bankId":    bankIdStr,
		})

		bankId, err := strconv.ParseUint(bankIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		resp, err := tnt.Call("getBankById", []interface{}{ bankId })
		if err != nil {
			log.Printf("%s => cannot get bank %d by id: %v\n", context, bankId, err)
			w.WriteHeader(500)
			return
		}

		if (len(resp.Data) == 0) {
			log.Printf("%s => no such bank with id: %d\n", context, bankId)
			w.WriteHeader(404)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			if jsonStr != "" {
				writeResponse(w, r, requestId, jsonStr)
			} else {
				w.WriteHeader(404)
			}
		} else {
			log.Printf("%s => cannot convert bank reply for id: %d\n", context, bankId)
			w.WriteHeader(500)
		}
	}
}

type BankIco struct {
	BankId  uint32 `json:"bank_id"`
	IcoData string `json:"ico_data"`
}

func handlerBankIco(conf ServerConfig) (string, EndpointCallback) {
	return "/bank/{id:[0-9]+}/ico", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		params := mux.Vars(r)
		bankIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBankIco", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"bankId":    bankIdStr,
		})

		icoFilePath := path.Join(conf.BanksIcoDir, bankIdStr+".svg")

		if _, err := os.Stat(icoFilePath); os.IsNotExist(err) {
			w.WriteHeader(404)
			return
		}

		data, err := ioutil.ReadFile(icoFilePath)
		if err != nil {
			log.Printf("%s => cannot read file: %s", context, icoFilePath)
			w.WriteHeader(500)
			return
		}

		id, err := strconv.ParseUint(bankIdStr, 10, 32)
		bankId := checkConvertionUint(uint32(id), err, context+" => BankIco.BankId")

		ico := &BankIco{BankId: bankId, IcoData: string(data)}
		jsonByteArr, _ := json.Marshal(ico)
		writeResponse(w, r, requestId, string(jsonByteArr))
	}
}

func handlerBanksBatch(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/banks", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getBanksBatch", []interface{}{ jsonStr })
		if err != nil {
			log.Printf("%s => cannot get banks batch: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert banks batch reply to json str: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}

func handlerBanksList(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/banks", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		go logRequest(w, r, requestId, "")

		resp, err := tnt.Call("getBanksList", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get banks list: %v\n", context, err)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert banks list reply to json str\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}
