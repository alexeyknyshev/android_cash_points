package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

type TestLogger struct {
	ch chan string
}

type Logger interface {
	getChan() chan string
	logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string) error
	logResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string) error
	prepareResponse(w http.ResponseWriter, r *http.Request) (bool, int64)
	writeResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string)
	writeHeader(w http.ResponseWriter, r *http.Request, requestId int64, code int)
}

func (logger TestLogger) getChan() chan string {
	return logger.ch
}

func (logger TestLogger) logWriter() {
	log.Println(<-logger.getChan())
}

func (logger TestLogger) logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string) error {
	go logger.logWriter()
	endpointStr := r.URL.Path
	if requestBody != "" {
		endpointStr = endpointStr + " =>"
		body := ""
		body = " " + requestBody
		endpointStr = endpointStr + body
	}
	logStr := getRequestContexString(r) + " Request: " + r.Method + " " + endpointStr
	//log.Printf("%s Request: %s %s", getRequestContexString(r), r.Method, endpointStr)
	logger.ch <- logStr
	return nil
}

func (logger TestLogger) logResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string) error {
	go logger.logWriter()
	endpointStr := r.URL.Path
	if responseBody != "" {
		endpointStr = endpointStr + " =>"
		body := ""
		body = " " + responseBody
		endpointStr = endpointStr + body
	}
	logStr := getRequestContexString(r) + " Response: " + r.Method + " " + endpointStr
	logger.ch <- logStr
	//log.Printf("%s: Response: %s %s", getRequestContexString(r), r.Method, endpointStr)
	return nil
}

func (logger TestLogger) prepareResponse(w http.ResponseWriter, r *http.Request) (bool, int64) {
	requestId, err := getRequestUserId(r)
	if err != nil {
		logStr := getRequestContexString(r) + " prepareResponse " + err.Error()
		logger.ch <- logStr
		//log.Printf("%s prepareResponse %v\n", getRequestContexString(r), err)
		w.WriteHeader(http.StatusBadRequest)
		return false, 0
	}
	if requestId == 0 {
		strReqId := strconv.FormatInt(requestId, 10)
		logStr := getRequestContexString(r) + " prepareResponse unexpected requestId: " + strReqId
		logger.ch <- logStr
		//log.Printf("%s prepareResponse unexpected requestId: %d\n", getRequestContexString(r), requestId)
		w.WriteHeader(http.StatusBadRequest)
		return false, 0
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Id", strconv.FormatInt(requestId, 10))
	return true, requestId
}

func (logger TestLogger) writeResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string) {
	io.WriteString(w, responseBody)
	logger.logResponse(w, r, requestId, responseBody)
}

func (logger TestLogger) writeHeader(w http.ResponseWriter, r *http.Request, requestId int64, code int) {
	w.WriteHeader(code)
	logger.logResponse(w, r, requestId, "code "+strconv.FormatInt(int64(code), 10))
}
