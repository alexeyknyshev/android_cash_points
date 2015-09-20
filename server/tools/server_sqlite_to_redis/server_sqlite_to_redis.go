package main

import (
    "os"
    "log"
    "strconv"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/mediocregopher/radix.v2/redis"
)

type Town struct {
    Id        uint32
    Name      string
    NameTr    string
    RegionId  uint32
    RegionalCenter bool
    Latitude  float32
    Longitude float32
    Zoom      uint32
}

func migrate(townsDb *sql.DB, cpDb *sql.DB, redisCli *redis.Client) {
    var townsCount int
    err := townsDb.QueryRow(`SELECT COUNT(*) FROM towns`).Scan(&townsCount)
    if err != nil {
        log.Fatalf("migrate: towns: %v\n", err)
    }

    rows, err := townsDb.Query(`SELECT id, name, name_tr, region_id,
                                       regional_center, latitude,
                                       longitude, zoom FROM towns`)
    if err != nil {
        log.Fatalf("migrate: towns: %v\n", err)
    }

    currentTownIdx := 1
    for rows.Next() {
        town := new(Town)
        err = rows.Scan(&town.Id, &town.Name, &town.NameTr,
                        &town.RegionId, &town.RegionalCenter,
                        &town.Latitude, &town.Longitude, &town.Zoom)
        if err != nil {
            log.Fatal(err)
        }

        err = redisCli.Cmd("HMSET", "town:" + strconv.FormatUint(uint64(town.Id), 10),
                                    "name", town.Name,
                                    "name_tr", town.NameTr,
                                    "region_id", town.RegionId,
                                    "regional_center", town.RegionalCenter,
                                    "latitude", town.Latitude,
                                    "longitude", town.Longitude,
                                    "zoom", town.Zoom).Err

        currentTownIdx++

        if currentTownIdx % 500 == 0 {
            log.Printf("[%d/%d] Towns processed\n", currentTownIdx, townsCount)
        }
    }
    log.Printf("[%d/%d] Towns processed\n", townsCount, townsCount)
}

func main() {
    args := os.Args[1:]

    if len(args) == 0 {
        log.Fatal("Towns db file path is not specified")
    }

    if len(args) == 1 {
        log.Fatal("Cashpoints db file path is not specified")
    }

    if len(args) == 2 {
        log.Fatal("Redis database url is not specified")
    }

    townsDbPath := args[0]
    cashpointsDbPath := args[1]
    redisUrl := args[2]

    townsDb, err := sql.Open("sqlite3", townsDbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer townsDb.Close()

    cashpointsDb, err := sql.Open("sqlite3", cashpointsDbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer cashpointsDb.Close()

    redisCli, err := redis.Dial("tcp", redisUrl)
    if err != nil {
        log.Fatal(err)
    }
    defer redisCli.Close()

    migrate(townsDb, cashpointsDb, redisCli)
}
