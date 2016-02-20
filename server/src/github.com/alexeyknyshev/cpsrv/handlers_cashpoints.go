package main

import (
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"log"
	"net/http"
	"strconv"
)

var MAX_CLUSTER_COUNT uint64 = 32

func handlerCashpoint(tnt *tarantool.Connection) EndpointCallback {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}
		go logRequest(w, r, requestId, "")

		params := mux.Vars(r)
		cashPointIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpoint", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"cashPointId": cashPointIdStr,
		})

		cashPointId, err := strconv.ParseUint(cashPointIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		resp, err := tnt.Call("getCashpointById", []interface{}{ cashPointId })
		if err != nil {
			log.Printf("%s => cannot get cashpoint %d by id: %v\n", context, cashPointId, err)
			w.WriteHeader(500)
			return
		}

		if (len(resp.Data) == 0) {
			log.Printf("%s => no such cashpoint with id: %d\n", context, cashPointId)
			w.WriteHeader(404)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert cashpoint reply for id: %d\n", context, cashPointId)
			w.WriteHeader(500)
		}
	}
}

func handlerCashpointsBatch(tnt *tarantool.Connection) EndpointCallback {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointsBatch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getCashpointsBatch", []interface{}{ jsonStr })
		if err != nil {
			log.Printf("%s => cannot get cashpoints batch: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert cashpoints batch reply to json str: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}

func handlerNearbyCashPoints(tnt *tarantool.Connection) EndpointCallback {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerNearbyCashPoints", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getNearbyCashpoints", []interface{}{ jsonStr })
		if err != nil {
			log.Printf("%s => cannot get neraby cashpoints: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert nearby cashpoints batch reply to json str: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}

func handlerNearbyClusters(tnt *tarantool.Connection) EndpointCallback {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerNearbyClusters", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getNearbyClusters", []interface{}{ jsonStr, MAX_CLUSTER_COUNT })
		if err != nil {
			log.Printf("%s => cannot get neraby clusters: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert nearby clusters batch reply to json str: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}
