package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/tarantool/go-tarantool"
	"io"
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

func checkConvertionUint(val uint32, err error, context string) uint32 {
	if err != nil {
		log.Printf("%s: uint conversion err => %v\n", context, err)
		return 0
	}
	return val
}

func logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string/*, redisCliPool *pool.Pool*/) error {
	endpointStr := r.URL.Path
	if requestBody != "" {
		endpointStr = endpointStr + " =>"
		body := ""
		
// 		redisCli, err := redisCliPool.Get()
// 		if err != nil {
// 			log.Fatal("logRequest: cannot get redisCli from pool")
// 			return err
// 		}
// 		defer redisCliPool.Put(redisCli)

// 		if isTestingModeEnabled(redisCli) {
// 			prettyJson, err := jsonPrettify(requestBody)
// 			if err == nil {
// 				body = "\n" + prettyJson
// 			} else {
// 				body = " " + requestBody
// 			}
// 		} else {
			body = " " + requestBody
// 		}
		endpointStr = endpointStr + body
	}
	log.Printf("%s Request: %s %s", getRequestContexString(r), r.Method, endpointStr)
	return nil
}

func logResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string/*, redisCliPool *pool.Pool*/) error {
	endpointStr := r.URL.Path
	if responseBody != "" {
		endpointStr = endpointStr + " =>"
		body := ""

// 		redisCli, err := redisCliPool.Get()
// 		if err != nil {
// 			log.Fatal("logRequest: cannot get redisCli from pool")
// 			return err
// 		}
// 		defer redisCliPool.Put(redisCli)
// 
// 		if isTestingModeEnabled(redisCli) {
// 			prettyJson, err := jsonPrettify(responseBody)
// 			if err == nil {
// 				body = "\n" + prettyJson
// 			} else {
// 				body = " " + responseBody
// 			}
// 		} else {
			body = " " + responseBody
// 		}
		endpointStr = endpointStr + body
	}
	log.Printf("%s: Response: %s %s", getRequestContexString(r), r.Method, endpointStr)
	return nil
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

func prepareResponse(w http.ResponseWriter, r *http.Request) (bool, int64) {
	requestId, err := getRequestUserId(r)
	if err != nil {
		log.Printf("%s prepareResponse %v\n", getRequestContexString(r), err)
		w.WriteHeader(401)
		return false, 0
	}
	if requestId == 0 {
		log.Printf("%s prepareResponse unexpected requestId: %d\n", getRequestContexString(r), requestId)
		w.WriteHeader(401)
		return false, 0
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Id", strconv.FormatInt(requestId, 10))
	return true, requestId
}

func writeResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string/*, redisCliPool *pool.Pool*/) {
	io.WriteString(w, responseBody)
	go logResponse(w, r, requestId, responseBody/*, redisCliPool*/)
}

type EndpointCallback func(w http.ResponseWriter, r *http.Request)

func handlerPing(tnt *tarantool.Connection) (string, EndpointCallback) {
	return "/ping", func(w http.ResponseWriter, r *http.Request) {
		ok, requestId := prepareResponse(w, r)
		if ok == false {
			return
		}

		go logRequest(w, r, requestId, "")
		msg := &Message{Text: "pong"}
		jsonByteArr, _ := json.Marshal(msg)
		writeResponse(w, r, requestId, string(jsonByteArr))
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
		Reconnect: 1 * time.Second,
		MaxReconnects: 3,
		User: serverConfig.TntUser,
 		Pass: serverConfig.TntPass,
	}
	tnt, err := tarantool.Connect(serverConfig.TntUrl, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer tnt.Close()

	router := mux.NewRouter()
	router.HandleFunc(handlerPing(tnt)).Methods("GET")
	router.HandleFunc(handlerCashpoint(tnt)).Methods("GET")
	router.HandleFunc(handlerCashpointCreate(tnt)).Methods("POST")
	router.HandleFunc(handlerCashpointsBatch(tnt)).Methods("POST")
	router.HandleFunc(handlerTown(tnt)).Methods("GET")
	router.HandleFunc(handlerTownsBatch(tnt)).Methods("POST")
	router.HandleFunc(handlerTownsList(tnt)).Methods("GET")
	router.HandleFunc(handlerBank(tnt)).Methods("GET")
	router.HandleFunc(handlerBankIco(serverConfig)).Methods("GET")
	router.HandleFunc(handlerBanksList(tnt)).Methods("GET")
	router.HandleFunc(handlerBanksBatch(tnt)).Methods("POST")
	router.HandleFunc(handlerNearbyCashPoints(tnt)).Methods("POST")
	router.HandleFunc(handlerNearbyClusters(tnt)).Methods("POST")

	if serverConfig.TestingMode {
		router.HandleFunc(handlerCoordToQuadKey(tnt)).Methods("POST")
		router.HandleFunc(handlerQuadTreeBranch(tnt)).Methods("GET")
		router.HandleFunc(handlerCashpointDelete(tnt)).Methods("DELETE")
		router.HandleFunc(handlerSpaceMetrics(tnt)).Methods("GET")
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