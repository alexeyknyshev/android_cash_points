package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	//"github.com/tarantool/go-tarantool"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

func handlerBank(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/bank/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}
		logger.logRequest(w, r, requestId, "")

		params := mux.Vars(r)
		bankIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBank", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"bankId":    bankIdStr,
		})

		bankId, err := strconv.ParseUint(bankIdStr, 10, 64)
		if err != nil {
			logger.writeHeader(w, r, requestId, http.StatusBadRequest)
			return
		}

		resp, err := tnt.Call("getBankById", []interface{}{bankId})
		if err != nil {
			log.Printf("%s => cannot get bank %d by id: %v\n", context, bankId, err)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		if len(resp.Data) == 0 {
			log.Printf("%s => no such bank with id: %d\n", context, bankId)
			logger.writeHeader(w, r, requestId, http.StatusNotFound)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			if jsonStr != "" {
				logger.writeResponse(w, r, requestId, jsonStr)
			} else {
				logger.writeHeader(w, r, requestId, http.StatusNotFound)
			}
		} else {
			log.Printf("%s => cannot convert bank reply for id: %d\n", context, bankId)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}

type BankIco struct {
	BankId  uint32 `json:"bank_id"`
	IcoData string `json:"ico_data"`
}

func handlerBankIco(handlerContext HandlerContext, conf ServerConfig) (string, EndpointCallback) {
	return "/bank/{id:[0-9]+}/ico", func(w http.ResponseWriter, r *http.Request) {
		//tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
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
			logger.writeHeader(w, r, requestId, http.StatusNotFound)
			return
		}

		data, err := ioutil.ReadFile(icoFilePath)
		if err != nil {
			log.Printf("%s => cannot read file: %s", context, icoFilePath)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		id, err := strconv.ParseUint(bankIdStr, 10, 32)
		bankId := checkConvertionUint(uint32(id), err, context+" => BankIco.BankId")

		ico := &BankIco{BankId: bankId, IcoData: string(data)}
		jsonByteArr, _ := json.Marshal(ico)
		logger.writeResponse(w, r, requestId, string(jsonByteArr))
	}
}

func handlerBanksBatch(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/banks", func(w http.ResponseWriter, r *http.Request) {
		tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			logger.logRequest(w, r, requestId, "")
			logger.writeHeader(w, r, requestId, http.StatusBadRequest)
			return
		}

		logger.logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getBanksBatch", []interface{}{jsonStr})
		if err != nil {
			log.Printf("%s => cannot get banks batch: %v => %s\n", context, err, jsonStr)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			logger.writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert banks batch reply to json str: %s\n", context, jsonStr)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}

func handlerBanksList(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/banks", func(w http.ResponseWriter, r *http.Request) {
		tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		logger.logRequest(w, r, requestId, "")

		resp, err := tnt.Call("getBanksList", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get banks list: %v\n", context, err)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			logger.writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert banks list reply to json str\n", context, jsonStr)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}
