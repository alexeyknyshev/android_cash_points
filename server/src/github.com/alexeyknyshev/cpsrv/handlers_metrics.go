package main

import (
	"log"
	"net/http"
	"strconv"
)

// TODO: access control
func handlerSpaceMetrics(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/metrics/space", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.Logger()
		ok, requestId := prepareResponse(w, r, logger)
		if ok == false {
			return
		}
		logger.logRequest(w, r, requestId, "")
		context := getRequestContexString(r) + " " + getHandlerContextString("handlerSpaceMetrics", map[string]string{
			"requestId": strconv.FormatInt(requestId, 10),
		})

		resp, err := handlerContext.Tnt().Call("getSpaceMetrics", []interface{}{})
		if err != nil {
			log.Printf("%s => cannot get space metrics: %v\n", context, err)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
			return
		}

		data := resp.Data[0].([]interface{})[0]
		if jsonStr, ok := data.(string); ok {
			writeResponse(w, r, requestId, jsonStr, logger)
		} else {
			log.Printf("%s => cannot convert space metrics reply\n", context)
			writeHeader(w, r, requestId, http.StatusInternalServerError, logger)
		}
	}
}
