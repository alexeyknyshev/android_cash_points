package main

import (
	"io"
	"log"
	"os"
	//    "fmt"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mediocregopher/radix.v2/redis"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"
	"unicode"
)

// ========================================================

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isAlphaNumericString(s string) bool {
	for _, c := range s {
		//        print("\\u" + strconv.FormatInt(int64(c), 16) + "\n")
		if !isAlphaNumeric(c) {
			return false
		}
	}
	return true
}

// ========================================================

const JsonNullResponse string = `{"id":null}`
const JsonLoginTooShortResponse = `{"id":null,"msg":"Login is too short"}`
const JsonLoginInvalidCharResponse = `{"id":null,"msg":"Login contains invalid characters"}`
const JsonPwdTooShortResponse = `{"id":null,"msg":"Password is too short"}`
const JsonPwdInvalidCharResponse = `{"id":null,"msg":"Password contains invalid characters"}`

// ========================================================

type ServerConfig struct {
	TownsDataBase      string `json:"TownsDataBase"`
	CashPointsDataBase string `json:"CashPointsDataBase"`
	CertificateDir     string `json:"CertificateDir"`
	Port               uint64 `json:"Port"`
	UserLoginMinLength uint64 `json:"UserLoginMinLength"`
	UserPwdMinLength   uint64 `json:"UserPwdMinLength"`
	UseTLS             bool   `json:"UseTLS"`
	RedisHost          string `json:"RedisHost"`
	ReqResLogTTL       uint64 `json:"ReqResLogTTL"`
}

func getRequestContexString(r *http.Request) string {
	return r.RemoteAddr
}

func getHandlerContextString(funcName string, requestId int64, idList ...string) string {
	result := funcName + ":" + strconv.FormatInt(requestId, 10)
	if len(idList) > 0 {
		result = result + "("
		for _, id := range idList {
			result = result + id + ","
		}
		result = result + ")"
	}
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

// ========================================================

func checkConvertionUint(val uint32, err error, context string) uint32 {
	if err != nil {
		log.Printf("%s: uint conversion err => %v\n", context, err)
		return 0
	}
	return val
}

func checkConvertionFloat(val float32, err error, context string) float32 {
	if err != nil {
		log.Printf("%s: float conversion err => %v\n", context, err)
		return 0.0
	}
	return val
}

// ========================================================

func logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string) error {
	/*
	   path := r.URL.Path
	   timeStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	   requestStr := "request:" + timeStr

	   err := redis_cli.Cmd("HMSET", requestStr,
	                                 "path", path,
	                                 "data", requestBody,
	                                 "time", timeStr,
	                                 "user_id", requestId).Err
	   if err != nil {
	       log.Printf("logRequest: %v\n", err)
	       return err
	   }

	   err = redis_cli.Cmd("EXPIRE", requestStr, REQ_RES_LOG_TTL).Err
	   if err != nil {
	       log.Printf("logRequest: %v\n", err)
	   }

	   return err
	*/
	return nil
}

func logResponse(context string, requestId int64, responseBody string) error {
	/*
	   timeStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	   responseStr := "response:" + timeStr

	   err := redis_cli.Cmd("HMSET", responseStr,
	                                 "data", responseBody,
	                                 "time", timeStr,
	                                 "user_id", requestId).Err
	   if err != nil {
	       log.Printf("logResponse: %v\n", err)
	       return err
	   }

	   err = redis_cli.Cmd("EXPIRE", responseStr, REQ_RES_LOG_TTL).Err
	   if err != nil {
	       log.Printf("logResponse: %v\n", err)
	   }

	   return err
	*/
	return nil
}

func writeResponse(w http.ResponseWriter, r *http.Request, requestId int64, responseBody string) {
	io.WriteString(w, responseBody)
	go logResponse(getRequestContexString(r), requestId, responseBody)
}

func prepareResponse(w http.ResponseWriter, r *http.Request) (bool, int64) {
	requestId, err := getRequestUserId(r)
	if err != nil {
		log.Printf("%s prepareResponse %v\n", getRequestContexString(r), err)
		writeResponse(w, r, 0, JsonNullResponse)
		return false, 0
	}
	if requestId == 0 {
		log.Printf("%s prepareResponse unexpected requestId: %d\n", getRequestContexString(r), requestId)
		writeResponse(w, r, 0, JsonNullResponse)
		return false, 0
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Id", strconv.FormatInt(requestId, 10))
	return true, requestId
}

// ========================================================

func preloadRedisScripts(redisCli *redis.Client) {
	redis_scripts = make(map[string]string)
	response := redisCli.Cmd("SCRIPT", "LOAD", "if redis.call('EXISTS', 'user:' .. ARGV[1]) == 1 then\n"+
		"    return 1\n"+
		"end\n"+
		"if redis.call('HMSET', 'user:' .. ARGV[1], 'password', ARGV[2]) == 1 then\n"+
		"    return 2\n"+
		"end\n"+
		"return 0")
	if response.Err != nil {
		log.Fatalf("preloadRedisScripts: %v\n", response.Err)
	}
	scriptSha, err := response.Str()
	if err != nil {
		log.Fatalf("preloadRedisScripts: %v\n", err)
	}
	redis_scripts[script_user_create] = scriptSha
}

// ========================================================

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Town struct {
	Id        uint32  `json:"id"`
	Name      string  `json:"name"`
	NameTr    string  `json:"name_tr"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	Zoom      uint32  `json:"zoom"`
}

type Bank struct {
	Id       uint32 `json:"id"`
	Name     string `json:"name"`
	NameTr   string `json:"name_tr"`
	RegionId uint32 `json:"region_id"`
}

type SearchNearbyRequest struct {
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
	Radius    float32 `json:"radius"`
}

type SearchNearbyResponse struct {
    CashPointIds []uint32 `json:"cash_points"`
}

type CashPoint struct {
	Id             uint32  `json:"id"`
	Type           string  `json:"type"`
	BankId         uint32  `json:"bank_id"`
	TownId         uint32  `json:"town_id"`
	Longitude      float32 `json:"longitude"`
	Latitude       float32 `json:"latitude"`
	Address        string  `json:"address"`
	AddressComment string  `json:"address_comment"`
	MetroName      string  `json:"metro_name"`
	FreeAccess     bool    `json:"free_access"`
	MainOffice     bool    `json:"main_office"`
	WithoutWeekend bool    `json:"without_weekend"`
	RoundTheClock  bool    `json:"round_the_clock"`
	WorksAsShop    bool    `json:"works_as_shop"`
	Schedule       string  `json:"schedule"`
	Tel            string  `json:"tel"`
	Additional     string  `json:"additional"`
	Rub            bool    `json:"rub"`
	Usd            bool    `json:"usd"`
	Eur            bool    `json:"eur"`
	CashIn         bool    `json:"cash_in"`
}

type CashPointIdsInTown struct {
	TownId       uint32   `json:"town_id"`
	BankId       uint32   `json:"bank_id"`
	CashPointIds []uint32 `json:"cash_points"`
}

var BuildDate string

var users_db *sql.DB
var redis_cli *redis.Client

var redis_scripts map[string]string

const script_user_create = "USERCREATE"

const SERVER_DEFAULT_CONFIG = "config.json"

var MIN_LOGIN_LENGTH uint64 = 4
var MIN_PWD_LENGTH uint64 = 4

var REQ_RES_LOG_TTL uint64 = 60

// ========================================================

func handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + getHandlerContextString("handlerUserCreate", requestId)

	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		go logRequest(w, r, requestId, "")
		log.Printf("%s => malformed json\n", context)
		w.WriteHeader(400)
		return
	}
	userJsonStr, _ := json.Marshal(user)
	go logRequest(w, r, requestId, string(userJsonStr))

	if !isAlphaNumericString(user.Login) {
		writeResponse(w, r, requestId, JsonLoginInvalidCharResponse)
		return
	}

	if !isAlphaNumericString(user.Password) {
		writeResponse(w, r, requestId, JsonPwdInvalidCharResponse)
		return
	}

	result := redis_cli.Cmd("EVALSHA", redis_scripts[script_user_create], 0,
		user.Login, user.Password)
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	retcode, err := result.Int()
	if err != nil {
		log.Printf("%s => %v: redis retcode cannot be converted to int\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if retcode == 1 {
		// user already exists
		w.WriteHeader(409)
		return
	} else if retcode == 2 {
		// redis HMSET internall err
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

// ========================================================

func handlerTown(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	townId := params["id"]

	context := getRequestContexString(r) + getHandlerContextString("handlerTown", requestId, townId)

	result := redis_cli.Cmd("GET", "town:"+townId)
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s => no such town id\n", context)
		w.WriteHeader(404)
		return
	}

	jsonStr, err := result.Str()
	if err != nil {
		log.Printf("%s => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	writeResponse(w, r, requestId, jsonStr)
}

func handlerBank(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	bankId := params["id"]

	context := getRequestContexString(r) + getHandlerContextString("handlerBank", requestId, bankId)

	result := redis_cli.Cmd("GET", "bank:"+bankId)
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s => no such bank id\n", context)
		w.WriteHeader(404)
		return
	}

	jsonStr, err := result.Str()
	if err != nil {
		log.Printf("%s => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	writeResponse(w, r, requestId, jsonStr)
}

func handlerBankCreate(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " handlerBankCreate"

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}

	go logRequest(w, r, requestId, jsonStr)
}

func handlerCashpoint(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	cashPointId := params["id"]

	context := getRequestContexString(r) + getHandlerContextString("handlerCashpoint", requestId, cashPointId)

	result := redis_cli.Cmd("GET", "cp:"+cashPointId)
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s => no such cashpoint id\n", context)
		w.WriteHeader(404)
		return
	}

	jsonStr, err := result.Str()
	if err != nil {
		log.Printf("%s => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	writeResponse(w, r, requestId, jsonStr)
}

func handlerCashpointCreate(w http.ResponseWriter, r *http.Request) {
}

func handlerCashpointsByTownAndBank(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	townIdStr := params["town_id"]
	bankIdStr := params["bank_id"]

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointsByTownAndBank", requestId, townIdStr, bankIdStr)

	result := redis_cli.Cmd("SINTER", "town:"+townIdStr+":cp", "bank:"+bankIdStr+":cp")
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, err := result.List()
	if err != nil {
		log.Printf("%s => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	id, err := strconv.ParseUint(townIdStr, 10, 32)
	townId := checkConvertionUint(uint32(id), err, context+" => CashPointIds.TownId")

	id, err = strconv.ParseUint(bankIdStr, 10, 32)
	bankId := checkConvertionUint(uint32(id), err, context+" => CashPointIds.BankId")

	ids := CashPointIdsInTown{TownId: townId, BankId: bankId}
	if len(data) == 0 {
		ids.CashPointIds = make([]uint32, 0)
	}

	for i, idStr := range data {
		id, err := strconv.ParseUint(idStr, 10, 32)
		id32 := checkConvertionUint(uint32(id), err, context+" => CashPointIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
		ids.CashPointIds = append(ids.CashPointIds, id32)
	}

	jsonByteArr, _ := json.Marshal(ids)
	jsonStr := string(jsonByteArr)
	writeResponse(w, r, requestId, jsonStr)
}

func handlerSearchCashPoinstsNearby(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerSearchCashPoinstsNearby", requestId)

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}

	go logRequest(w, r, requestId, jsonStr)

    result := redis_cli.Cmd("CPSEARCHNEARBY", jsonStr)
    if result.Err != nil {
        log.Printf("%s => %v\n", context, result.Err)
        w.WriteHeader(500)
        return
    }

    data, err := result.List()
    if err != nil {
        log.Printf("%s => %v\n", context, err)
        w.WriteHeader(500)
        return
    }

    res := SearchNearbyResponse{CashPointIds: make([]uint32, 0)}

    for i, idStr := range data {
        id, err := strconv.ParseUint(idStr, 10, 32)
        id32 := checkConvertionUint(uint32(id), err, context+" => CashPointIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
        res.CashPointIds = append(res.CashPointIds, id32)
    }

    jsonByteArr, _ := json.Marshal(res)
    writeResponse(w, r, requestId, string(jsonByteArr))
}

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
	log.Println("CashPoints server build: " + BuildDate)

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
	}

	certPath := ""
	pkeyPath := ""

	if serverConfig.UseTLS {
		certPath = path.Join(serverConfig.CertificateDir, "cert.pem")
		pkeyPath = path.Join(serverConfig.CertificateDir, "key.pem")

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			log.Fatalf("No such cert file for tls: %s\n", certPath)
		}

		if _, err := os.Stat(pkeyPath); os.IsNotExist(err) {
			log.Fatalf("No such private key file for tls: %s\n", pkeyPath)
		}
	}

	redis_cli, err = redis.Dial("tcp", serverConfig.RedisHost)
	if err != nil {
		log.Fatal(err)
	}
	defer redis_cli.Close()

	redis_cli.Cmd("HMSET", "settings", "user_login_min_length", serverConfig.UserLoginMinLength,
		"user_pwd_min_length", serverConfig.UserPwdMinLength)

	preloadRedisScripts(redis_cli)

	REQ_RES_LOG_TTL = serverConfig.ReqResLogTTL

	router := mux.NewRouter()
	router.HandleFunc("/user", handlerUserCreate).Methods("POST")
	router.HandleFunc("/town/{id:[0-9]+}", handlerTown)
	router.HandleFunc("/bank/{id:[0-9]+}", handlerBank)
	router.HandleFunc("/bank", handlerBankCreate).Methods("POST")
	router.HandleFunc("/cashpoint", handlerCashpointCreate).Methods("POST")
	router.HandleFunc("/cashpoint/{id:[0-9]+}", handlerCashpoint)
	router.HandleFunc("/town/{town_id:[0-9]+}/bank/{bank_id:[0-9]+}/cashpoints", handlerCashpointsByTownAndBank)
	router.HandleFunc("/search/caspoints/nearby", handlerSearchCashPoinstsNearby).Methods("POST")

	port := ":" + strconv.FormatUint(serverConfig.Port, 10)
	log.Println("Listening 127.0.0.1" + port)

	server := &http.Server{
		Addr:           port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if serverConfig.UseTLS {
		log.Println("Using TLS encryption")
		log.Println("Certificate path: " + certPath)
		log.Println("Private key path: " + pkeyPath)
		err = server.ListenAndServeTLS(certPath, pkeyPath)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Fatal(err)
	}
}
