package main

import (
    "os"
    "fmt"
    "log"
    "strconv"
    "net/http"
    "database/sql"
    "encoding/json"
    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
)

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

func handlerTown(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    townId := params["id"]

    stmt, err := towns_db.Prepare("SELECT id, name, name_tr, latitude, longitude, zoom FROM towns WHERE id = ?")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

    town := new(Town)
    err = stmt.QueryRow(townId).Scan(&town.Id, &town.Name, &town.NameTr, &town.Latitude, &town.Longitude, &town.Zoom)
    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Fprintf(w, "{ id: null }")
            return
        } else {
            log.Fatal(err)
        }
    }

    jsonStr, _ := json.Marshal(town)
    fmt.Fprintf(w, string(jsonStr))
}

func handlerCashpoint(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    cashPointId := params["id"]

    stmt, err := cp_db.Prepare("SELECT id, type, bank_id, town_id, longitude, latitude, address, address_comment, metro_name, free_access, main_office, without_weekend, round_the_clock, works_as_shop, schedule_general, tel, additional, rub, usd, eur, cash_in FROM cashpoints WHERE id = ?")
    if err != nil {
        log.Fatal(err)
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
            fmt.Fprintf(w, "{ id: null }")
            return
        } else {
            log.Fatal(err)
        }
    }

    jsonStr, _ := json.Marshal(cp)
    fmt.Fprintf(w, string(jsonStr))
}

func handlerCashpointsByTownAndBank(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    townId, _ := strconv.ParseUint(params["town_id"], 10, 32)
    bankId, _ := strconv.ParseUint(params["bank_id"], 10, 32)

    stmt, err := cp_db.Prepare("SELECT id FROM cashpoints WHERE town_id = ? AND bank_id = ?")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

    rows, err := stmt.Query(params["town_id"], params["bank_id"])
    if err != nil {
        log.Fatal(err)
    }

    ids := CashPointIdsInTown{ TownId: uint32(townId), BankId: uint32(bankId) }

    for rows.Next() {
        var id uint32
        if err := rows.Scan(&id); err != nil {
            log.Fatal(err)
        }
        ids.CashPointIds = append(ids.CashPointIds, id)
    }

    jsonStr, _ := json.Marshal(ids)
    fmt.Fprintf(w, string(jsonStr))
}

func main() {
    args := os.Args[1:]

    if _, err := os.Stat(args[0]); os.IsNotExist(err) {
        log.Fatal("no such file or directory: %s\n", args[0])
        os.Exit(1)
    }

    if _, err := os.Stat(args[1]); os.IsNotExist(err) {
        log.Fatal("no such file or directory: %s\n", args[1])
        os.Exit(2)
    }

    var err error
    towns_db, err = sql.Open("sqlite3", args[0])
    if err != nil {
        log.Fatal(err)
        os.Exit(3)
    }
    defer towns_db.Close()

    cp_db, err = sql.Open("sqlite3", args[1])
    if err != nil {
        log.Fatal(err)
        os.Exit(4)
    }
    defer cp_db.Close()

    router := mux.NewRouter()
    router.HandleFunc("/town/{id:[0-9]+}", handlerTown)
    router.HandleFunc("/cashpoint/{id:[0-9]+}", handlerCashpoint)
    router.HandleFunc("/town/{town_id:[0-9]+}/bank/{bank_id:[0-9]+}/cashpoints", handlerCashpointsByTownAndBank)

    http.Handle("/", router)
    http.ListenAndServe(":8080", nil)
}