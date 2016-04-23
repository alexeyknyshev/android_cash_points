package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

func dbOpen(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP TABLE IF EXISTS metro")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE metro (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						latitude REAL,
						longitude REAL,
						town_id INTEGER,
						branch_id INTEGER,
						name TEXT,
						ext TEXT
						);`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type GeoData struct {
	Type   string    `json:"type"`
	Coords []float64 `json:"coordinates"`
}

type RepairOfEscalators struct {
	ROF string `json:"RepairOfEscalators"`
}

type Cells struct {
	GlobalId                  int                  `json:"global_id"`
	Name                      string               `json:"Name"`
	Longitude_WGS84           string               `json:"Longitude_WGS84"`
	Latitude_WGS84            string               `json:"Latitude_WGS84"`
	NameOfStation             string               `json:"NameOfStation"`
	Line                      string               `json:"Line"`
	ModeOnEvenDays            string               `json:"ModeOnEvenDays"`
	ModeOnOddDays             string               `json:"ModeOnOddDays"`
	FullFeaturedBPAAmount     int                  `json:"FullFeaturedBPAAmount"`
	LittleFunctionalBPAAmount int                  `json:"LittleFunctionalBPAAmount"`
	BPAAmount                 int                  `json:"BPAAmount"`
	Repair                    []RepairOfEscalators `json:"RepairOfEscalators"`
	Geo                       GeoData              `json:"geoData"`
}

type MetroTuple struct {
	Id     string `json:"Id"`
	Number int    `json:"Number"`
	Cells  Cells  `json:"Cells"`
}

func main() {
	args := os.Args[1:]
	fmt.Println("Migration has begun")
	if len(args) == 0 {
		log.Fatal("Towns db file path is not specified")
	}
	if len(args) == 1 {
		log.Fatal("Metro json file path is not specified")
	}
	dbPath := args[0]
	jsonPath := args[1]
	db, err := dbOpen(dbPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	jsonFile, _ := os.Open(jsonPath)
	json.NewDecoder(jsonFile)
	dec := json.NewDecoder(jsonFile)

	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}

	var tuple MetroTuple
	town_id := int64(4) //Moscow
	branchMap := make(map[string]int64)
	stmt, err := db.Prepare(`INSERT INTO metro (latitude, longitude, town_id, branch_id, name, ext) VALUES (?, ?, ?, ?, ?, ?)`)
	for dec.More() {

		// decode an array value
		err := dec.Decode(&tuple)
		if err != nil {
			log.Fatal(err)
		}
		latitude := tuple.Cells.Geo.Coords[1]
		longitude := tuple.Cells.Geo.Coords[0]
		name := tuple.Cells.NameOfStation
		ext := tuple.Cells.Name
		branchName := tuple.Cells.Line
		if branchMap[branchName] == 0 {
			branchMap[branchName] = int64(len(branchMap)) + int64(1)
		}
		_, err = stmt.Exec(latitude, longitude, town_id, branchMap[branchName], name, ext)
		if err != nil {
			log.Fatal(err)
		}
	}
	// read closing bracket
	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migration was done susseccfully")
	fmt.Println(branchMap)

}
