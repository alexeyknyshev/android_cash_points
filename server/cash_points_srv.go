package main

import (
    "os"
    "io"
    "log"
    "fmt"
    "path"
    "time"
    "errors"
    "strconv"
    "unicode"
    "net/http"
    "database/sql"
    "encoding/json"
    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
    "github.com/mediocregopher/radix.v2/redis"
)

// ========================================================

func uintToBool(val uint32) bool {
    if (val > 0) {
        return true
    }
    return false
}

func isAlphaNumeric(s string) bool {
    for _, c := range s {
        if !unicode.IsLetter(c) || !unicode.IsNumber(c) || c != '_' {
            return false
        }
    }
    return true
}

// ========================================================

const JsonNullResponse string      = `{"id":null}`
const JsonLoginTooShortResponse    = `{"id":null,"msg":"Login is too short"}`
const JsonLoginInvalidCharResponse = `{"id":null,"msg":"Login contains invalid characters"}`
const JsonPwdTooShortResponse      = `{"id":null,"msg":"Password is too short"}`
const JsonPwdInvalidCharResponse   = `{"id":null,"msg":"Password contains invalid characters"}`

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
    TownId        uint32   `json:"town_id"`
    BankId        uint32   `json:"bank_id"`
    CashPointIds  []uint32 `json:"cash_points"`
}

var BuildDate string

var users_db *sql.DB
var redis_cli *redis.Client

var MIN_LOGIN_LENGTH uint64 = 4
var MIN_PWD_LENGTH uint64 = 4

var REQ_RES_LOG_TTL uint64 = 60

// ========================================================

func handlerUserCreate(w http.ResponseWriter, r *http.Request) {
    ok, requestId := prepareResponse(w, r)
    if ok == false {
        return
    }

    decoder := json.NewDecoder(r.Body)
    var user User
    err := decoder.Decode(&user)
    if err != nil {
        go logRequest(w, r, requestId, "")
        log.Println("Malformed User json")
        w.WriteHeader(400)
        return
    }
    userJsonStr, _ := json.Marshal(user)
    go logRequest(w, r, requestId, string(userJsonStr))

    if len(user.Login) < int(MIN_LOGIN_LENGTH) {
        writeResponse(w, r, requestId, JsonLoginTooShortResponse)
        return
    }

    if !isAlphaNumeric(user.Login) {
        writeResponse(w, r, requestId, JsonLoginInvalidCharResponse)
        return
    }

    if len(user.Password) < int(MIN_PWD_LENGTH) {
        writeResponse(w, r, requestId, JsonPwdTooShortResponse)
        return
    }

    if !isAlphaNumeric(user.Password) {
        writeResponse(w, r, requestId, JsonPwdInvalidCharResponse)
        return
    }

    stmt, err := users_db.Prepare(`INSERT INTO users (login, password) VALUES (?, ?)`)
    if err != nil {
        log.Fatalf("%s users: %v", getRequestContexString(r), err)
        writeResponse(w, r, requestId, JsonNullResponse)
        return
    }
    defer stmt.Close()

    res, err2 := stmt.Exec(user.Login, user.Password)
    if err != nil {
        log.Printf("%s users: %v\n", getRequestContexString(r), err2)
        writeResponse(w, r, requestId, JsonNullResponse)
        return
    }

    jsonStr := fmt.Sprintf(`{"id":%v}`, res)
    writeResponse(w, r, requestId, jsonStr)
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

    result := redis_cli.Cmd("HGETALL", "town:" + townId)
    if result.Err != nil {
        log.Printf("handlerTown: %v\n", result.Err)
        writeResponse(w, r, requestId, JsonNullResponse)
        return
    }

    data, err := result.Map()
    if err != nil {
        log.Printf("handlerTown: %v\n", err)
        writeResponse(w, r, requestId, JsonNullResponse)
        return
    }

    context := getRequestContexString(r) + " handlerTown:" + townId

    if len(data) == 0 {
        log.Printf("%s: no such town id\n", context)
        w.WriteHeader(404)
        return
    }

    town := new(Town)
    town.Name, _   = data["name"]
    town.NameTr, _ = data["name_tr"]

    id, err := strconv.ParseUint(townId, 10, 32)
    town.Id = checkConvertionUint(uint32(id), err, context + " => Town.Id")

    latitude, err := strconv.ParseFloat(data["latitude"], 32)
    town.Latitude = checkConvertionFloat(float32(latitude), err, context + " => Town.Latitude")

    longitude, err := strconv.ParseFloat(data["longitude"], 32)
    town.Longitude = checkConvertionFloat(float32(longitude), err, context + " => Town.Longitude")

    zoom, err := strconv.ParseUint(data["zoom"], 10, 32)
    town.Zoom = checkConvertionUint(uint32(zoom), err, context + " => Town.Zoom")

    jsonByteArr, _ := json.Marshal(town)
    jsonStr := string(jsonByteArr)
    writeResponse(w, r, requestId, jsonStr)
}

func handlerBank(w http.ResponseWriter, r * http.Request) {
    ok, requestId := prepareResponse(w, r)
    if ok == false {
        return
    }
    go logRequest(w, r, requestId, "")

    params := mux.Vars(r)
    bankId := params["id"]

    context := getRequestContexString(r) + " handlerBank:" + bankId

    result := redis_cli.Cmd("HGETALL", "bank:" + bankId)
    if result.Err != nil {
        log.Printf("%s => %v\n", context, result.Err)
        w.WriteHeader(500)
        return
    }

    data, err := result.Map()
    if err != nil {
        log.Printf("%s => %v\n", context, err)
        w.WriteHeader(500)
        return
    }

    if len(data) == 0 {
        log.Printf("%s => no such bank id\n", context)
        w.WriteHeader(404)
        return
    }

    bank := new(Bank)

    bank.Name   = data["name"]
    bank.NameTr = data["name_tr"]

    id, err := strconv.ParseUint(bankId, 10, 32)
    bank.Id = checkConvertionUint(uint32(id), err, context + " => Bank.Id")

    regionId, err := strconv.ParseUint(data["region_id"], 10, 32)
    bank.RegionId = checkConvertionUint(uint32(regionId), err, context + " => Bank.RegionId")

    jsonByteArr, _ := json.Marshal(bank)
    jsonStr := string(jsonByteArr)
    writeResponse(w, r, requestId, jsonStr)
}

func handlerBankCreate(w http.ResponseWriter, r *http.Request) {
    ok, requestId := prepareResponse(w, r)
    if ok == false {
        return
    }

    decoder := json.NewDecoder(r.Body)
    var bank Bank
    err := decoder.Decode(&bank)
    if err != nil {
        go logRequest(w, r, requestId, "")
        log.Println("Malformed User json")
        w.WriteHeader(400)
        return
    }
    userJsonStr, _ := json.Marshal(bank)
    go logRequest(w, r, requestId, string(userJsonStr))
}

func handlerCashpoint(w http.ResponseWriter, r *http.Request) {
    ok, requestId := prepareResponse(w, r)
    if ok == false {
        return
    }
    go logRequest(w, r, requestId, "")

    params := mux.Vars(r)
    cashPointId := params["id"]

    context := getRequestContexString(r) + " handlerCashpoint:" + cashPointId

    result := redis_cli.Cmd("HGETALL", "cp:" + cashPointId)
    if result.Err != nil {
        log.Printf("%s => %v\n", context, result.Err)
        w.WriteHeader(500)
        return
    }

    data, err := result.Map()
    if err != nil {
        log.Printf("%s => %v\n", context, err)
        w.WriteHeader(500)
        return
    }

    if len(data) == 0 {
        log.Printf("%s => no such cashpoint id\n", context)
        w.WriteHeader(404)
        return
    }

    cp := new(CashPoint)

    cp.Type           = data["type"]
    cp.Address        = data["address"]
    cp.AddressComment = data["address_comment"]
    cp.MetroName      = data["metro_name"]
    cp.Schedule       = data["schedule"]
    cp.Tel            = data["tel"]
    cp.Additional     = data["additional"]

    id, err := strconv.ParseUint(cashPointId, 10, 32)
    cp.Id = checkConvertionUint(uint32(id), err, context + " => CashPoint.Id")

    bankId, err := strconv.ParseUint(data["bank_id"], 10, 32)
    cp.BankId = checkConvertionUint(uint32(bankId), err, context + " => CashPoint.BankId")

    townId, err := strconv.ParseUint(data["town_id"], 10, 32)
    cp.TownId = checkConvertionUint(uint32(townId), err, context + " => CashPoint.TownId")

    latitude, err := strconv.ParseFloat(data["latitude"], 32)
    cp.Latitude = checkConvertionFloat(float32(latitude), err, context + " => CashPoint.Latitude")

    longitude, err := strconv.ParseFloat(data["longitude"], 32)
    cp.Longitude = checkConvertionFloat(float32(longitude), err, context + " => CashPoint.Longitude")

    freeAccess, err := strconv.ParseUint(data["free_access"], 10, 32)
    cp.FreeAccess = uintToBool(checkConvertionUint(uint32(freeAccess), err, context + " => CashPoint.FreeAccess"))

    mainOffice, err := strconv.ParseUint(data["main_office"], 10, 32)
    cp.MainOffice = uintToBool(checkConvertionUint(uint32(mainOffice), err, context + " => CashPoint.MainOffice"))

    withoutWeekend, err := strconv.ParseUint(data["without_weekend"], 10, 32)
    cp.WithoutWeekend = uintToBool(checkConvertionUint(uint32(withoutWeekend), err, context + " => CashPoint.WithoutWeekend"))

    roundTheClock, err := strconv.ParseUint(data["round_the_clock"], 10, 32)
    cp.RoundTheClock = uintToBool(checkConvertionUint(uint32(roundTheClock), err, context + " => CashPoint.RoundTheClock"))

    worksAsShop, err := strconv.ParseUint(data["works_as_shop"], 10, 32)
    cp.WorksAsShop = uintToBool(checkConvertionUint(uint32(worksAsShop), err, context + " => CashPoint.WorksAsShop"))

    rub, err := strconv.ParseUint(data["rub"], 10, 32)
    cp.Rub = uintToBool(checkConvertionUint(uint32(rub), err, context + " => CashPoint.Rub"))

    usd, err := strconv.ParseUint(data["usd"], 10, 32)
    cp.Usd = uintToBool(checkConvertionUint(uint32(usd), err, context + " => CashPoint.Usd"))

    eur, err := strconv.ParseUint(data["eur"], 10, 32)
    cp.Eur = uintToBool(checkConvertionUint(uint32(eur), err, context + " => CashPoint.Eur"))

    cashIn, err := strconv.ParseUint(data["cash_in"], 10, 32)
    cp.CashIn = uintToBool(checkConvertionUint(uint32(cashIn), err, context + " => CashPoint.CashIn"))

    jsonByteArr, _ := json.Marshal(cp)
    jsonStr := string(jsonByteArr)
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
    townId, _ := strconv.ParseUint(params["town_id"], 10, 32)
    bankId, _ := strconv.ParseUint(params["bank_id"], 10, 32)

    context := getRequestContexString(r) + " " + getHandlerContextString("handlerCashpointsByTownAndBank", requestId, params["town_id"], params["bank_id"])

    result := redis_cli.Cmd("SINTER", "town:" + params["town_id"] + ":cp",
                                      "bank:" + params["bank_id"] + ":cp")
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

    ids := CashPointIdsInTown{ TownId: uint32(townId), BankId: uint32(bankId) }
    if len(data) == 0 {
        ids.CashPointIds = make([]uint32, 0)
    }

    for i, idStr := range data {
        id, err := strconv.ParseUint(idStr, 10, 32)
        id32 := checkConvertionUint(uint32(id), err, context + " => CashPointIds[" + strconv.FormatInt(int64(i), 10) + "] = " + idStr)
        ids.CashPointIds = append(ids.CashPointIds, id32)
    }

    jsonByteArr, _ := json.Marshal(ids)
    jsonStr := string(jsonByteArr)
    writeResponse(w, r, requestId, jsonStr)
}

func main() {
    log.SetFlags(log.Flags() | log.Lmicroseconds)
    log.Println("CashPoints server build: " + BuildDate)

    args := os.Args[1:]

    if len(args) == 0 {
        log.Fatal("Config file path is not specified")
    }

    configFilePath := args[0]
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

    MIN_LOGIN_LENGTH = serverConfig.UserLoginMinLength
    MIN_PWD_LENGTH = serverConfig.UserPwdMinLength

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

    REQ_RES_LOG_TTL = serverConfig.ReqResLogTTL

    router := mux.NewRouter()
    router.HandleFunc("/user", handlerUserCreate).Methods("POST")
    router.HandleFunc("/town/{id:[0-9]+}", handlerTown)
    router.HandleFunc("/bank/{id:[0-9]+}", handlerBank)
    router.HandleFunc("/bank", handlerBankCreate).Methods("POST")
    router.HandleFunc("/cashpoint", handlerCashpointCreate).Methods("POST")
    router.HandleFunc("/cashpoint/{id:[0-9]+}", handlerCashpoint)
    router.HandleFunc("/town/{town_id:[0-9]+}/bank/{bank_id:[0-9]+}/cashpoints", handlerCashpointsByTownAndBank)

    port := ":" + strconv.FormatUint(serverConfig.Port, 10)
    log.Println("Listening 127.0.0.1" + port)

    server := &http.Server{
        Addr:           port,
        Handler:        router,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    //http.Handle("/", router)
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

