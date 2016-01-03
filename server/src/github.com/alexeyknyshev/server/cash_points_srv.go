package main

import (
	"io"
	"log"
	"os"
	//"fmt"
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	//"github.com/fiam/gounidecode/unidecode"
	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"math"
	"net/http"
	"path"
	"path/filepath"
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
		//print("\\u" + strconv.FormatInt(int64(c), 16) + "\n")
		if !isAlphaNumeric(c) {
			return false
		}
	}
	return true
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func removeDuplicates(arr []string) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, v := range arr {
		if _, ok := seen[v]; !ok {
			result = append(result, v)
			seen[v] = struct{}{}
		}
	}
	return result
}

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
	RedisScriptsDir    string `json:"RedisScriptsDir"`
	ReqResLogTTL       uint64 `json:"ReqResLogTTL"`
	UUID_TTL           uint64 `json:"UUID_TTL"`
	BanksIcoDir        string `json:"BanksIcoDir"`
	TestingMode        bool   `json:"TestingMode"`
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
	path := r.URL.Path
	/*
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
	endpointStr := path
	if requestBody != "" {
		endpointStr = endpointStr + " => " + requestBody
	}
	log.Printf("%s Request: %s %s", getRequestContexString(r), r.Method, endpointStr)
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

func redisListResponseExpected(w http.ResponseWriter, r *redis.Resp, context string) ([]string, bool) {
	if r.IsType(redis.Str) {
		errStr, _ := r.Str()
		log.Printf("%s: redis => %s\n", context, errStr)
		w.WriteHeader(500)
		data := make([]string, 0)
		return data, false
	}

	data, err := r.List()
	if err != nil {
		log.Printf("%s => %v\n", context, err)
		w.WriteHeader(500)
		data := make([]string, 0)
		return data, false
	}

	return data, true
}

// ========================================================

func preloadRedisScriptSrc(redisCli *redis.Client, srcFilePath string) string {
	context := "preloadRedisScripts: " + srcFilePath

	buf := bytes.NewBuffer(nil)
	file, err := os.Open(srcFilePath)
	if err != nil {
		log.Fatalf("%s => %v\n", context, err)
	}
	io.Copy(buf, file)
	file.Close()
	src := string(buf.Bytes())

	response := redisCli.Cmd("SCRIPT", "LOAD", src)
	if response.Err != nil {
		log.Fatalf("%s => %v\n", context, response.Err)
	}
	scriptSha, err := response.Str()
	if err != nil {
		log.Fatalf("%s => %v\n", context, err)
	}
	return scriptSha
}

func preloadRedisScripts(redisCli *redis.Client, scriptsDir string) {
	redis_scripts = make(map[string]string)

	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		log.Fatalf("preloadRedisScripts: No such directory file: %s\n", scriptsDir)
	}

	log.Printf("Flushing redis script cache")
	redisCli.Cmd("SCRIPT", "FLUSH")

	filepath.Walk(scriptsDir, func(path string, fi os.FileInfo, _ error) error {
		if fi.IsDir() == false {
			fileBaseName := fi.Name()
			fileExt := filepath.Ext(fileBaseName)
			if strings.ToLower(fileExt) == ".lua" {
				logStr := "Loading redis script: " + fileBaseName
				defer func() {
					log.Printf(logStr)
				}()
				cmdName := strings.ToUpper(strings.TrimSuffix(fileBaseName, fileExt))
				scriptSha := preloadRedisScriptSrc(redisCli, path)
				redis_scripts[cmdName] = scriptSha
				logStr = logStr + " => " + cmdName + "(" + scriptSha + ")"
			}
		}
		return nil
	})

	return
}

func dropTestData(redisCli *redis.Client) {
	result := redisCli.Cmd("KEYS", "test_*")
	if result.Err != nil {
		log.Fatalf("Failed to drop test data. Cannot get 'test_*' result: redis => %v", result.Err)
		return
	}

	data, err := result.List()
	if err != nil {
		log.Fatalf("Failed to drop test data. Cannot 'test_*' keys to list: redis => %v", result.Err)
		return
	}

	for _, idStr := range data {
		log.Printf("Removing test data redis key: %s", idStr)
		result = redisCli.Cmd("DEL", idStr)
		if result.Err != nil {
			log.Fatalf("Failed to drop test data key '%s': redis => %v", idStr, result.Err)
			return
		}
	}
}

// ========================================================

type Message struct {
	Text string `json:"text"`
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Session struct {
	Key string `json:"key"`
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

type BankCreateRequest struct {
	Name     string `json:"name"`
	Licence  uint32 `json:"licence"`
	RegionId uint32 `json:"region_id"`
	Tel      string `json:"tel"`
}

type SearchNearbyRequest struct {
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
	Radius    float32 `json:"radius"`
	Zoom      uint32  `json:"zoom"`
	Filter    string  `json:",string"`
	//Radius    float32 `json:"radius"`
}

type SearchNearbyRequestInternal struct {
	QuadKeys []string `json:"quadkeys"`
	Filter   string   `json:"filter"`
}

func (req *SearchNearbyRequest) Validate() error {
	err := isGeoCoordValid(req.Longitude, req.Latitude)
	if err != nil {
		return err
	}

	if req.Radius <= 0.0 {
		return errors.New("radius must be positive")
	}

	if req.Zoom > MAX_VALID_ZOOM {
		return errors.New("zoom is out of range")
	}
	return nil
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
	Timestamp      int64   `json:"timestamp"`
	Version        uint32  `json:"version"`
}

func (cp *CashPoint) Validate() error {
	if cp.TownId == 0 {
		return errors.New("No required field 'town_id'")
	}

	if cp.BankId == 0 {
		return errors.New("No required field 'bank_id'")
	}

	if cp.Longitude == 0 {
		return errors.New("No required field 'longitude'")
	}

	if cp.Latitude == 0 {
		return errors.New("No required field 'latitude'")
	}

	return nil
}

type CashPointIdsInTown struct {
	TownId       uint32   `json:"town_id"`
	BankId       uint32   `json:"bank_id"`
	CashPointIds []uint32 `json:"cash_points"`
}

type CashPointIds struct {
	CashPointIds []uint32 `json:"cash_points"`
}

type TownIds struct {
	TownIds []uint32 `json:"towns"`
}

type TownList struct {
	TownList []map[string]*json.RawMessage `json:"towns"`
}

type RegionList struct {
	RegionList []map[string]*json.RawMessage `json:"regions"`
}

type BankIds struct {
	BankIds []uint32 `json:"banks"`
}

type BankList struct {
	BankList []map[string]*json.RawMessage `json:"banks"`
}

type BankIco struct {
	BankId  uint32 `json:"bank_id"`
	IcoData string `json:"ico_data"`
}

//var BuildDate string

var redis_cli_pool *pool.Pool

var redis_scripts map[string]string

const script_user_create = "USERCREATE"
const script_user_login = "USERLOGIN"
const script_bank_create = "BANKCREATE"
const script_search_nearby = "SEARCHNEARBY"
const script_towns_batch = "TOWNSBATCH"
const script_regions_batch = "REGIONSBATCH"
const script_banks_batch = "BANKSBATCH"
const script_cashpoints_batch = "CASHPOINTSBATCH"
const script_cashpoints_history = "CASHPOINTSHISTORY"
const script_cluster_data_batch = "CLUSTERDATABATCH"
const script_cluster_data_town = "CLUSTERDATATOWN"
const script_cluster_data = "CLUSTERDATA"

const SERVER_DEFAULT_CONFIG = "config.json"

const UUID_TTL_MIN = 10
const UUID_TTL_MAX = 1000

const MAX_VALID_ZOOM = 16
const MIN_CLUSTER_ZOOM = 10

var MIN_LOGIN_LENGTH uint64 = 4
var MIN_PWD_LENGTH uint64 = 4

var REQ_RES_LOG_TTL uint64 = 60
var MAX_CLUSTER_COUNT uint64 = 32

// ========================================================

func handlerPing(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	go logRequest(w, r, requestId, "")
	msg := &Message{Text: "pong"}
	jsonByteArr, _ := json.Marshal(msg)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

// ========================================================

func handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerUserCreate", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}

	go logRequest(w, r, requestId, string(jsonStr))

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_user_create], 0, jsonStr)
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	ret, err := result.Str()
	if err != nil {
		log.Printf("%s => %v: redis '%s' result cannot be converted to string\n", context, result.Err, script_user_create)
		w.WriteHeader(500)
		return
	}

	if strings.HasPrefix(ret, "User with already exists") {
		// user already exists
		w.WriteHeader(409)
		return
	} else if strings.HasPrefix(ret, "User login") {
		w.WriteHeader(400)
		return
	} else if strings.HasPrefix(ret, "User password") {
		w.WriteHeader(400)
		return
	} else if ret != "" {
		// redis script internal err
		log.Printf("%s => %s\n", context, ret)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

func handlerUserDelete(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerUserDelete", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}

	log.Printf("%s", jsonStr)
}

func handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerUserLogin", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}

	newUuid, err := uuid.NewV4()
	if err != nil {
		log.Printf("%s => %v\n", context, err)
		return
	}

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	uuidStr := newUuid.String()
	result := redisCli.Cmd("EVALSHA", redis_scripts[script_user_login], 0, jsonStr, uuidStr)
	if result.Err != nil {
		log.Printf("%s => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	ret, err := result.Str()
	if err != nil {
		log.Printf("%s => %v: redis '%s' result cannot be converted to string\n", context, result.Err, script_user_login)
		w.WriteHeader(500)
		return
	}

	code := 500
	switch ret {
	case "":
		sess := Session{Key: uuidStr}
		jsonByteArr, _ := json.Marshal(sess)
		writeResponse(w, r, requestId, string(jsonByteArr))
		return
	case "Invalid password":
		code = 417
	case "No such user account":
		code = 417
	default:
		log.Printf("%s: redis => %s", context, ret)
	}

	w.WriteHeader(code)
}

// ========================================================

func handlerTownList(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownList", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s: => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("ZRANGE", "towns", 0, -1)
	if result.Err != nil {
		log.Fatal("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, err := result.List()
	if err != nil {
		log.Printf("%s: redis => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	res := new(TownIds)
	if len(data) == 0 {
		res.TownIds = make([]uint32, 0)
	}

	for i, idStr := range data {
		id, err := strconv.ParseUint(idStr, 10, 32)
		id32 := checkConvertionUint(uint32(id), err, context+" => TownIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
		res.TownIds = append(res.TownIds, id32)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerTownsBatch(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerTownList", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}
	go logRequest(w, r, requestId, jsonStr)

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_towns_batch], 0, jsonStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, ok := redisListResponseExpected(w, result, context)
	if ok == false {
		return
	}

	res := new(TownList)
	if len(data) == 0 {
		res.TownList = make([]map[string]*json.RawMessage, 0)
	}

	for _, townJson := range data {
		var town map[string]*json.RawMessage
		//log.Printf("%s => %s\n", context, townJson)
		json.Unmarshal([]byte(townJson), &town)
		res.TownList = append(res.TownList, town)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerTown(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	townId := params["id"]

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerTown", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
		"townId":    townId,
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("GET", "town:"+townId)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s => no such townId=%s\n", context, townId)
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

func handlerRegions(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerTown", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_regions_batch], 0)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, ok := redisListResponseExpected(w, result, context)
	if ok == false {
		return
	}

	res := new(RegionList)
	if len(data) == 0 {
		res.RegionList = make([]map[string]*json.RawMessage, 0)
	}

	for _, regionJson := range data {
		var region map[string]*json.RawMessage
		json.Unmarshal([]byte(regionJson), &region)
		res.RegionList = append(res.RegionList, region)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

// ========================================================

func handlerBankList(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerBankList", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("SMEMBERS", "banks")

	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, ok := redisListResponseExpected(w, result, context)
	if ok == false {
		return
	}

	res := new(BankIds)
	if len(data) == 0 {
		res.BankIds = make([]uint32, 0)
	}

	for i, idStr := range data {
		id, err := strconv.ParseUint(idStr, 10, 32)
		id32 := checkConvertionUint(uint32(id), err, context+" => BankIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
		res.BankIds = append(res.BankIds, id32)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerBanksBatch(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerBanksBatch", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}
	go logRequest(w, r, requestId, jsonStr)

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_banks_batch], 0, jsonStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Str) {
		errStr, _ := result.Str()
		log.Printf("%s: redis => %s\n", context, errStr)
		w.WriteHeader(500)
		return
	}

	data, ok := redisListResponseExpected(w, result, context)
	if ok == false {
		return
	}

	res := new(BankList)
	if len(data) == 0 {
		res.BankList = make([]map[string]*json.RawMessage, 0)
	}

	for _, bankJson := range data {
		var town map[string]*json.RawMessage
		json.Unmarshal([]byte(bankJson), &town)
		res.BankList = append(res.BankList, town)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerBank(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	bankId := params["id"]

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerBank", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
		"bankId":    bankId,
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("GET", "bank:"+bankId)

	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s => no such bankId=%s\n", context, bankId)
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

func handlerBankIco(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	params := mux.Vars(r)
	bankIdStr := params["id"]

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerBankIco", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
		"bankId":    bankIdStr,
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("HGET", "settings", "banks_ico_dir")

	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s: redis => no such settings entry: %s\n", context, "banks_ico_dir")
		w.WriteHeader(500)
		return
	}

	banksIcoDir, err := result.Str()
	if err != nil {
		log.Printf("%s: redis => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	icoFilePath := path.Join(banksIcoDir, bankIdStr+".svg")

	if _, err := os.Stat(icoFilePath); os.IsNotExist(err) {
		w.WriteHeader(404)
		return
	}

	data, err := ioutil.ReadFile(icoFilePath)
	if err != nil {
		log.Printf("%s => cannot read file: %s", context, icoFilePath)
		w.WriteHeader(500)
		return
	}

	id, err := strconv.ParseUint(bankIdStr, 10, 32)
	bankId := checkConvertionUint(uint32(id), err, context+" => BankIco.BankId")

	ico := &BankIco{BankId: bankId, IcoData: string(data)}
	jsonByteArr, _ := json.Marshal(ico)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerBankCreate(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerBankCreate", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

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

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpoint", map[string]string{
		"requestId":   strconv.FormatInt(requestId, 10),
		"cashPointId": cashPointId,
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("GET", "cp:"+cashPointId)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s => no such cashPointId=%s\n", context, cashPointId)
		w.WriteHeader(404)
		return
	}

	jsonStr, err := result.Str()
	if err != nil {
		log.Printf("%s: redis => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	writeResponse(w, r, requestId, jsonStr)
}

func handlerCashpointsBatch(w http.ResponseWriter, r *http.Request) {
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

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_cashpoints_batch], 0, jsonStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Str) == false {
		log.Printf("%s: redis => script result type is not string\n", context)
		w.WriteHeader(500)
		return
	}

	jsonRes, _ := result.Str()
	if strings.HasPrefix(jsonRes, "{") == false {
		log.Printf("%s: redis => %s\n", context, jsonRes)
		w.WriteHeader(500)
		return
	}

	writeResponse(w, r, requestId, jsonRes)
}

func handlerCashpointsHistory(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointsHistory", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}
	go logRequest(w, r, requestId, jsonStr)

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_cashpoints_history], 0, jsonStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Str) {
		errStr, _ := result.Str()
		log.Printf("%s: redis => %s\n", context, errStr)
		w.WriteHeader(500)
		return
	}

	data, err := result.List()
	if err != nil {
		log.Printf("%s: redis => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	ids := &CashPointIds{CashPointIds: make([]uint32, 0)}

	for i, idStr := range data {
		id, err := strconv.ParseUint(idStr, 10, 32)
		id32 := checkConvertionUint(uint32(id), err, context+" => CashPointIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
		ids.CashPointIds = append(ids.CashPointIds, id32)
	}

	jsonByteArr, _ := json.Marshal(ids)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func getNextCashpointId(redisCli *redis.Client) (uint32, error) {
	context := "getNextCashpointId"

	testingMode := false
	result := redisCli.Cmd("HGET", "settings", "testing_mode")
	if result.Err == nil {
		res, err := result.Int()
		if err == nil && res == 1 {
			testingMode = true
		}
	}

	nextIdKey := ""

	if testingMode {
		nextIdKey = "test_cp_next_id"
	} else {
		nextIdKey = "cp_next_id"
	}

	result = redisCli.Cmd("INCR", nextIdKey)

	if result.Err != nil {
		return 0, result.Err
	}

	res, err := result.Int()
	if err != nil {
		return 0, err
	}

	log.Printf("%s: generated new cashpoint id: %d", context, res)

	return uint32(res), nil
}

func getGeoRectPart(minLon, maxLon, minLat, maxLat *float32, lon, lat float32) string {
	midLon := (*minLon + *maxLon) * 0.5
	midLat := (*minLat + *maxLat) * 0.5

	if lat < midLat {
		*maxLat = midLat
		if lon < midLon {
			*maxLon = midLon
			return "0"
		} else {
			*minLon = midLon
			return "1"
		}
	} else {
		*minLat = midLat
		if lon < midLon {
			*maxLon = midLon
			return "2"
		} else {
			*minLon = midLon
			return "3"
		}
	}
}

func isGeoCoordValid(lon, lat float32) error {
	if math.Abs(float64(lon)) > 180.0 {
		return errors.New("longitude is out of range")
	} else if math.Abs(float64(lat)) > 85.0 {
		return errors.New("latitude is out of range")
	}
	return nil
}

func getQuadKey(lon, lat float32, maxZoom uint32) string {
	var minLon float32 = -180.0
	var maxLon float32 = 180.0

	var minLat float32 = -85.0
	var maxLat float32 = 85.0

	quadKey := ""
	for zoom := uint32(0); zoom < maxZoom; zoom++ {
		quadKey += getGeoRectPart(&minLon, &maxLon, &minLat, &maxLat, lon, lat)
	}
	return quadKey
}

func handlerCashpointCreate(w http.ResponseWriter, r *http.Request) {
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

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	cpData := CashPoint{
		Id:        0,
		TownId:    0,
		BankId:    0,
		Longitude: 0.0,
		Latitude:  0.0,
		Version:   1,
	}

	err = json.Unmarshal([]byte(jsonStr), &cpData)
	if err != nil {
		log.Printf("%s: failed to unpack json: %v", context, err)
		log.Printf("%s: %s", context, jsonStr)
		w.WriteHeader(500)
		return
	}

	err = cpData.Validate()
	if err != nil {
		log.Printf("%s: invalid cashpoint data: %v", context, err)
		w.WriteHeader(500)
		return
	}

	bankIdStr := strconv.FormatUint(uint64(cpData.BankId), 10)
	result := redisCli.Cmd("EXISTS", "bank:"+bankIdStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	exists, err := result.Int()
	if err != nil {
		log.Printf("%s: redis => %v", context, err)
		w.WriteHeader(500)
		return
	}

	if exists == 0 {
		log.Printf("%s: invalid bank_id: %s", context, bankIdStr)
		w.WriteHeader(500)
		return
	}

	townIdStr := strconv.FormatUint(uint64(cpData.TownId), 10)
	result = redisCli.Cmd("EXISTS", "town:"+townIdStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	exists, err = result.Int()
	if err != nil {
		log.Printf("%s: redis => %v", context, err)
		w.WriteHeader(500)
		return
	}

	if exists == 0 {
		log.Printf("%s: invlaid town_id: %d", context, townIdStr)
		w.WriteHeader(500)
		return
	}

	cpData.Id, err = getNextCashpointId(redisCli)
	if err != nil {
		log.Printf("%s: getNextCashpointId: redis => %v", context, err)
		w.WriteHeader(500)
		return
	}

	cpData.Timestamp = time.Now().Unix()

	idStr := strconv.FormatUint(uint64(cpData.Id), 10)

	jsonCpData, err := json.Marshal(cpData)
	if err != nil {
		log.Printf("%s: cannot pack new cp data: input = %s", jsonStr)
		w.WriteHeader(500)
		return
	}

	quadKey := getQuadKey(cpData.Longitude, cpData.Latitude, MAX_VALID_ZOOM)
	for i := 0; i < len(quadKey); i++ {
		clusterName := "cluster:" + quadKey[:(i+1)]
		result = redisCli.Cmd("SADD", clusterName, cpData.Id)
		if result.Err != nil {
			log.Printf("%s: cannot add new cashpoint to cluster: redis => %v\n", context, result.Err)
			w.WriteHeader(500)
			return
		}
	}

	redisCli.Cmd("SET", "cp:"+idStr, string(jsonCpData))
	redisCli.Cmd("ZADD", "cp:history", cpData.Timestamp, cpData.Id)
	redisCli.Cmd("GEOADD", "cashpoints", cpData.Longitude, cpData.Latitude, cpData.Id)
	redisCli.Cmd("SADD", "bank:"+bankIdStr+":cp", cpData.Id)

	log.Printf("%s: created cashpoint with id: %s", context, idStr)

	res := &CashPointIds{CashPointIds: make([]uint32, 0)}
	res.CashPointIds = append(res.CashPointIds, cpData.Id)
	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerCashpointDelete(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}
	go logRequest(w, r, requestId, "")

	params := mux.Vars(r)
	idStr := params["id"]

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointDelete", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
		"id":        idStr,
	})

	id, err := strconv.ParseUint(idStr, 10, 32)
	cpId := checkConvertionUint(uint32(id), err, context+" => id")
	if cpId == 0 {
		w.WriteHeader(500)
		return
	}

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("GET", "cp:"+idStr)
	if result.Err != nil {
		log.Printf("%s: cannot get cashpoint by id = %s: redis => %v\n", context, idStr, result.Err)
		w.WriteHeader(500)
		return
	}

	if result.IsType(redis.Nil) {
		log.Printf("%s: cannot delete cashpoint by id = %s: no such id\n", context, idStr)
		w.WriteHeader(404)
		return
	}

	cpData, err := result.Str()
	if err != nil {
		log.Printf("%s: cannot convert cashpoint data to string for id = %s: redis => %v\n", context, idStr, err)
		w.WriteHeader(500)
		return
	}

	cp := CashPoint{Id: 0}
	json.Unmarshal([]byte(cpData), &cp)
	if cp.Id == 0 {
		log.Printf("%s: cannot parse cashpoint json data for id = %s", context, idStr)
		w.WriteHeader(500)
		return
	}

	townCp := "town:" + strconv.FormatUint(uint64(cp.TownId), 10) + ":cp"
	result = redisCli.Cmd("SREM", townCp, cp.Id)
	if result.Err != nil {
		log.Printf("%s: cannot remove cashpoint id = %s from town cp set = %s", context, idStr, townCp)
		w.WriteHeader(500)
		return
	}

	bankCp := "bank:" + strconv.FormatUint(uint64(cp.BankId), 10) + ":cp"
	result = redisCli.Cmd("SREM", bankCp, cp.Id)
	if result.Err != nil {
		log.Printf("%s: cannot remove cashpoint id = %s from bank cp set = %s", context, idStr, bankCp)
		w.WriteHeader(500)
		return
	}

	result = redisCli.Cmd("DEL", "cp:"+idStr)
	if result.Err != nil {
		log.Printf("%s: cannot remove cashpoint id = %s", context, idStr)
		w.WriteHeader(500)
		return
	}

	result = redisCli.Cmd("ZREM", "cp:history", cp.Id)
	if result.Err != nil {
		log.Printf("%s: cannot remove cashpoint id = %s from history", context, idStr)
		w.WriteHeader(500)
		return
	}

	geoSet := "cashpoints"
	result = redisCli.Cmd("ZREM", geoSet, cp.Id)
	if result.Err != nil {
		log.Printf("%s: cannot remove cashpoint id = %s from geo set = %s", context, idStr, geoSet)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
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

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointsByTownAndBank", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
		"townId":    townIdStr,
		"bankId":    bankIdStr,
	})

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("SINTER", "town:"+townIdStr+":cp", "bank:"+bankIdStr+":cp")
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, err := result.List()
	if err != nil {
		log.Printf("%s: redis => %v\n", context, err)
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
	writeResponse(w, r, requestId, string(jsonByteArr))
}

// ========================================================

func handlerNearbyCashPoints(w http.ResponseWriter, r *http.Request) {
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

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_search_nearby], 1, "cashpoints", jsonStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, ok := redisListResponseExpected(w, result, context)
	if ok == false {
		return
	}

	res := CashPointIds{CashPointIds: make([]uint32, 0)}

	for i, idStr := range data {
		id, err := strconv.ParseUint(idStr, 10, 32)
		id32 := checkConvertionUint(uint32(id), err, context+" => CashPointIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
		res.CashPointIds = append(res.CashPointIds, id32)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerNearbyTowns(w http.ResponseWriter, r *http.Request) {
	ok, requestId := prepareResponse(w, r)
	if ok == false {
		return
	}

	context := getRequestContexString(r) + " " + getHandlerContextString("handlerNearbyTowns", map[string]string{
		"requestId": strconv.FormatInt(requestId, 10),
	})

	jsonStr, err := getRequestJsonStr(r, context)
	if err != nil {
		go logRequest(w, r, requestId, "")
		w.WriteHeader(400)
		return
	}

	go logRequest(w, r, requestId, jsonStr)

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	result := redisCli.Cmd("EVALSHA", redis_scripts[script_search_nearby], 1, "towns", jsonStr)
	if result.Err != nil {
		log.Printf("%s: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	data, ok := redisListResponseExpected(w, result, context)
	if ok == false {
		return
	}

	res := TownIds{TownIds: make([]uint32, 0)}

	for i, idStr := range data {
		id, err := strconv.ParseUint(idStr, 10, 32)
		id32 := checkConvertionUint(uint32(id), err, context+" => TownIds["+strconv.FormatInt(int64(i), 10)+"] = "+idStr)
		res.TownIds = append(res.TownIds, id32)
	}

	jsonByteArr, _ := json.Marshal(res)
	writeResponse(w, r, requestId, string(jsonByteArr))
}

func handlerNearbyClusters(w http.ResponseWriter, r *http.Request) {
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

	redisCli, err := redis_cli_pool.Get()
	if err != nil {
		log.Fatal("%s => %v\n", context, err)
		return
	}
	defer redis_cli_pool.Put(redisCli)

	request := SearchNearbyRequest{}
	err = json.Unmarshal([]byte(jsonStr), &request)
	if err != nil {
		log.Printf("%s: cannot unpack json request: %v\n", context, err)
		w.WriteHeader(400)
		return
	}

	err = request.Validate()
	if err != nil {
		log.Printf("%s: invalid request data: %v", context, err)
		w.WriteHeader(400)
		return
	}

	start := time.Now()
	var result *redis.Resp
	if request.Zoom < MIN_CLUSTER_ZOOM {
		result = redisCli.Cmd("EVALSHA", redis_scripts[script_cluster_data_town], 0, jsonStr, MAX_CLUSTER_COUNT)
	} else {
		result = redisCli.Cmd("EVALSHA", redis_scripts[script_cluster_data], 0, jsonStr)
	}
	elapsed := time.Since(start)
	log.Printf("%s: cluster lua time: %v", context, elapsed)

	if result.Err != nil {
		log.Printf("%s: cannot get cluster data: redis => %v\n", context, result.Err)
		w.WriteHeader(500)
		return
	}

	jsonStr, err = result.Str()
	if err != nil {
		log.Printf("%s: cannot convert cluster data to string: redis => %v\n", context, err)
		w.WriteHeader(500)
		return
	}

	writeResponse(w, r, requestId, jsonStr)
}

// ========================================================

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
	//log.Println("CashPoints server build: " + BuildDate)

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

	redis_cli_pool, err = pool.New("tcp", serverConfig.RedisHost, 16)
	if err != nil {
		log.Fatal(err)
	}
	redis_cli, err := redis_cli_pool.Get()
	defer redis_cli_pool.Put(redis_cli)

	if serverConfig.UUID_TTL < UUID_TTL_MIN {
		serverConfig.UUID_TTL = UUID_TTL_MIN
	} else if serverConfig.UUID_TTL > UUID_TTL_MAX {
		serverConfig.UUID_TTL = UUID_TTL_MAX
	}

	redis_cli.Cmd("HMSET", "settings",
		"user_login_min_length", serverConfig.UserLoginMinLength,
		"user_password_min_length", serverConfig.UserPwdMinLength,
		"uuid_ttl", serverConfig.UUID_TTL,
		"banks_ico_dir", serverConfig.BanksIcoDir,
		"testing_mode", boolToInt(serverConfig.TestingMode))

	preloadRedisScripts(redis_cli, serverConfig.RedisScriptsDir)
	dropTestData(redis_cli)

	REQ_RES_LOG_TTL = serverConfig.ReqResLogTTL

	router := mux.NewRouter()
	router.HandleFunc("/ping", handlerPing).Methods("GET")
	router.HandleFunc("/user", handlerUserCreate).Methods("POST")
	router.HandleFunc("/user", handlerUserDelete).Methods("DELETE")
	router.HandleFunc("/login", handlerUserLogin).Methods("POST")
	router.HandleFunc("/towns", handlerTownList).Methods("GET")
	router.HandleFunc("/towns", handlerTownsBatch).Methods("POST")
	router.HandleFunc("/regions", handlerRegions)
	router.HandleFunc("/town/{id:[0-9]+}", handlerTown)
	router.HandleFunc("/bank/{id:[0-9]+}", handlerBank)
	router.HandleFunc("/bank/{id:[0-9]+}/ico", handlerBankIco).Methods("GET")
	router.HandleFunc("/bank", handlerBankCreate).Methods("POST")
	router.HandleFunc("/banks", handlerBankList).Methods("GET")
	router.HandleFunc("/banks", handlerBanksBatch).Methods("POST")
	router.HandleFunc("/cashpoint", handlerCashpointCreate).Methods("POST")
	router.HandleFunc("/cashpoint/{id:[0-9]+}", handlerCashpoint).Methods("GET")
	router.HandleFunc("/cashpoints", handlerCashpointsBatch).Methods("POST")
	router.HandleFunc("/cashpoints/history", handlerCashpointsHistory).Methods("POST")
	router.HandleFunc("/town/{town_id:[0-9]+}/bank/{bank_id:[0-9]+}/cashpoints", handlerCashpointsByTownAndBank)
	router.HandleFunc("/nearby/cashpoints", handlerNearbyCashPoints).Methods("POST")
	router.HandleFunc("/nearby/towns", handlerNearbyTowns).Methods("POST")
	router.HandleFunc("/nearby/clusters", handlerNearbyClusters).Methods("POST")

	if serverConfig.TestingMode {
		router.HandleFunc("/cashpoint/{id:[0-9]+}", handlerCashpointDelete).Methods("DELETE")
	}

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
