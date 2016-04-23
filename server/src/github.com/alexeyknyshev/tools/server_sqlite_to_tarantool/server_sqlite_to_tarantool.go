package main

import (
	"database/sql"
	"errors"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tarantool/go-tarantool"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	//"bytes"
	//"sort"
	"time"
	//"io"
	"unicode"
)

func boolToInt(val bool) uint {
	if val {
		return 1
	}
	return 0
}

type Town struct {
	Id              uint32  `json:"id"`
	Name            string  `json:"name"`
	NameTr          string  `json:"name_tr"`
	RegionId        *uint32 `json:"region_id"`
	RegionalCenter  bool    `json:"regional_center"`
	Latitude        float32 `json:"latitude"`
	Longitude       float32 `json:"longitude"`
	Zoom            uint32  `json:"zoom"`
	Big             bool    `json:"big"`
	CashpointsCount uint32  `json:"cashpoints_count"`
}

type Region struct {
	Id        uint32  `json:"id"`
	Name      string  `json:"name"`
	NameTr    string  `json:"name_tr"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	Zoom      uint32  `json:"zoom"`
}

type Bank struct {
	Id        uint32   `json:"id"`
	Name      string   `json:"name"`
	NameTr    string   `json:"name_tr"`
	NameTrAlt string   `json:"name_tr_alt"`
	Town      string   `json:"town"`
	Licence   uint32   `json:"licence"`
	Rating    uint32   `json:"rating"`
	Tel       string   `json:"tel"`
	Partners  []uint32 `json:"partners"`
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
	Version        uint32  `json:"version"`
	Timestamp      uint32  `json:"timestamp"`
}

type ClusterData struct {
	QuadKey   string  `json:"quadkey"`
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
	Size      uint32  `json:"size"`
}

// ================= Schedule =================

type ScheduleBreak struct {
	From int `json:"f"`
	To   int `json:"t"`
}

type ScheduleDay struct {
	Day   int `json:"-"`
	From  int `json:"f"`
	To    int `json:"t"`
	Breaks *[]ScheduleBreak `json:"b,omitempty"`
}

type Schedule struct {
	Mon *ScheduleDay `json:"mon,omitempty"`
	Tue *ScheduleDay `json:"tue,omitempty"`
	Wed *ScheduleDay `json:"wed,omitempty"`
	Thu *ScheduleDay `json:"thu,omitempty"`
	Fri *ScheduleDay `json:"fri,omitempty"`
	Sat *ScheduleDay `json:"sat,omitempty"`
	Sun *ScheduleDay `json:"sun,omitempty"`
	Breaks *[]ScheduleBreak `json:"b,omitempty"`
}

func (s *Schedule) SetCommonTime(from, to int) {
	if s.Mon != nil {
		s.Mon.From = from
		s.Mon.To = to
	}
	if s.Tue != nil {
		s.Tue.From = from
		s.Tue.To = to
	}
	if s.Wed != nil {
		s.Wed.From = from
		s.Wed.To = to
	}
	if s.Thu != nil {
		s.Thu.From = from
		s.Thu.To = to
	}
	if s.Fri != nil {
		s.Fri.From = from
		s.Fri.To = to
	}
	if s.Sat != nil {
		s.Sat.From = from
		s.Sat.To = to
	}
	if s.Sun != nil {
		s.Sun.From = from
		s.Sun.To = to
	}

// 	b := &ScheduleBreak{
// 		From: from,
// 		To:   to,
// 	}
// 	s.Breaks = append(b)
}

func (s *Schedule) Clear() {
	s.Mon = nil
	s.Tue = nil
	s.Wed = nil
	s.Thu = nil
	s.Fri = nil
	s.Sat = nil
	s.Sun = nil

	s.Breaks = nil
}

func (s *Schedule) Merge(o *Schedule) {
// 	sj, _ := json.Marshal(s)
// 	oj, _ := json.Marshal(o)
// 	log.Printf("merging %s with %s", string(sj), string(oj))

	if s.Mon == nil { s.Mon = o.Mon }
	if s.Tue == nil { s.Tue = o.Tue }
	if s.Wed == nil { s.Wed = o.Wed }
	if s.Thu == nil { s.Thu = o.Thu }
	if s.Fri == nil { s.Fri = o.Fri }
	if s.Sat == nil { s.Sat = o.Sat }
	if s.Sun == nil { s.Sun = o.Sun }

	if o.Breaks != nil {
		if s.Breaks != nil {
			*s.Breaks = append(*s.Breaks, *o.Breaks...)
		} else {
			s.Breaks = o.Breaks
		}
	}
}

func (s *Schedule) AppendDayRange(dayRange []int) {
	for _, d := range dayRange {
		switch (d) {
			case 0: if s.Mon == nil { s.Mon = new(ScheduleDay) }
			case 1: if s.Tue == nil { s.Tue = new(ScheduleDay) }
			case 2: if s.Wed == nil { s.Wed = new(ScheduleDay) }
			case 3: if s.Thu == nil { s.Thu = new(ScheduleDay) }
			case 4: if s.Fri == nil { s.Fri = new(ScheduleDay) }
			case 5: if s.Sat == nil { s.Sat = new(ScheduleDay) }
			case 6: if s.Sun == nil { s.Sun = new(ScheduleDay) }
		}
	}
}

var Days = [...]string{  "пн.", "вт.", "ср.", "чт.", "пт.", "сб.", "вс." }

const KT_INVALID = -1
const KT_BREAK = 0
const KT_DAY_RANGE = 1
const KT_DAY_SINGLE = 2

func keyType(s string) int {
	if (s == "перерыв") {
		return KT_BREAK // break
	} else if strings.IndexAny(s, "-—") != -1 {
		return KT_DAY_RANGE // day range
	} else {
		for _, d := range Days {
			if s == d {
				return KT_DAY_SINGLE
			}
		}
		return KT_INVALID // single day
	}
}

func parseDay(s string) int {
	for i, d := range Days {
		if d == s {
			return i
		}
	}
	return -1
}

func ParseDayRange(s string) ([]int, error) {
	result := make([]int, 0)
	s = strings.TrimRight(s, ":")
	s = strings.Replace(s, "—", "-", -1)
	s = strings.Replace(s, "–", "-", -1)
	parts := strings.Split(s, ",")
	for _, p := range parts {
		dayRange := strings.Split(p, "-")
		if len(dayRange) == 2 {
			rangeStart := parseDay(dayRange[0])
			rangeEnd := parseDay(dayRange[1])

			if rangeStart == -1 {
				return result, fmt.Errorf("invalid day range start: %s => %s", p, dayRange[0])
			} else if rangeEnd == -1 {
				return result, fmt.Errorf("invalid day range end: %s => %s", p, dayRange[1])
			} else if rangeStart >= rangeEnd {
				return result, fmt.Errorf("invalid day range, start gt end: %s", p)
			} else {
				for i := rangeStart; i <= rangeEnd; i++ {
					result = append(result, i)
				}
			}
		} else {
			day := parseDay(p)
			if day != -1 {
				result = append(result, day)
			} else {
				return result, fmt.Errorf("unknown day range format: %s", p)
			}
		}
	}
	return result, nil
}

func ParseTime(s string) (min int, ok bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		log.Printf("wrong time format: %s", s)
		min = 0; ok = false
		return
	}

	h, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Printf("wrong time format: %s", s)
		min = 0; ok = false
		return
	}

	min, err = strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("wrong time format: %s", s)
		min = 0; ok = false
		return
	}

	min = h * 60 + min
	ok = true
	return
}

// Splits strings like: 09:00—15:30. Return range in minutes from 00:00
func SplitTime(s string) (from int, to int, ok bool) {
	if (s == "круглосуточно") {
		from = 0; to = 1439; ok = true
		return
	}

	s = strings.Replace(s, "—", "-", -1)
	s = strings.Replace(s, "–", "-", -1)
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		log.Printf("wrong parts count in time range: %s", s)
		from = 0; to = 0; ok = false
		return
	}

	from, ok = ParseTime(parts[0])
	if !ok {
		from = 0; to = 0
		return
	}

	to, ok = ParseTime(parts[1])
	if !ok {
		from = 0; to = 0
		return
	}

	ok = true
	return
}

func ParseSchedule(schedule string) (Schedule, error) {
	var result Schedule
	if schedule == "" {
		return result, nil
	}

	strings.Replace(schedule, "<br/>", "\n", -1)

	parseKey := func(key string) (Schedule, error) {
		var result Schedule
		fields := strings.Split(key, ",")

		kType := -1
		for _, f := range fields {
// 			log.Printf("parsing dayRange part %s", f)
			kType = keyType(f)
// 			log.Printf("detected key type: %d", kType)

			if kType == KT_INVALID {
				continue
			} else if kType == KT_BREAK {
				// TODO: support parsing breaks
			} else if kType == KT_DAY_RANGE || kType == KT_DAY_SINGLE {
				dayRange, err := ParseDayRange(f)
// 				log.Printf("dayRange: %v", dayRange)
				if err != nil {
					return result, err
				}
				result.AppendDayRange(dayRange)
			}
		}

		return result, nil
	}

	lineList := strings.Split(schedule, "\n")
	for _, line := range lineList {
		fields := strings.Fields(line)
// 		fJson, _ := json.Marshal(fields)
// 		log.Printf("fields: %s", string(fJson))
		var tmpSchedule Schedule
		var err error
		for _, f := range fields {
			if strings.LastIndex(f, ":") == len(f) - 1 { // key
				f = f[:len(f) - 1]
// 				log.Printf("parsing key: %s", f)
				tmpSchedule, err = parseKey(f)
				if err != nil {
					return result, fmt.Errorf("cannot parse key: %s => %v", f, err)
				}
			} else { // value
				from, to, ok := SplitTime(f)
				if ok {
					tmpSchedule.SetCommonTime(from, to)
					j, _ := json.Marshal(tmpSchedule)
					log.Printf("%s", string(j))
					result.Merge(&tmpSchedule)
					tmpSchedule.Clear()
				} else {
					return result, fmt.Errorf("cannot split time: %s", f)
				}
			}
		}
	}

	return result, nil
}

// ================= Schedule =================

type Task struct {
	Zoom     uint32
	TopLat   float32
	BotLat   float32
	LeftLon  float32
	RightLon float32
	QuadKey  string
}

type TaskResult struct {
	Zoom      uint32
	Points    []uint32
	Longitude float32
	Latitude  float32
	QuadKey   string
}

func StringClear(r rune) rune {
	if r == '\n' {
		return ','
	}
	if unicode.IsPrint(r) {
		return r
	}
	return -1
}

func (cp *CashPoint) Postprocess() {
	cp.Address = strings.Map(StringClear, cp.Address)
	cp.AddressComment = strings.Map(StringClear, cp.AddressComment)
	//cp.Schedule = strings.Map(StringClear, cp.Schedule)
	schedule, err := ParseSchedule(cp.Schedule)
	if err != nil {
		log.Printf("Cannot parse schedule for cashpoint with id: %d", cp.Id)
		return
	}
	scheduleJson, _ := json.Marshal(schedule)
// 	log.Printf("%s", string(scheduleJson))
	cp.Schedule = string(scheduleJson)

	cp.Tel = strings.Map(StringClear, cp.Tel)
	cp.Additional = strings.Map(StringClear, cp.Additional)
}

func newTask(zoom uint32, topLat, botLat, leftLon, rightLon float32, quadKey string) *Task {
	return &Task{Zoom: zoom, TopLat: topLat, BotLat: botLat, LeftLon: leftLon, RightLon: rightLon, QuadKey: quadKey}
}

func getRegionIdList(topLat, botLat, leftLon, rightLon float32, stmt *sql.Stmt, mutex *sync.Mutex) (TaskResult, error) {
	context := "getRegionIdList"
	result := TaskResult{}

	mutex.Lock()
	defer mutex.Unlock()

	rows, err := stmt.Query(topLat, botLat, leftLon, rightLon)
	if err != nil {
		return result, err
	}

	result.Points = make([]uint32, 0)
	result.Longitude = 0.0
	result.Latitude = 0.0

	for rows.Next() {
		var id uint32 = 0
		var longitude float32 = 0.0
		var latitude float32 = 0.0

		err = rows.Scan(&id, &longitude, &latitude)
		if err != nil {
			log.Fatalf("%s: sql scan error: %v\n", context, err)
		}

		result.Longitude += longitude
		result.Latitude += latitude

		result.Points = append(result.Points, id)
	}

	count := len(result.Points)
	if count > 0 {
		result.Longitude = result.Longitude / float32(count)
		result.Latitude = result.Latitude / float32(count)
	}

	return result, nil
}

func doTask(task *Task, maxZoom uint32, asyncSubCount int, stmt *sql.Stmt, dbMutex *sync.Mutex, wg *sync.WaitGroup, c chan TaskResult) {
	context := "doTask"
	if asyncSubCount > 0 {
		asyncSubCount--
		defer wg.Done()
	}

	//log.Printf("%s: added task for quadkey = %s", context, task.QuadKey)

	result, err := getRegionIdList(task.TopLat, task.BotLat, task.LeftLon, task.RightLon, stmt, dbMutex)
	if err != nil {
		log.Fatalf("%s: cannot get cp ids for task (quadKey = %s): sql error: %v", context, task.QuadKey, err)
		return
	}

	count := len(result.Points)
	if count != 0 {

		// prepare subtasks

		if task.Zoom < maxZoom {
			nextZoom := task.Zoom + 1

			var midLat float32 = (task.TopLat + task.BotLat) * 0.5
			var midLon float32 = (task.LeftLon + task.RightLon) * 0.5

			taskList := make([]*Task, 0)
			taskList = append(taskList, newTask(nextZoom, task.TopLat, midLat, task.LeftLon, midLon, task.QuadKey+"0"))
			taskList = append(taskList, newTask(nextZoom, midLat, task.BotLat, task.LeftLon, midLon, task.QuadKey+"2"))
			taskList = append(taskList, newTask(nextZoom, task.TopLat, midLat, midLon, task.RightLon, task.QuadKey+"1"))
			taskList = append(taskList, newTask(nextZoom, midLat, task.BotLat, midLon, task.RightLon, task.QuadKey+"3"))

			asyncSubCount /= len(taskList)
			for _, task := range taskList {
				if asyncSubCount > 0 {
					wg.Add(1)
					go doTask(task, maxZoom, asyncSubCount, stmt, dbMutex, wg, c)
				} else {
					doTask(task, maxZoom, asyncSubCount, stmt, dbMutex, wg, c)
				}
			}
		}

		// write cluster data

		//		log.Printf("%s: finished task (quadkey = %s): count = %d, lon = %f, lat = %f\n", context, task.QuadKey, count, avgLon, avgLat)
		result.Zoom = task.Zoom
		result.QuadKey = task.QuadKey
		c <- result
	}
}

func getGeoRectPart(minLon, maxLon, minLat, maxLat *float64, lon, lat float64) string {
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

const CHAN_BUFFER_SIZE = 512

type CPClusteringRequest struct {
	Id        uint32
	Longitude float64
	Latitude  float64
}

type CPClusteringResponse struct {
	Id      uint32
	QuadKey string
	Zoom    uint32
}

func clusteringWorker(in chan CPClusteringRequest, minZoom, maxZoom uint32) chan CPClusteringResponse {
	context := "clusteringWorker"
	out := make(chan CPClusteringResponse, CHAN_BUFFER_SIZE)
	go func() {
		log.Printf("%s: waiting for task", context)
		for request := range in {
			// 			log.Printf("%s: got task: id = %d, lon = %f, lat = %f", context, request.Id, request.Longitude, request.Latitude)
			response := CPClusteringResponse{Id: request.Id}

			var minLon float64 = -180.0
			var maxLon float64 = 180.0

			var minLat float64 = -90.0
			var maxLat float64 = 90.0

			quadKey := ""
			for zoom := uint32(0); zoom < maxZoom; zoom++ {
				quadKey += getGeoRectPart(&minLon, &maxLon, &minLat, &maxLat, request.Longitude, request.Latitude)
				if zoom >= minZoom {
					response.QuadKey = quadKey
					response.Zoom = zoom
					//log.Printf("%s: response ready: id = %d, quadkey = %s", context, response.Id, response.QuadKey)
					out <- response
				}
			}
			// 			log.Printf("%s: clustering finished for cashpoint: %d", context, request.Id)
		}
		close(out)
	}()
	return out
}

func mergeResponseChannels(channels []chan CPClusteringResponse) chan CPClusteringResponse {
	context := "mergeResponseChannels"

	var wg sync.WaitGroup
	out := make(chan CPClusteringResponse, CHAN_BUFFER_SIZE * 4)

	output := func(c chan CPClusteringResponse) {
		for response := range c {
			// 			log.Printf("%s: got response: id = %d, quadkey = %s", context, response.Id, response.QuadKey)
			out <- response
		}
		wg.Done()
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go output(c)
		log.Printf("%s: started chan merger", context)
	}

	go func() {
		wg.Wait()
		close(out)
		log.Printf("%s: stopped chan merger", context)
	}()

	return out
}

func getTntSpaceId(tnt *tarantool.Connection, name string) (uint32, error) {
	resp, err := tnt.Call("getSpaceId", []interface{}{ name })
	if err != nil {	
		return 0, err
	}
	id := resp.Data[0].([]interface{})[0].(uint64)
	if id == 0 {
		return 0, errors.New("no such tarantool space: " + name)
	}
	return uint32(id), nil
}

func tntSpaceClear(tnt *tarantool.Connection, spaceId uint32) error {
	resp, err := tnt.Call("spaceTruncate", []interface{}{ spaceId })
	if err != nil {
		return err
	}
	ok := resp.Data[0].([]interface{})[0].(bool)
	if !ok {
		return errors.New("no such tarantool space: " + strconv.FormatUint(uint64(spaceId), 10))
	}
	return nil
}

func migrateMessages(townsDb *sql.DB, tnt *tarantool.Connection) {
	
}

func migrateTowns(townsDb, cpDb *sql.DB, tnt *tarantool.Connection) {
	spaceId, err := getTntSpaceId(tnt, "towns")
	if err != nil {
		log.Fatalf("cannot get space id: %v", err)
	}

	err = tntSpaceClear(tnt, spaceId)
	if err != nil {
		log.Fatalf("cannot drop space '%s': %v", "towns", err)
	}

	context := "migrateTowns"

	var townsCount int
	err = townsDb.QueryRow(`SELECT COUNT(*) FROM towns`).Scan(&townsCount)
	if err != nil {
		log.Fatalf("%s: %v\n", context, err)
	}

	rows, err := townsDb.Query(`SELECT id, name, name_tr, region_id,
                                       regional_center, latitude,
                                       longitude, zoom, has_emblem FROM towns`)

	if err != nil {
		log.Fatalf("%s: %v\n", context, err)
	}

	stmt, err := cpDb.Prepare(`SELECT COUNT(*) FROM cashpoints WHERE town_id = ?`)
	if err != nil {
		log.Fatalf("%s: sql prepare error: %v\n", context, err)
		return
	}

	currentTownIdx := 1
	for rows.Next() {
		town := Town{}
		var regionId uint32 = 0
		err = rows.Scan(&town.Id, &town.Name, &town.NameTr,
			&regionId, &town.RegionalCenter,
			&town.Latitude, &town.Longitude,
			&town.Zoom, &town.Big)
		if err != nil {
			log.Fatal(err)
		}

		err = stmt.QueryRow(town.Id).Scan(&town.CashpointsCount)
		if err != nil {
			log.Fatal(err)
		}

		if regionId != 0 {
			town.RegionId = new(uint32)
			*town.RegionId = regionId
		}

		coord := []float32{ town.Longitude, town.Latitude }

		resp, err := tnt.Insert(spaceId, []interface{}{
			uint(town.Id), coord, town.Name,
			town.NameTr, regionId, town.RegionalCenter,
			town.Zoom, town.Big, town.CashpointsCount,
		})
		if err != nil {
			log.Println("Insert")
			log.Println("Error", err)
			log.Println("Code", resp.Code)
			log.Println("Data", resp.Data)
			return;
		}

		currentTownIdx++

		if currentTownIdx%500 == 0 {
			log.Printf("[%d/%d] Towns processed\n", currentTownIdx, townsCount)
		}
	}

	log.Printf("[%d/%d] Towns processed\n", townsCount, townsCount)
}

func migrateRegions(townsDb *sql.DB, tnt *tarantool.Connection) {
	spaceId, err := getTntSpaceId(tnt, "regions")
	if err != nil {
		log.Fatalf("cannot get space id: %v", err)
	}

	err = tntSpaceClear(tnt, spaceId)
	if err != nil {
		log.Fatalf("cannot drop space '%s': %v", "regions", err)
	}

	context := "migrateRegions"

	var regionsCount int
	err = townsDb.QueryRow(`SELECT COUNT(*) FROM regions`).Scan(&regionsCount)
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}

	rows, err := townsDb.Query(`SELECT id, name, name_tr,
                                       latitude, longitude, zoom FROM regions`)
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}

	for rows.Next() {
		region := Region{}
		err = rows.Scan(&region.Id, &region.Name, &region.NameTr,
			&region.Latitude, &region.Longitude, &region.Zoom)
		if err != nil {
			log.Fatal(err)
		}

		coord := []float32{ region.Longitude, region.Latitude }

		resp, err := tnt.Insert(spaceId, []interface{}{
			uint32(region.Id), coord, region.Name, region.NameTr, region.Zoom,
		})
		
		if err != nil {
			log.Println("Insert")
			log.Println("Error", err)
			log.Println("Code", resp.Code)
			log.Println("Data", resp.Data)
			return;
		}
	}

	log.Printf("[%d/%d] Regions processed\n", regionsCount, regionsCount)
}

func migrateBanks(banksDb *sql.DB, tnt *tarantool.Connection) {
	spaceId, err := getTntSpaceId(tnt, "banks")
	if err != nil {
		log.Fatalf("cannot get space id: %v", err)
	}

	err = tntSpaceClear(tnt, spaceId)
	if err != nil {
		log.Fatalf("cannot drop space '%s': %v", "towns", err)
	}

	context := "migrateBanks"

	var banksCount int
	err = banksDb.QueryRow(`SELECT COUNT(*) FROM banks`).Scan(&banksCount)
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}

	rows, err := banksDb.Query(`SELECT id, name, name_tr, name_tr_alt, town,
                                       licence, rating, tel FROM banks`)
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}

	bankList := make([]Bank, 0)

	currentBankIdx := 1
	var lastBankId uint32 = 0
	for rows.Next() {
		bank := Bank{}
		bank.Partners = make([]uint32, 0)

		var nameTr sql.NullString
		err = rows.Scan(&bank.Id, &bank.Name, &nameTr, &bank.NameTrAlt,
			&bank.Town, &bank.Licence, &bank.Rating, &bank.Tel)
		if err != nil {
			log.Fatal(err)
		}

		if bank.Id > lastBankId {
			lastBankId = bank.Id
		}
		//log.Printf("bank processed: %d", bank.Id)

		if nameTr.Valid {
			bank.NameTr = nameTr.String
		} else {
			bank.NameTr = ""
		}

		bankList = append(bankList, bank)
	}

	stmt, err := banksDb.Prepare(`SELECT partner_id FROM partners WHERE id = ?`)
	if err != nil {
		log.Fatalf("%s: sql prepare error: %v\n", context, err)
		return
	}
	defer stmt.Close()

	for i := 0; i < len(bankList); i++ {
		bankId := bankList[i].Id
		partnerRows, err := stmt.Query(bankId)
		if err != nil {
			log.Fatalf("%s: sql partners get error: %v\n", context, err)
		}

		for partnerRows.Next() {
			var partnerId uint32
			partnerRows.Scan(&partnerId)
			if partnerId > 0 {
				bankList[i].Partners = append(bankList[i].Partners, partnerId)
			}
		}

// 		jsonData, err := json.Marshal(bankList[i])
// 		if err != nil {
// 			log.Fatal(err)
// 		}

		bank := bankList[i]

		resp, err := tnt.Insert(spaceId, []interface{}{
			uint32(bank.Id), bank.Name, bank.NameTr, bank.NameTrAlt,
			bank.Partners, bank.Town, bank.Licence, bank.Rating, bank.Tel,
		})
		
		if err != nil {
			log.Println("Insert")
			log.Println("Error", err)
			log.Println("Code", resp.Code)
			log.Println("Data", resp.Data)
			return;
		}

		currentBankIdx++

		if currentBankIdx%100 == 0 {
			log.Printf("[%d/%d] Banks processed\n", currentBankIdx, banksCount)
		}
	}

	log.Printf("[%d/%d] Banks processed\n", banksCount, banksCount)
}

func migrateCashpoints(cpDb *sql.DB, tnt *tarantool.Connection) {
	spaceId, err := getTntSpaceId(tnt, "cashpoints")
	if err != nil {
		log.Fatalf("cannot get space id: %v", err)
	}

	err = tntSpaceClear(tnt, spaceId)
	if err != nil {
		log.Fatalf("cannot drop space '%s': %v", "towns", err)
	}

	context := "migrateCashpoints"
	var cashpointsCount int
	err = cpDb.QueryRow(`SELECT COUNT(*) FROM cashpoints`).Scan(&cashpointsCount)
	if err != nil {
		log.Fatalf("%s: cashpoints: %v\n", context, err)
	}

	rows, err := cpDb.Query(`SELECT id, type, bank_id, town_id,
                                    longitude, latitude,
                                    address, address_comment,
                                    metro_name, free_access,
                                    main_office, without_weekend,
                                    round_the_clock, works_as_shop,
                                    schedule_general, tel, additional,
                                    rub, usd, eur, cash_in FROM cashpoints`)
	if err != nil {
		log.Fatalf("%s: cashpoints: %v\n", context, err)
	}

	currentCashpointIndex := 1
	var lastCashpointId uint32 = 0
	for rows.Next() {
		cp := new(CashPoint)
		cp.Version = 0
		cp.Timestamp = 0
		err = rows.Scan(&cp.Id, &cp.Type, &cp.BankId, &cp.TownId,
			&cp.Longitude, &cp.Latitude,
			&cp.Address, &cp.AddressComment,
			&cp.MetroName, &cp.FreeAccess,
			&cp.MainOffice, &cp.WithoutWeekend,
			&cp.RoundTheClock, &cp.WorksAsShop,
			&cp.Schedule, &cp.Tel, &cp.Additional,
			&cp.Rub, &cp.Usd, &cp.Eur, &cp.CashIn)
		if err != nil {
			log.Fatal(err)
		}

		cp.Postprocess()

		if cp.Id > lastCashpointId {
			lastCashpointId = cp.Id
		}

		//cashpointIdStr := strconv.FormatUint(uint64(cp.Id), 10)
		//townIdStr := strconv.FormatUint(uint64(cp.TownId), 10)
		//bankIdStr := strconv.FormatUint(uint64(cp.BankId), 10)

// 		jsonData, err := json.Marshal(cp)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

		coord := []float32{ cp.Longitude, cp.Latitude }

		resp, err := tnt.Insert(spaceId, []interface{}{
			uint32(cp.Id), coord, cp.Type, cp.BankId, cp.TownId,
			cp.Address, cp.AddressComment,
			cp.MetroName, cp.FreeAccess,
			cp.MainOffice, cp.WithoutWeekend,
			cp.RoundTheClock, cp.WorksAsShop,
			cp.Schedule, cp.Tel, cp.Additional,
			cp.Rub, cp.Usd, cp.Eur, cp.CashIn,
			cp.Version,
		})

		if err != nil {
			log.Println("Insert")
			log.Println("Error", err)
			log.Println("Code", resp.Code)
			log.Println("Data", resp.Data)
			return;
		}

		currentCashpointIndex++

		if currentCashpointIndex%500 == 0 {
			log.Printf("[%d/%d] Cashpoints processed\n", currentCashpointIndex, cashpointsCount)
		}
	}

	log.Printf("[%d/%d] Cashpoints processed\n", cashpointsCount, cashpointsCount)
}

func migrateClusters(cpDb *sql.DB, tnt *tarantool.Connection) (map[string][]uint32, error) {
	context := "migrateClusters"

	taskCount := 4
	channelsRequest := make([]chan CPClusteringRequest, taskCount)
	channelsResponse := make([]chan CPClusteringResponse, taskCount)

	var minZoom uint32 = 10
	var maxZoom uint32 = 16

	for i := 0; i < taskCount; i++ {
		channelsRequest[i] = make(chan CPClusteringRequest, CHAN_BUFFER_SIZE)
		channelsResponse[i] = clusteringWorker(channelsRequest[i], minZoom, maxZoom)
	}

	log.Printf("%s: %d workers started", context, taskCount)

	var cashpointsCount int
	err := cpDb.QueryRow("SELECT COUNT(*) FROM cashpoints WHERE hidden = 0").Scan(&cashpointsCount)
	if err != nil {
		log.Fatalf("%s: cashpoints: %v\n", context, err)
	}

	rows, err := cpDb.Query("SELECT id, longitude, latitude FROM cashpoints WHERE hidden = 0")
	if err != nil {
		log.Fatalf("%s: cashpoints: %v\n", context, err)
	}

	quadKeySet := make(map[string][]uint32, 0)
	
	wait := make(chan bool)
	go func() {
		cashpointIndex := 0
		progress := 0.0
		for response := range mergeResponseChannels(channelsResponse) {
			//log.Printf("%s: got response: id = %d, quadkey = %s", context, response.Id, response.QuadKey)
			if _, ok := quadKeySet[response.QuadKey]; !ok {
				quadKeySet[response.QuadKey] = make([]uint32, 0)
			}
			quadKeySet[response.QuadKey] = append(quadKeySet[response.QuadKey], response.Id)
			
			cashpointIndex++

			newProgress := math.Floor(float64(cashpointIndex) / float64(cashpointsCount) / float64(maxZoom - minZoom) * 100.0)
			if newProgress > progress {
				progress = newProgress
				log.Printf("%s: [%3d%%] clustering done", context, int(progress))
			}
		}
		log.Printf("%s: all (%d) respones processed", context, cashpointIndex)
		wait <- true
	}()

	cp := CPClusteringRequest{}
	currentCashpointIndex := 0
	for rows.Next() {
		err = rows.Scan(&cp.Id, &cp.Longitude, &cp.Latitude)
		if err != nil {
			log.Fatalf("%s: sql scan error: %v\n", context, err)
			return quadKeySet, err
		}

		taskId := currentCashpointIndex % taskCount
		currentCashpointIndex++

		//log.Printf("%s: sending request to worker id = %d", context, taskId)
		channelsRequest[taskId] <- cp
	}

	for i := 0; i < taskCount; i++ {
		close(channelsRequest[i])
	}

	log.Printf("%s: all requests sent", context)

	<-wait
	log.Printf("%s: all tasks finished", context)

	return quadKeySet, nil
}

type SortQuadKeys []string

func (s SortQuadKeys) Len() int {
	return len(s)
}

func (s SortQuadKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortQuadKeys) Less(i, j int) bool {
	if (len(s[i]) < len(s[j])) {
		return false
	}

	if (len(s[i]) == len(s[j])) {
		return s[i] < s[j]
	}

	return true
}

func migrateClustersGeo(tnt *tarantool.Connection, quadKeySet map[string][]uint32) {
	cpSpaceId, err := getTntSpaceId(tnt, "cashpoints")
	if err != nil {
		log.Fatalf("cannot get space id: %v", err)
	}

	clustersSpaceId, err := getTntSpaceId(tnt, "clusters")
	if err != nil {
		log.Fatalf("cannot get space id: %v", err)
	}

	tntSpaceClear(tnt, clustersSpaceId)

	context := "migrateClustersGeo"

	for quadKey, cpList := range quadKeySet {
		var avgLon float32 = 0.0
		var avgLat float32 = 0.0

		for _, id := range cpList {
			resp, err := tnt.Select(cpSpaceId, 0, 0, 1, tarantool.IterEq, []interface{}{uint(id)})
			if err != nil {
				log.Fatal("%s: cannot get cashpoint data by id: %d", context, id)
			}

			tuples := resp.Tuples()
// 			log.Println(tuples[0])
// 			log.Println(tuples[0][1])
			lon := tuples[0][1].([]interface{})[0].(float32)
			lat := tuples[0][1].([]interface{})[1].(float32)
// 			log.Println(lon)
// 			log.Println(lat)

			avgLon += lon
			avgLat += lat			

// 			return;
		}

		cpCount := len(cpList)
		avgLon = avgLon / float32(cpCount)
		avgLat = avgLat / float32(cpCount)

		resp, err := tnt.Insert(clustersSpaceId, []interface{}{
			quadKey, []float32{ avgLon, avgLat }, cpList, cpCount,
		})
		if err != nil {
			log.Println("Insert")
			log.Println("Error", err)
			log.Println("Code", resp.Code)
			log.Println("Data", resp.Data)
			return;
		}

// 		log.Println(quadKey)
	}
}

func migrate(townsDb, cpDb, banksDb *sql.DB, tnt *tarantool.Connection) {
	migrateMessages(townsDb, tnt)
	migrateTowns(townsDb, cpDb, tnt)
	migrateRegions(townsDb, tnt)
	migrateCashpoints(cpDb, tnt)
	migrateBanks(banksDb, tnt)
	quadKeyList, err := migrateClusters(cpDb, tnt)
	if err != nil {
		log.Fatalf("migrateClusters: cannot get list of quadkeys")
		return
	}
	migrateClustersGeo(tnt, quadKeyList)
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
		log.Fatal("Banks db file path is not specified")
	}

	if len(args) == 3 {
		log.Fatal("Tarantool database url is not specified")
	}

	townsDbPath := args[0]
	cashpointsDbPath := args[1]
	banksDbPath := args[2]
	tntUrl := args[3]

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

	banksDb, err := sql.Open("sqlite3", banksDbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer banksDb.Close()

	opts := tarantool.Opts{
		Reconnect: 1 * time.Second,
		MaxReconnects: 3,
		User: "admin",
 		Pass: "admin",
	}
	tnt, err := tarantool.Connect(tntUrl, opts)
	if err != nil {
		log.Fatal(err)
	}

	migrate(townsDb, cashpointsDb, banksDb, tnt)
}
