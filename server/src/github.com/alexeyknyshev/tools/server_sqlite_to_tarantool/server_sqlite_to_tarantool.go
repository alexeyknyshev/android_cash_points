package main

import (
	"database/sql"
	"errors"
	//"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tarantool/go-tarantool"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	//"bytes"
	//"sort"
	"time"
	//"io"
)

func boolToInt(val bool) uint {
	if val {
		return 1
	}
	return 0
}

type Town struct {
	Id             uint32  `json:"id"`
	Name           string  `json:"name"`
	NameTr         string  `json:"name_tr"`
	RegionId       *uint32 `json:"region_id"`
	RegionalCenter bool    `json:"regional_center"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	Zoom           uint32  `json:"zoom"`
	Big            bool    `json:"big"`
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

const CHAN_BUFFER_SIZE = 512

type CPClusteringRequest struct {
	Id        uint32
	Longitude float32
	Latitude  float32
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

			var minLon float32 = -180.0
			var maxLon float32 = 180.0

			var minLat float32 = -85.0
			var maxLat float32 = 85.0

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

func migrateTowns(townsDb *sql.DB, tnt *tarantool.Connection) {
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

// 		if town.Id > lastTownId {
// 			lastTownId = town.Id
// 		}

		if regionId != 0 {
			town.RegionId = new(uint32)
			*town.RegionId = regionId
		}

		coord := []float32{ town.Longitude, town.Latitude }

		resp, err := tnt.Insert(spaceId, []interface{}{
			uint(town.Id), coord, town.Name,
			town.NameTr, regionId, town.RegionalCenter,
			town.Zoom, town.Big,
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
	//log.Printf("Banks count: %d", len(bankList))

// 	err = redisCli.Cmd("SET", "bank_next_id", lastBankId).Err
// 	if err != nil {
// 		log.Fatal(err)
// 	}

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

// 		err = redisCli.Cmd("SET", "bank:"+strconv.FormatUint(uint64(bankId), 10), string(jsonData)).Err
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 
// 		err = redisCli.Cmd("SADD", "banks", bankId).Err
// 		if err != nil {
// 			log.Fatal(err)
// 		}

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

// 		err = redisCli.Cmd("SADD", "cp:town:"+townIdStr, cp.Id).Err
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 
// 		err = redisCli.Cmd("SADD", "cp:bank:"+bankIdStr, cp.Id).Err
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 
// 		err = redisCli.Cmd("ZADD", "cp:history", 0, cp.Id).Err
// 		if err != nil {
// 			log.Fatal(err)
// 		}

		currentCashpointIndex++

		if currentCashpointIndex%500 == 0 {
			log.Printf("[%d/%d] Cashpoints processed\n", currentCashpointIndex, cashpointsCount)
		}
	}
// 	err = redisCli.Cmd("SET", "cp_next_id", lastCashpointId).Err
// 	if err != nil {
// 		log.Fatal(err)
// 	}
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
			
// 			result := redisCli.Cmd("SADD", "cluster:"+response.QuadKey, response.Id)
// 			if result.Err != nil {
// 				log.Printf("%s: cannot add cp:%d to cluster:%s", context, response.Id, response.QuadKey)
// 				break
// 			}

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
	migrateTowns(townsDb, tnt)
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
