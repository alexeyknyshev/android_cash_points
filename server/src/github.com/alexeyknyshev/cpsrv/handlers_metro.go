package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func handlerMetroList(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/town/{townid:[0-9]+}/metro", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}

		params := mux.Vars(r)
		townIdStr := params["townid"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerMetroList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"townId":    townIdStr,
		})
		logger.logRequest(w, r, requestId, "")

		townId, err := strconv.ParseUint(townIdStr, 10, 64)
		if err != nil {
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		resp, err := handlerContext.Tnt().Call("getMetroList", []interface{}{townId})
		if err != nil {
			log.Printf("%s => cannot get metro list: %v\n", context, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert metro list reply to json str\n", context, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}

func handlerMetro(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/metro/{metroid:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}

		params := mux.Vars(r)
		metroIdStr := params["metroid"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerMetro", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"metroId":   metroIdStr,
		})
		logger.logRequest(w, r, requestId, "")

		metroId, err := strconv.ParseUint(metroIdStr, 10, 64)
		if err != nil {
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		resp, err := handlerContext.Tnt().Call("getMetroById", []interface{}{metroId})
		if err != nil {
			log.Printf("%s => cannot get metro tuple: %v\n", context, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		if len(resp.Data) == 0 {
			log.Printf("%s => no such metro with metro id:%d\n", context, metroId)
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
			log.Printf("%s => cannot convert metro reply for metro id:%d\n", context, metroId)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}

func handlerMetroBatch(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/metro", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}
		context := getRequestContexString(r) + " " + getHandlerContextString("handlerMetroBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			logger.logRequest(w, r, requestId, "")
			writeHeader(w, r, requestId, http.StatusBadRequest, logger)
			return
		}

		logger.logRequest(w, r, requestId, jsonStr)
		resp, err := handlerContext.Tnt().Call("getMetroBatch", []interface{}{jsonStr})
		if err != nil {
			log.Printf("%s => cannot get metro batch: %v => %s\n", context, err, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert metro batch reply to json str: %s\n", context, jsonStr)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}
