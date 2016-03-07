package main

import (
	"github.com/tarantool/go-tarantool"
	"log"
	"net/http"
	"strconv"
)

// TODO: access control
func handlerSpaceMetrics(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/metrics/space", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}
		go logRequest(w, r, requestId, "")

		context := getRequestContexString(r) + " " + getHandlerContextString("handlerSpaceMetrics", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		resp, err := tnt.Call("getSpaceMetrics", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get space metrics: %v\n", context, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr)
		} else {
			log.Printf("%s => cannot convert space metrics reply\n", context)
			writeHeader(w, r, requestId, http.StatusInternalServerError)
		}
	}
}
