package main

import (
    "os"
    "io"
    "log"
    "path"
    "strconv"
    "net/http"
    "database/sql"
    "encoding/json"
    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
)

// ========================================================

type ServerConfig struct {
    TownsDataBase      string `json:"TownsDataBase"`
    CashPointsDataBase string `json:"CashPointsDataBase"`
    CertificateDir     string `json:"CertificateDir"`
    Port               uint64 `json:"Port"`
}

const JsonNullResponse string = `{"id":null}`

func getRequestContexString(r *http.Request) string {
    return r.RemoteAddr
}

func prepareResponse(w http.ResponseWriter, r *http.Request) bool {
    contextStr := getRequestContexString(r)

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    requestIdStr := r.Header.Get("Id")
    if requestIdStr == "" {
        log.Println(contextStr + ` Request header val "Id" is not set`)
        return false
    }
    requestId, err := strconv.ParseUint(requestIdStr, 10, 32)
    if err != nil {
        log.Println(contextStr + ` Request header val "Id" uint conversion failed: ` + requestIdStr)
        io.WriteString(w, JsonNullResponse)
        return false
    }
    w.Header().Set("Id", strconv.FormatUint(requestId, 10))
    return true
}

// ========================================================

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

var towns_db *sql.DB
var cp_db *sql.DB

// ========================================================

func handlerTown(w http.ResponseWriter, r *http.Request) {
    if prepareResponse(w, r) == false {
        return
    }

    params := mux.Vars(r)
    townId := params["id"]

    stmt, err := towns_db.Prepare(`SELECT id, name, name_tr, latitude,
                                          longitude, zoom FROM towns WHERE id = ?`)
    if err != nil {
        log.Fatalf("%s %v", getRequestContexString(r), err)
    }
    defer stmt.Close()

    town := new(Town)
    err = stmt.QueryRow(townId).Scan(&town.Id, &town.Name, &town.NameTr,
                                     &town.Latitude, &town.Longitude, &town.Zoom)
    if err != nil {
        if err == sql.ErrNoRows {
            io.WriteString(w, JsonNullResponse)
            return
        } else {
            log.Fatalf("%s %v", getRequestContexString(r), err)
        }
    }

    jsonStr, _ := json.Marshal(town)
    io.WriteString(w, string(jsonStr))
}

func handlerCashpoint(w http.ResponseWriter, r *http.Request) {
    if prepareResponse(w, r) == false {
        return
    }

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
        log.Fatalf("%s %v", getRequestContexString(r), err)
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
            io.WriteString(w, JsonNullResponse)
            return
        } else {
            log.Fatalf("%s %v", getRequestContexString(r), err)
        }
    }

    jsonStr, _ := json.Marshal(cp)
    io.WriteString(w, string(jsonStr))
}

func handlerCashpointsByTownAndBank(w http.ResponseWriter, r *http.Request) {
    if prepareResponse(w, r) == false {
        return
    }

    params := mux.Vars(r)
    townId, _ := strconv.ParseUint(params["town_id"], 10, 32)
    bankId, _ := strconv.ParseUint(params["bank_id"], 10, 32)

    stmt, err := cp_db.Prepare("SELECT id FROM cashpoints WHERE town_id = ? AND bank_id = ?")
    if err != nil {
        log.Fatalf("%s %v", getRequestContexString(r), err)
    }
    defer stmt.Close()

    rows, err := stmt.Query(params["town_id"], params["bank_id"])
    if err != nil {
        if err == sql.ErrNoRows {
            io.WriteString(w, JsonNullResponse)
            return
        } else {
            log.Fatalf("%s %v", getRequestContexString(r), err)
        }
    }

    ids := CashPointIdsInTown{ TownId: uint32(townId), BankId: uint32(bankId) }

    for rows.Next() {
        var id uint32
        if err := rows.Scan(&id); err != nil {
            log.Fatalf("%s %v", getRequestContexString(r), err)
        }
        ids.CashPointIds = append(ids.CashPointIds, id)
    }

    jsonStr, _ := json.Marshal(ids)
    io.WriteString(w, string(jsonStr))
}

func main() {
    log.SetFlags(log.Flags() | log.Lmicroseconds)

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

    router := mux.NewRouter()
    router.HandleFunc("/town/{id:[0-9]+}", handlerTown)
    router.HandleFunc("/cashpoint/{id:[0-9]+}", handlerCashpoint)
    router.HandleFunc("/town/{town_id:[0-9]+}/bank/{bank_id:[0-9]+}/cashpoints", handlerCashpointsByTownAndBank)

    port := ":" + strconv.FormatUint(serverConfig.Port, 10)
    log.Println("Certificate path: " + certPath)
    log.Println("Private key path: " + pkeyPath)
    log.Println("Listening 127.0.0.1" + port)

    http.Handle("/", router)
    err = http.ListenAndServeTLS(port, certPath, pkeyPath, nil)
    if err != nil {
        log.Fatal(err)
    }
}

