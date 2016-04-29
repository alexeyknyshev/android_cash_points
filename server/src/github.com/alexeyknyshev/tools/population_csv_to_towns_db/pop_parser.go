package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	name       = 0
	population = 1
)

func dbOpen(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	query, err := db.Query(`PRAGMA table_info("towns")`)
	var id, name, t, none string
	for query.Next() {
		query.Scan(&id, &name, &t, &none, &none, &none)
		if name == "population" {
			if strings.ToLower(t) != "integer" {
				err = errors.New("column population have type " + t + " expected INTEGER")
				return nil, err
			}
			query.Close()
			if err != nil {
				return nil, err
			}
			_, err = db.Exec(`UPDATE towns SET population = NULL`)
			return db, err
		}
	}
	query.Close()
	_, err = db.Exec(`ALTER TABLE towns 
					ADD population INTEGER`)
	return db, err
}

type popTuple struct {
	name       string
	population int64
}

func parseRecord(record []string) (popTuple, error) {
	var tuple popTuple
	tuple.name = record[name]
	var err error
	tuple.population, err = strconv.ParseInt(record[population], 10, 64)
	return tuple, err
}

func getMatchingTuples(tupleArray []popTuple, name string) []popTuple {
	var tempTuples []popTuple
	for _, tuple := range tupleArray {
		if strings.Contains(tuple.name, name) {
			tempTuples = append(tempTuples, tuple)
		}
	}
	return tempTuples
}

func searchCertainName(tempTuple []popTuple, name string) (popTuple, bool) {
	var tuple popTuple
	for _, tuple := range tempTuple {
		if tuple.name == name {
			return tuple, true
		}
	}
	return tuple, false
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("Towns db file path is not specified")
	}
	if len(args) == 1 {
		log.Fatal("population csv file path is not specified")
	}
	dbPath := args[0]
	csvPath := args[1]
	csvFile, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	db, err := dbOpen(dbPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()
	keyWords := []string{"автономный округ", "республика", " край ", "область", "муниципальный район", "городской округ"}
	skipRows := []string{"городское поселение", "район"}

	counter := 0
	skip := false
	tupleArray := []popTuple{}
	totalWarning := 0
	cont := false

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		for _, key := range keyWords {
			if strings.Contains(strings.ToLower(record[name]), key) {
				skip = false
				cont = true
				break
			}
		}
		if cont {
			cont = false
			continue
		}
		if skip {
			continue
		}
		for _, key := range skipRows {
			if strings.Contains(strings.ToLower(record[name]), key) {
				cont = true
				break
			}
		}
		if cont {
			cont = false
			continue
		}
		if strings.Contains(strings.ToLower(record[name]), "в том числе внутригородские") {
			skip = true
			continue
		}
		tempTuple, err := parseRecord(record)
		if err != nil {
			continue
		}
		tupleArray = append(tupleArray, tempTuple)
		counter++
	}
	fmt.Println("Amount of successful parse towns -", counter)

	rows, err := db.Query("SELECT id, name FROM towns")
	if err != nil {
		log.Fatal("db.Query err:", err)
	}
	rowCount := 0
	var searchTuple popTuple
	var found bool
	matchingTowns := 0
	popMap := make(map[int]int64)
	for rows.Next() {

		rowCount += 1
		if rowCount%500 == 0 {
			fmt.Println("processed", rowCount, "tuples")
		}
		var name, strId string
		rows.Scan(&strId, &name)
		if err != nil {
			log.Fatal("row.Scan err:", err)
		}

		tempTuples := getMatchingTuples(tupleArray, name)
		matchCount := len(tempTuples)
		if matchCount == 0 {
			totalWarning += 1
			continue
		}
		if matchCount > 1 {
			searchTuple, found = searchCertainName(tempTuples, "г. "+name)
			if !found {
				searchTuple, found = searchCertainName(tempTuples, "пгт. "+name)
				if !found {
					totalWarning += 1
					continue
				}
			}
		}
		if matchCount == 1 {
			searchTuple = tempTuples[0]
		}
		_ = searchTuple
		matchingTowns += 1
		id, _ := strconv.ParseInt(strId, 10, 64)
		popMap[int(id)] = searchTuple.population
	}
	rows.Close()
	fmt.Println("popArray len =", len(popMap))
	fmt.Println("total warning -", totalWarning)
	fmt.Println("matching towns -", matchingTowns)

	stmt, err := db.Prepare(`UPDATE towns SET population = ?
						WHERE id = ?`)
	if err != nil {
		log.Fatal("db.Prepare err:", err)
	}
	for id, value := range popMap {
		_, err = stmt.Exec(strconv.FormatInt(value, 10), strconv.FormatInt(int64(id), 10))
		if err != nil {
			fmt.Println("update query err:", err)
		}
	}
	stmt.Close()
	fmt.Println("Migration was done successfully")
}
