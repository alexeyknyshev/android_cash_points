package main

import (
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

var MAX_CLUSTER_COUNT uint64 = 32

var MIN_QUADKEY_LENGTH int = 10
var MAX_QUADKEY_LENGTH int = 16

func handlerCashpoint(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/cashpoint/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
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
			if jsonStr != "" {
				writeResponse(w, r, requestId, jsonStr)
			} else {
				w.WriteHeader(404)
			}
		} else {
			log.Printf("%s => cannot convert cashpoint reply for id: %d\n", context, cashPointId)
			w.WriteHeader(500)
		}
	}
}

func handlerCashpointsBatch(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/cashpoints", func(w http.ResponseWriter, r *http.Request) {
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

func handlerNearbyCashPoints(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/nearby/cashpoints", func(w http.ResponseWriter, r *http.Request) {
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

func handlerNearbyClusters(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/nearby/clusters", func(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("%s => cannot get nearby clusters: %v => %s\n", context, err, jsonStr)
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

func handlerQuadTreeBranch(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/quadtree/branch/{quadKey:[0-3]+}", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}
		go logRequest(w, r, requestId, "")

		params := mux.Vars(r)
		quadKeyStr := params["quadKey"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerQuadTreeBranch", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"quadKey": quadKeyStr,
		})

		quadKeyStrLen := len(quadKeyStr)
		if quadKeyStrLen > MAX_QUADKEY_LENGTH || quadKeyStrLen < MIN_QUADKEY_LENGTH {
			w.WriteHeader(400)
			return
		}

		resp, err := tnt.Call("getQuadTreeBranch", []interface{}{ quadKeyStr })
		if err != nil {
			log.Printf("%s => cannot get quad tree branch: %v => %s\n", context, err, quadKeyStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert quad tree branch reply to json str: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}

func handlerCashpointCreate(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/cashpoint", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointCreate", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("cashpointProposePatch", []interface{}{ jsonStr })
		if err != nil {
			log.Printf("%s => cannot propose patch: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if cashpointId, ok := data.(uint64); ok {
			if cashpointId != 0 {
				jsonData, _ := json.Marshal(cashpointId)
				writeResponse(w, r, requestId, string(jsonData))
			} else {
				w.WriteHeader(500)
			}
		} else {
			log.Printf("%s => cannot convert response to uint64 for request json str: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}

func handlerCashpointDelete(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/cashpoint/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		params := mux.Vars(r)
		cashPointIdStr := params["id"]

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointDelete", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
			"cashPointId": cashPointIdStr,
		})

		cashPointId, err := strconv.ParseUint(cashPointIdStr, 10, 64)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, "")

		resp, err := tnt.Call("deleteCashpointById", []interface{}{ cashPointId })
		if err != nil {
			log.Printf("%s => cannot delete cashpoint by id: %v => %s\n", context, err, cashPointIdStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if done, ok := data.(bool); ok {
			if done {
				w.WriteHeader(200)
			} else {
				w.WriteHeader(404)
			}
		} else {
			log.Printf("%s => cannot convert response to bool for request cashpoint id: %s\n", context, cashPointIdStr)
			w.WriteHeader(500)
		}
	}
}

func handlerCoordToQuadKey(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/quadkey", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerCoordToQuadKey", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		jsonStr, err := getRequestJsonStr(r, context)
		if err != nil {
			go logRequest(w, r, requestId, "")
			w.WriteHeader(400)
			return
		}

		go logRequest(w, r, requestId, jsonStr)

		resp, err := tnt.Call("getQuadKeyFromCoord", []interface{}{ jsonStr })
		if err != nil {
			log.Printf("%s => cannot convert coord to quadkey: %v => %s\n", context, err, jsonStr)
			w.WriteHeader(500)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStrResp, ok := data.(string); ok {
			if jsonStrResp != "" {
				writeResponse(w, r, requestId, jsonStrResp)
			} else {
				w.WriteHeader(400)
			}
		} else {
			log.Printf("%s => cannot convert response for quadkey from coord: %s\n", context, jsonStr)
			w.WriteHeader(500)
		}
	}
}
