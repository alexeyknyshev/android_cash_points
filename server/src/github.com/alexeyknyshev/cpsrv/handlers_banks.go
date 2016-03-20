package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

func handlerBank(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/bank/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
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
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		resp, err := handlerContext.Tnt().Call("getBankById", []interface{}{bankId})
		if err != nil {
			log.Printf("%s => cannot get bank %d by id: %v\n", context, bankId, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		if len(resp.Data) == 0 {
			log.Printf("%s => no such bank with id: %d\n", context, bankId)
			writeHeader(w, r, requestId, http.StatusNotFound, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			if jsonStr != "" {
				writeResponse(w, r, requestId, jsonStr, logger)
			} else {
				writeHeader(w, r, requestId, http.StatusNotFound, logger)
			}
		} else {
			log.Printf("%s => cannot convert bank reply for id: %d\n", context, bankId)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}

type BankIco struct {
	BankId  uint32 `json:"bank_id"`
	IcoData string `json:"ico_data"`
}

func handlerBankIco(handlerContext HandlerContext, conf ServerConfig) (string, EndpointCallback) {
	return "/bank/{id:[0-9]+}/ico", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
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
			writeHeader(w, r, requestId, http.StatusNotFound, logger)
			return
		}

		data, err := ioutil.ReadFile(icoFilePath)
		if err != nil {
			log.Printf("%s => cannot read file: %s", context, icoFilePath)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		id, err := strconv.ParseUint(bankIdStr, 10, 32)
		bankId := checkConvertionUint(uint32(id), err, context+" => BankIco.BankId")

		ico := &BankIco{BankId: bankId, IcoData: string(data)}
		jsonByteArr, _ := json.Marshal(ico)
		writeResponse(w, r, requestId, string(jsonByteArr), logger)
	}
}

func handlerBanksBatch(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/banks", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			logger.logRequest(w, r, requestId, "")
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		logger.logRequest(w, r, requestId, jsonStr)

		resp, err := handlerContext.Tnt().Call("getBanksBatch", []interface{}{jsonStr})
		if err != nil {
			log.Printf("%s => cannot get banks batch: %v => %s\n", context, err, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert banks batch reply to json str: %s\n", context, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}

func handlerBanksList(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/banks", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		logger.logRequest(w, r, requestId, "")

		resp, err := handlerContext.Tnt().Call("getBanksList", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get banks list: %v\n", context, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert banks list reply to json str\n", context, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}
