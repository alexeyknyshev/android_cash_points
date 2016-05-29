package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func handlerTown(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/town/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
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
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		resp, err := handlerContext.Tnt().Call("getTownById", []interface{}{townId})
		if err != nil {
			log.Printf("%s => cannot get town %d by id: %v\n", context, townId, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		if len(resp.Data) == 0 {
			log.Printf("%s => no such town with id: %d\n", context, townId)
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
			log.Printf("%s => cannot convert town reply for id: %d\n", context, townId)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}

func handlerTownsBatch(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/towns", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			logger.logRequest(w, r, requestId, "")
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		logger.logRequest(w, r, requestId, jsonStr)

		resp, err := handlerContext.Tnt().Call("getTownsBatch", []interface{}{jsonStr})
		if err != nil {
			log.Printf("%s => cannot get towns batch: %v => %s\n", context, err, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert towns batch reply to json str: %s\n", context, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}

func handlerTownsList(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/towns", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownsList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		logger.logRequest(w, r, requestId, "")

		resp, err := handlerContext.Tnt().Call("getTownsList", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get towns list: %v\n", context, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert towns list reply to json str\n", context, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}
