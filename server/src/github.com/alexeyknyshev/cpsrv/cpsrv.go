package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const SERVER_DEFAULT_CONFIG = "config.json"

type ServerConfig struct {
	TownsDataBase      string `json:"TownsDataBase"`
	CashPointsDataBase string `json:"CashPointsDataBase"`
	CertificateDir     string `json:"CertificateDir"`
	Port               uint64 `json:"Port"`
	UserLoginMinLength uint64 `json:"UserLoginMinLength"`
	UserPwdMinLength   uint64 `json:"UserPwdMinLength"`
	UseTLS             bool   `json:"UseTLS"`
	RedisHost          string `json:"RedisHost"`
	RedisScriptsDir    string `json:"RedisScriptsDir"`
	ReqResLogTTL       uint64 `json:"ReqResLogTTL"`
	UUID_TTL           uint64 `json:"UUID_TTL"`
	BanksIcoDir        string `json:"BanksIcoDir"`
	TestingMode        bool   `json:"TestingMode"`
	TntUser            string `json:"TntUser"`
	TntPass            string `json:"TntPass"`
	TntUrl             string `json:"TntUrl"`
}

type Message struct {
	Text string `json:"text"`
}

type HandlerContextStruct struct {
	Tnt *tarantool.Connection
	Logger
}

type HandlerContext interface {
	tnt() *tarantool.Connection
	logger() Logger
}

func (handler HandlerContextStruct) tnt() *tarantool.Connection {
	return handler.Tnt
}

func (handler HandlerContextStruct) logger() Logger {
	return handler.Logger
}

func makeHandlerContext(tnt *tarantool.Connection) HandlerContext {
	logger := &TestLogger{make(chan string)}
	handlerContext := &HandlerContextStruct{tnt, logger}
	return handlerContext
}

func checkConvertionUint(val uint32, err error, context string) uint32 {
	if err != nil {
		log.Printf("%s: uint conversion err => %v\n", context, err)
		return 0
	}
	return val
}

func getRequestContexString(r *http.Request) string {
	return r.RemoteAddr
}

func getHandlerContextString(funcName string, args map[string]string) string {
	result := funcName + "("
	i := 0
	argsCount := len(args)
	for argName, argVal := range args {
		result = result + argName + "=" + argVal
		if i < argsCount-1 {
			result = result + ","
		}
		i++
	}
	result = result + ")"

	return result
}

func getRequestUserId(r *http.Request) (int64, error) {
	requestIdStr := r.Header.Get("Id")
	if requestIdStr == "" {
		return 0, errors.New(`Request header val "Id" is not set`)
	}
	requestId, err := strconv.ParseInt(requestIdStr, 10, 64)
	if err != nil {
		return 0, errors.New(`Request header val "Id" uint conversion failed: ` + requestIdStr)
	}
	return requestId, nil
}

func getRequestJsonStr(r *http.Request, context string) (string, error) {
	jsonStr, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s => malformed json\n", context)
		return "", err
	}
	return string(jsonStr), nil
}

type EndpointCallback func(w http.ResponseWriter, r *http.Request)

func handlerPing(handlerContext HandlerContext) (string, EndpointCallback) {
	return "/ping", func(w http.ResponseWriter, r *http.Request) {
		logger := handlerContext.logger()
		ok, requestId := logger.prepareResponse(w, r)
		if ok == false {
			return
		}

		logger.logRequest(w, r, requestId, "")
		msg := &Message{Text: "pong"}
		jsonByteArr, _ := json.Marshal(msg)
		logger.writeResponse(w, r, requestId, string(jsonByteArr))
	}
}

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	args := os.Args[1:]

	configFilePath := SERVER_DEFAULT_CONFIG
	if len(args) > 0 {
		configFilePath = args[0]
		log.Printf("Loading config file: %s\n", configFilePath)
	} else {
		log.Printf("Loading default config file: %s\n", configFilePath)
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Fatalf("No such config file: %s\n", configFilePath)
	}

	configFile, _ := os.Open(configFilePath)
	decoder := json.NewDecoder(configFile)
	serverConfig := ServerConfig{}
	err := decoder.Decode(&serverConfig)
	if err != nil {
		log.Fatalf("Failed to decode config file: %s\nError: %v\n", configFilePath, err)
		return
	}

	if serverConfig.TestingMode {
		log.Printf("WARNING: Server started is TESTING mode! Make sure it is not prod server.")
	}

	opts := tarantool.Opts{
		Reconnect:     1 * time.Second,
		MaxReconnects: 3,
		User:          serverConfig.TntUser,
		Pass:          serverConfig.TntPass,
	}
	tnt, err := tarantool.Connect(serverConfig.TntUrl, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer tnt.Close()

	handlerContext := makeHandlerContext(tnt)
	// f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()
	// log.SetOutput(f)

	router := mux.NewRouter()
	router.HandleFunc(handlerPing(handlerContext)).Methods("GET")
	router.HandleFunc(handlerCashpoint(handlerContext)).Methods("GET")
	router.HandleFunc(handlerCashpointCreate(handlerContext)).Methods("POST")
	router.HandleFunc(handlerCashpointsBatch(handlerContext)).Methods("POST")
	router.HandleFunc(handlerTown(handlerContext)).Methods("GET")
	router.HandleFunc(handlerTownsBatch(handlerContext)).Methods("POST")
	router.HandleFunc(handlerTownsList(handlerContext)).Methods("GET")
	router.HandleFunc(handlerBank(handlerContext)).Methods("GET")
	router.HandleFunc(handlerBankIco(handlerContext, serverConfig)).Methods("GET")
	router.HandleFunc(handlerBanksList(handlerContext)).Methods("GET")
	router.HandleFunc(handlerBanksBatch(handlerContext)).Methods("POST")
	router.HandleFunc(handlerNearbyCashPoints(handlerContext)).Methods("POST")
	router.HandleFunc(handlerNearbyClusters(handlerContext)).Methods("POST")

	if serverConfig.TestingMode {
		router.HandleFunc(handlerCoordToQuadKey(handlerContext)).Methods("POST")
		router.HandleFunc(handlerQuadTreeBranch(handlerContext)).Methods("GET")
		router.HandleFunc(handlerCashpointDelete(handlerContext)).Methods("DELETE")
		router.HandleFunc(handlerSpaceMetrics(handlerContext)).Methods("GET")
	}

	port := strconv.FormatUint(serverConfig.Port, 10)
	log.Println("Listening port: " + port)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
