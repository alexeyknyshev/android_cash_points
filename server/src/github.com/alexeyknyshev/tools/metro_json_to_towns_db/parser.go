package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strconv"
)

const PARSERING_FILE = "metro.json"
const DATABASE_NAME = "towns.db"

func dbOpen() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./"+DATABASE_NAME)
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

func insertRow(db *sql.DB, latitude, longitude float64, town_id, branch_id int64, name, ext string) error {
	query := "INSERT INTO metro (latitude, longitude, town_id, branch_id, name, ext) VALUES ('" +
		strconv.FormatFloat(latitude, 'g', -1, 64) + "','" +
		strconv.FormatFloat(longitude, 'g', -1, 64) + "','" +
		strconv.FormatInt(town_id, 10) + "','" +
		strconv.FormatInt(branch_id, 10) + "','" +
		name + "', '" + ext + "');"
	fmt.Println(query)
	_, err := db.Exec(query)
	return err
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
	Cells  Cells  `json:Cells`
}

func main() {
	//os.Remove("./" + DATABASE_NAME)
	db, err := dbOpen()
	if err != nil {
		return
	}
	defer db.Close()
	jsonFile, _ := os.Open(PARSERING_FILE)
	json.NewDecoder(jsonFile)
	dec := json.NewDecoder(jsonFile)

	t, err := dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	var tuple MetroTuple
	var latitude float64
	var longitude float64
	town_id := int64(4) //Moscow
	var branchName string
	var name string
	var ext string
	branchMap := make(map[string]int64)
	for dec.More() {

		// decode an array value
		err := dec.Decode(&tuple)
		if err != nil {
			log.Fatal(err)
		}
		latitude = tuple.Cells.Geo.Coords[0]
		longitude = tuple.Cells.Geo.Coords[1]
		name = tuple.Cells.NameOfStation
		ext = tuple.Cells.Name
		branchName = tuple.Cells.Line
		if branchMap[branchName] == 0 {
			branchMap[branchName] = int64(len(branchMap)) + int64(1)
		}
		insertRow(db, latitude, longitude, town_id, branchMap[branchName], name, ext)
		fmt.Println(tuple.Cells.Geo.Coords[0])
	}
	// read closing bracket
	t, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%T: %v\n", t, t)

	fmt.Println(branchMap)

}
