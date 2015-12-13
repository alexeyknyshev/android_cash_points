package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mediocregopher/radix.v2/redis"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
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
	Id        uint32 `json:"id"`
	Name      string `json:"name"`
	NameTr    string `json:"name_tr"`
	NameTrAlt string `json:"name_tr_alt"`
	Town      string `json:"town"`
	Licence   uint32 `json:"licence"`
	Rating    uint32 `json:"rating"`
	Tel       string `json:"tel"`
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

func migrateTowns(townsDb *sql.DB, redisCli *redis.Client) {
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
	var lastTownId uint32 = 0
	for rows.Next() {
		town := new(Town)
		var regionId uint32 = 0
		err = rows.Scan(&town.Id, &town.Name, &town.NameTr,
			&regionId, &town.RegionalCenter,
			&town.Latitude, &town.Longitude, &town.Zoom)
		if err != nil {
			log.Fatal(err)
		}

		if town.Id > lastTownId {
			lastTownId = town.Id
		}

		if regionId != 0 {
			town.RegionId = new(uint32)
			*town.RegionId = regionId
		}

		jsonData, err := json.Marshal(town)
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SET", "town:"+strconv.FormatUint(uint64(town.Id), 10), string(jsonData)).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("GEOADD", "towns", town.Longitude, town.Latitude, town.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		currentTownIdx++

		if currentTownIdx%500 == 0 {
			log.Printf("[%d/%d] Towns processed\n", currentTownIdx, townsCount)
		}
	}
	err = redisCli.Cmd("SET", "town_next_id", lastTownId).Err
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[%d/%d] Towns processed\n", townsCount, townsCount)
}

func migrateRegions(townsDb *sql.DB, redisCli *redis.Client) {
	var regionsCount int
	err := townsDb.QueryRow(`SELECT COUNT(*) FROM regions`).Scan(&regionsCount)
	if err != nil {
		log.Fatalf("migrate: regions: %v\n", err)
	}

	rows, err := townsDb.Query(`SELECT id, name, name_tr,
                                       latitude, longitude, zoom FROM regions`)
	if err != nil {
		log.Fatalf("migrate: regions: %v\n", err)
	}

	currentRegionIdx := 1
	for rows.Next() {
		region := new(Region)
		err = rows.Scan(&region.Id, &region.Name, &region.NameTr,
			&region.Latitude, &region.Longitude, &region.Zoom)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.Marshal(region)
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SET", "region:"+strconv.FormatUint(uint64(region.Id), 10), string(jsonData)).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("GEOADD", "regions", region.Longitude, region.Latitude, region.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		currentRegionIdx++

		if currentRegionIdx%500 == 0 {
			log.Printf("[%d/%d] Regions processed\n", currentRegionIdx, regionsCount)
		}
	}
	log.Printf("[%d/%d] Regions processed\n", regionsCount, regionsCount)
}

func migrateBanks(banksDb *sql.DB, redisCli *redis.Client) {
	var banksCount int
	err := banksDb.QueryRow(`SELECT COUNT(*) FROM banks`).Scan(&banksCount)
	if err != nil {
		log.Fatalf("migrate: banks: %v", err)
	}

	rows, err := banksDb.Query(`SELECT id, name, name_tr, name_tr_alt, town,
                                       licence, rating, tel FROM banks`)
	if err != nil {
		log.Fatalf("migrate: banks: %v", err)
	}

	currentBankIdx := 1
	var lastBankId uint32 = 0
	for rows.Next() {
		bank := new(Bank)
		var nameTr sql.NullString
		err = rows.Scan(&bank.Id, &bank.Name, &nameTr, &bank.NameTrAlt,
			&bank.Town, &bank.Licence, &bank.Rating, &bank.Tel)
		if err != nil {
			log.Fatal(err)
		}

		if bank.Id > lastBankId {
			lastBankId = bank.Id
		}

		if nameTr.Valid {
			bank.NameTr = nameTr.String
		} else {
			bank.NameTr = ""
		}

		jsonData, err := json.Marshal(bank)
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SET", "bank:"+strconv.FormatUint(uint64(bank.Id), 10), string(jsonData)).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SADD", "banks", bank.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		currentBankIdx++

		if currentBankIdx%100 == 0 {
			log.Printf("[%d/%d] Banks processed\n", currentBankIdx, banksCount)
		}
	}
	err = redisCli.Cmd("SET", "bank_next_id", lastBankId).Err
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[%d/%d] Banks processed\n", banksCount, banksCount)
}

func migrateCashpoints(cpDb *sql.DB, redisCli *redis.Client) {
	context := "migrateCashpoints"

	var cashpointsCount int
	err := cpDb.QueryRow(`SELECT COUNT(*) FROM cashpoints`).Scan(&cashpointsCount)
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
		/*
			_, err = stmt.Query(cp.Id, cp.Longitude, cp.Latitude)
			if err != nil {
				log.Fatalf("%s: cannot copy data to in-mem db: %v\n", context, err)
			}
		*/
		if cp.Id > lastCashpointId {
			lastCashpointId = cp.Id
		}

		cashpointIdStr := strconv.FormatUint(uint64(cp.Id), 10)
		townIdStr := strconv.FormatUint(uint64(cp.TownId), 10)
		bankIdStr := strconv.FormatUint(uint64(cp.BankId), 10)

		jsonData, err := json.Marshal(cp)
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SET", "cp:"+cashpointIdStr, string(jsonData)).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("GEOADD", "cashpoints", cp.Longitude, cp.Latitude, cp.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SADD", "town:"+townIdStr+":cp", cp.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("SADD", "bank:"+bankIdStr+":cp", cp.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		err = redisCli.Cmd("ZADD", "cp:history", 0, cp.Id).Err
		if err != nil {
			log.Fatal(err)
		}

		currentCashpointIndex++

		if currentCashpointIndex%500 == 0 {
			log.Printf("[%d/%d] Cashpoints processed\n", currentCashpointIndex, cashpointsCount)
		}
	}
	err = redisCli.Cmd("SET", "cp_next_id", lastCashpointId).Err
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[%d/%d] Cashpoints processed\n", cashpointsCount, cashpointsCount)
}

func uintPow(x, y uint32) uint32 {
	var result uint32 = 1
	for pow := uint32(0); pow < y; pow++ {
		result = result * x
	}
	return result
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

type CPClusteringRequest struct {
	Id        uint32
	Longitude float32
	Latitude  float32
}

type CPClusteringResponse struct {
	Id      uint32
	QuadKey string
}

func clusteringWorker(in chan CPClusteringRequest, maxZoom uint32) chan CPClusteringResponse {
	context := "clusteringWorker"
	out := make(chan CPClusteringResponse)
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
				response.QuadKey = quadKey
				// 				log.Printf("%s: response ready: id = %d, quadkey = %s", context, response.Id, response.QuadKey)
				out <- response
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
	out := make(chan CPClusteringResponse)

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

func migrateClustersNew(cpDb *sql.DB, redisCli *redis.Client) {
	context := "migrateClustersNew"

	log.Printf("%s: started", context)

	taskCount := 4
	channelsRequest := make([]chan CPClusteringRequest, taskCount)
	channelsResponse := make([]chan CPClusteringResponse, taskCount)

	cp := CPClusteringRequest{}

	var maxZoom uint32 = 16
	for i := 0; i < taskCount; i++ {
		channelsRequest[i] = make(chan CPClusteringRequest)
		channelsResponse[i] = clusteringWorker(channelsRequest[i], maxZoom)
	}

	log.Printf("%s: %d workers started", context, taskCount)

	var cashpointsCount int
	err := cpDb.QueryRow("SELECT COUNT(*) FROM cashpoints").Scan(&cashpointsCount)
	if err != nil {
		log.Fatalf("%s: cashpoints: %v\n", context, err)
	}

	rows, err := cpDb.Query("SELECT id, longitude, latitude FROM cashpoints")
	if err != nil {
		log.Fatalf("%s: cashpoints: %v\n", context, err)
	}

	wait := make(chan bool)
	go func() {
		cashpointIndex := 0
		progress := 0.0
		for response := range mergeResponseChannels(channelsResponse) {
			// 			log.Printf("%s: got response: id = %d, quadkey = %s", context, response.Id, response.QuadKey)
			result := redisCli.Cmd("SADD", "cluster:"+response.QuadKey, response.Id)
			if result.Err != nil {
				// 				log.Printf("%s: cannot add cp:%d to cluster:%s", context, response.Id, response.QuadKey)
				break
			}

			cashpointIndex++

			newProgress := math.Floor(float64(cashpointIndex) / float64(cashpointsCount) / float64(maxZoom) * 100.0)
			if newProgress > progress {
				progress = newProgress
				log.Printf("%s: [%3d%%] clustering done", context, int(progress))
			}
		}
		log.Printf("%s: all (%d) respones processed", context, cashpointIndex)
		wait <- true
	}()

	currentCashpointIndex := 0
	for rows.Next() {
		err = rows.Scan(&cp.Id, &cp.Longitude, &cp.Latitude)
		if err != nil {
			log.Fatalf("%s: sql scan error: %v\n", context, err)
			return
		}

		taskId := currentCashpointIndex % taskCount
		currentCashpointIndex++

		// 		log.Printf("%s: sending request to worker id = %d", context, taskId)
		channelsRequest[taskId] <- cp
	}

	for i := 0; i < taskCount; i++ {
		close(channelsRequest[i])
	}

	log.Printf("%s: all requests sent", context)

	<-wait
	log.Printf("%s: all tasks finished", context)
}

func migrateClusters(cpDb *sql.DB, redisCli *redis.Client) {
	context := "migrateClusters"

	var minLonWorld float32 = -180.0
	var maxLonWorld float32 = 180.0

	var minLatWorld float32 = -85.0
	var maxLatWorld float32 = 85.0

	minLon := minLonWorld
	maxLon := maxLonWorld

	minLat := minLatWorld
	maxLat := maxLatWorld

	err := cpDb.QueryRow("SELECT MIN(longitude), MAX(longitude), MIN(latitude), MAX(latitude) FROM cashpoints WHERE hidden = 0").Scan(&minLon, &maxLon, &minLat, &maxLat)
	if err != nil {
		log.Fatal("%s: cannot get region bounds: sql error: %v\n", context, err)
		return
	}

	minLon -= 1.0
	maxLon += 1.0

	minLat -= 1.0
	maxLat += 1.0

	log.Printf("%s: bounds = { [%f, %f], [%f, %f] }", context, minLon, maxLon, minLat, maxLat)

	stmt, err := cpDb.Prepare(`SELECT id, longitude, latitude FROM cashpoints
					WHERE latitude > ? AND latitude <= ? AND
					longitude > ? AND longitude <= ?`)
	if err != nil {
		log.Fatalf("%s: sql prepare error: %v\n", context, err)
		return
	}
	defer stmt.Close()

	var zoom uint32 = 0
	quadKey := ""
	var maxZoom uint32 = 16
	asyncSubCount := 5

	// beta ||
	task := newTask(zoom, minLatWorld, maxLatWorld, minLonWorld, maxLonWorld, quadKey)
	c := make(chan TaskResult)
	var wg sync.WaitGroup
	var dbMutex sync.Mutex

	wg.Add(1)
	go doTask(task, maxZoom, asyncSubCount, stmt, &dbMutex, &wg, c)

	go func() {
		clusterIndex := 0
		for t := range c {
			clusterIndex++

			if t.QuadKey != "" {
				// write cluster data

				log.Printf("%s: got task result: size = %d, quadkey = %s, lon = %f, lat = %f", context, len(t.Points), t.QuadKey, t.Longitude, t.Latitude)

				clusterGroupName := "zcluster:" + strconv.FormatUint(uint64(t.Zoom), 10)
				result := redisCli.Cmd("GEOADD", clusterGroupName, t.Longitude, t.Latitude, t.QuadKey)
				if result.Err != nil {
					log.Fatalf("%s: cannot add cluster group = %s: redis => %v\n", context, clusterGroupName, result.Err)
					return
				}

				clusterName := "cluster:" + t.QuadKey
				err = redisCli.Cmd("SADD", clusterName, t.Points).Err
				if err != nil {
					log.Printf("%s: cannot add ids into %s, redis => %v\n", context, clusterName, result.Err)
				}
			}

			log.Printf("%s: [%d] Clustering finished\n", context, clusterIndex)
		}
	}()

	wg.Wait()
}

func migrate(townsDb, cpDb, banksDb *sql.DB, redisCli *redis.Client) {
	migrateTowns(townsDb, redisCli)
	migrateRegions(townsDb, redisCli)
	migrateCashpoints(cpDb, redisCli)
	migrateBanks(banksDb, redisCli)
	//	migrateClusters(cpDb, redisCli)
	migrateClustersNew(cpDb, redisCli)
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
		log.Fatal("Redis database url is not specified")
	}

	townsDbPath := args[0]
	cashpointsDbPath := args[1]
	banksDbPath := args[2]
	redisUrl := args[3]

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

	redisCli, err := redis.Dial("tcp", redisUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer redisCli.Close()

	migrate(townsDb, cashpointsDb, banksDb, redisCli)
}
