package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func handlerTown(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/town/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}
		logger.logRequest(w, r, requestId, "")

		params := mux.Vars(r)
		townIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTown", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"townId":    townIdStr,
		})

		townId, err := strconv.ParseUint(townIdStr, 10, 64)
		if err != nil {
			logger.writeHeader(w, r, requestId, http.StatusBadRequest)
			return
		}

		resp, err := tnt.Call("getTownById", []interface{}{townId})
		if err != nil {
			log.Printf("%s => cannot get town %d by id: %v\n", context, townId, err)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		if len(resp.Data) == 0 {
			log.Printf("%s => no such town with id: %d\n", context, townId)
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
			log.Printf("%s => cannot convert town reply for id: %d\n", context, townId)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}

func handlerTownsBatch(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/towns", func(w http.ResponseWriter, r *http.Request) {
		tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			logger.logRequest(w, r, requestId, "")
			logger.writeHeader(w, r, requestId, http.StatusBadRequest)
			return
		}

		logger.logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getTownsBatch", []interface{}{jsonStr})
		if err != nil {
			log.Printf("%s => cannot get towns batch: %v => %s\n", context, err, jsonStr)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			logger.writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert towns batch reply to json str: %s\n", context, jsonStr)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}

func handlerTownsList(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/towns", func(w http.ResponseWriter, r *http.Request) {
		tnt := handlerContext.tnt()
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownsList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		logger.logRequest(w, r, requestId, "")

		resp, err := tnt.Call("getTownsList", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get towns list: %v\n", context, err)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			logger.writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert towns list reply to json str\n", context, jsonStr)
			logger.writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}
