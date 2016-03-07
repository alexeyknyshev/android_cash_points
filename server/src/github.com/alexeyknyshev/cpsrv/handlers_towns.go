package main

import (
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"log"
	"net/http"
	"strconv"
)

func handlerTown(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/town/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}
		go logRequest(w, r, requestId, "")

		params := mux.Vars(r)
		townIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTown", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"townId":    townIdStr,
		})

		townId, err := strconv.ParseUint(townIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := tnt.Call("getTownById", []interface{}{townId})
		if err != nil {
			log.Printf("%s => cannot get town %d by id: %v\n", context, townId, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(resp.Data) == 0 {
			log.Printf("%s => no such town with id: %d\n", context, townId)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			if jsonStr != "" {
				writeResponse(w, r, requestId, jsonStr)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			log.Printf("%s => cannot convert town reply for id: %d\n", context, townId)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func handlerTownsBatch(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/towns", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getTownsBatch", []interface{}{jsonStr})
		if err != nil {
			log.Printf("%s => cannot get towns batch: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert towns batch reply to json str: %s\n", context, jsonStr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func handlerTownsList(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/towns", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownsList", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		go logRequest(w, r, requestId, "")

		resp, err := tnt.Call("getTownsList", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get towns list: %v\n", context, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert towns list reply to json str\n", context, jsonStr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
