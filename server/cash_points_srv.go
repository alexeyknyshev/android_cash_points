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
)

// ========================================================

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
    RequestsDataBase   string `json:"RequestsDataBase"`
    CertificateDir     string `json:"CertificateDir"`
    Port               uint64 `json:"Port"`
    UserLoginMinLength uint64 `json:"UserLoginMinLength"`
    UserPwdMinLength   uint64 `json:"UserPwdMinLength"`
    UseTLS             bool   `json:"UseTLS"`
}

func getRequestContexString(r *http.Request) string {
    return r.RemoteAddr
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

func logRequest(w http.ResponseWriter, r *http.Request, requestId int64, requestBody string) error {
    stmt, err := requests_db.Prepare(`INSERT INTO requests (path, data, time, user_id) VALUES (?, ?, ?, ?)`)
    if err != nil {
        log.Fatalf("%s requests: %v", getRequestContexString(r), err)
        io.WriteString(w, JsonNullResponse)
        return err
    }
    defer stmt.Close()

    path := r.URL.Path

    reqId := sql.NullInt64{ Int64: requestId, Valid: requestId != 0 }
    reqBody := sql.NullString{ String: requestBody, Valid: requestBody != "" }

    _, err2 := stmt.Exec(path, reqBody, strconv.FormatInt(time.Now().Unix(), 10), reqId)
    if err2 != nil {
        log.Printf("%s requests: %v\n", getRequestContexString(r), err2)
        io.WriteString(w, JsonNullResponse)
        return err2
    }
    
    return nil
}

func logResponse(context string, requestId int64, responseBody string) error {
    stmt, err := requests_db.Prepare(`INSERT INTO responses (data, time, user_id) VALUES (?, ?, ?)`)
    if err != nil {
        log.Fatalf("%s responses: %v", context, err)
        return err
    }
    defer stmt.Close()

    reqId := sql.NullInt64{ Int64: requestId, Valid: requestId != 0 }
    resBody := sql.NullString{ String: responseBody, Valid: responseBody != "" }

    _, err2 := stmt.Exec(resBody, strconv.FormatInt(time.Now().Unix(), 10), reqId)
    if err2 != nil {
        log.Printf("%s responses: %v\n", context, err2)
        return err2
    }
    
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

var towns_db *sql.DB
var cp_db *sql.DB
var users_db *sql.DB
var requests_db *sql.DB

var MIN_LOGIN_LENGTH uint64 = 4
var MIN_PWD_LENGTH uint64 = 4

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
        writeResponse(w, r, requestId, JsonNullResponse)
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

    stmt, err := towns_db.Prepare(`SELECT id, name, name_tr, latitude,
                                          longitude, zoom FROM towns WHERE id = ?`)
    if err != nil {
        log.Fatalf("%s towns: %v", getRequestContexString(r), err)
    }
    defer stmt.Close()

    town := new(Town)
    err = stmt.QueryRow(townId).Scan(&town.Id, &town.Name, &town.NameTr,
                                     &town.Latitude, &town.Longitude, &town.Zoom)
    if err != nil {
        if err == sql.ErrNoRows {
            writeResponse(w, r, requestId, JsonNullResponse)
            return
        } else {
            log.Fatalf("%s towns: %v", getRequestContexString(r), err)
        }
    }

    jsonByteArr, _ := json.Marshal(town)
    jsonStr := string(jsonByteArr)
    writeResponse(w, r, requestId, jsonStr)
}

func handlerCashpoint(w http.ResponseWriter, r *http.Request) {
    ok, requestId := prepareResponse(w, r)
    if ok == false {
        return
    }
    go logRequest(w, r, requestId, "")

    params := mux.Vars(r)
    cashPointId := params["id"]

    stmt, err := cp_db.Prepare(`SELECT id, type, bank_id, town_id, longitude,
                                       latitude, address, address_comment,
                                       metro_name, free_access, main_office,
                                       without_weekend, round_the_clock,
                                       works_as_shop, schedule_general, tel,
                                       additional, rub, usd, eur,
                                       cash_in FROM cashpoints WHERE id = ?`)
    if err != nil {
        log.Fatalf("%s cashpoints: %v", getRequestContexString(r), err)
    }
    defer stmt.Close()

    cp := new(CashPoint)
    // Todo: parsing schedule
    err = stmt.QueryRow(cashPointId).Scan(&cp.Id, &cp.Type, &cp.BankId,
                                          &cp.TownId, &cp.Longitude, &cp.Latitude,
                                          &cp.Address, &cp.AddressComment,
                                          &cp.MetroName, &cp.FreeAccess,
                                          &cp.MainOffice, &cp.WithoutWeekend,
                                          &cp.RoundTheClock, &cp.WorksAsShop,
                                          &cp.Schedule, &cp.Tel, &cp.Additional,
                                          &cp.Rub, &cp.Usd, &cp.Eur, &cp.CashIn)
    if err != nil {
        if err == sql.ErrNoRows {
            writeResponse(w, r, requestId, JsonNullResponse)
            return
        } else {
            log.Fatalf("%s cashpoints: %v",getRequestContexString(r), err)
        }
    }

    jsonByteArr, _ := json.Marshal(cp)
    jsonStr := string(jsonByteArr)
    writeResponse(w, r, requestId, jsonStr)
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

    stmt, err := cp_db.Prepare("SELECT id FROM cashpoints WHERE town_id = ? AND bank_id = ?")
    if err != nil {
        log.Fatalf("%s cashpoints: %v", getRequestContexString(r), err)
    }
    defer stmt.Close()

    rows, err := stmt.Query(params["town_id"], params["bank_id"])
    if err != nil {
        if err == sql.ErrNoRows {
            writeResponse(w, r, requestId, JsonNullResponse)
            return
        } else {
            log.Fatalf("%s cashpoints: %v", getRequestContexString(r), err)
        }
    }

    ids := CashPointIdsInTown{ TownId: uint32(townId), BankId: uint32(bankId) }

    for rows.Next() {
        var id uint32
        if err := rows.Scan(&id); err != nil {
            log.Fatalf("%s cashpoints: %v", getRequestContexString(r), err)
        }
        ids.CashPointIds = append(ids.CashPointIds, id)
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

    certPath := path.Join(serverConfig.CertificateDir, "cert.pem")
    pkeyPath := path.Join(serverConfig.CertificateDir, "key.pem")

    if _, err := os.Stat(certPath); os.IsNotExist(err) {
        log.Fatalf("No such cert file for tls: %s\n", certPath)
    }

    if _, err := os.Stat(pkeyPath); os.IsNotExist(err) {
        log.Fatalf("No such private key file for tls: %s\n", pkeyPath)
    }

    towns_db, err = sql.Open("sqlite3", serverConfig.TownsDataBase)
    if err != nil {
        log.Fatal(err)
    }
    defer towns_db.Close()

    cp_db, err = sql.Open("sqlite3", serverConfig.CashPointsDataBase)
    if err != nil {
        log.Fatal(err)
    }
    defer cp_db.Close()

    requests_db, err = sql.Open("sqlite3", serverConfig.RequestsDataBase)
    if err != nil {
        log.Fatal(err)
    }
    defer requests_db.Close()

    router := mux.NewRouter()
    router.HandleFunc("/user", handlerUserCreate).Methods("POST")
    router.HandleFunc("/town/{id:[0-9]+}", handlerTown)
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

