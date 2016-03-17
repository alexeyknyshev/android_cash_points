package main

import (
	"log"
	"net/http"
)

type TestLogger struct {
	ch chan string
}

type Logger interface {
	getChan() chan string
	logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string) error
	logResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string) error
	logWriter(logStr string)
}

func (logger TestLogger) getChan() chan string {
	return logger.ch
}
func (logger TestLogger) goRoutineLogWriter() {
	log.Println(<-logger.getChan())
}

func (logger TestLogger) logWriter(logStr string) {
	go logger.goRoutineLogWriter()
	logger.ch <- logStr
}

func (logger TestLogger) logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string) error {
	go logger.goRoutineLogWriter()
	endpointStr := r.URL.Path
	if requestBody != "" {
		endpointStr = endpointStr + " =>"
		body := ""
		body = " " + requestBody
		endpointStr = endpointStr + body
	}
	logStr := getRequestContexString(r) + " Request: " + r.Method + " " + endpointStr
	logger.ch <- logStr
	return nil
}

func (logger TestLogger) logResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string) error {
	go logger.goRoutineLogWriter()
	endpointStr := r.URL.Path
	if responseBody != "" {
		endpointStr = endpointStr + " =>"
		body := ""
		body = " " + responseBody
		endpointStr = endpointStr + body
	}
	logStr := getRequestContexString(r) + " Response: " + r.Method + " " + endpointStr
	logger.ch <- logStr
	return nil
}
