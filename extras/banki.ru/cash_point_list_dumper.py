#!/usr/bin/python3

import os, sys, sqlite3
import json
import requests

batchSize = 25

def createJsonDataPrefetch(townId):
  return {
    "jsonrpc": "2.0",
    "method": "bankGeo/getObjectsByFilter",
    "params": {
       "with_empty_coordinates": True,
       "limit": 25,
       "type": ["office", "branch", "atm", "cash", "self_office"],
       "region_id": [townId]
    },
    "id": "2"
  }

def createJsonData(offset, townId):
  return {
    "jsonrpc": "2.0",
    "method": "bankGeo/getObjectsByFilter",
    "params": {
       "with_empty_coordinates": True,
       "limit": 25,
       "offset": offset,
#       "type": ["atm", "self_office"],
       "type": ["office", "branch", "atm", "cash", "self_office"],
       "region_id": [townId]
     },
    "id": "2"
  }

# idList : stringList
def createJsonFullData(idList):
  return {
    "jsonrpc": "2.0",
    "method": "bank/getBankObjectsData",
    "params": {
        "id_list": idList
    },
    "id": "9"
  }
    
#{,,{"id_list":
#["4328310","7296301","7593802","7296298","1189208","495345","495378","495400",
#"495403","495375","495411","495382","495396","495409","495389","495355","548018",
#"6572679","6579419","6579439","6579443","6579444","6579445","6587966","6588499"]},}

def getCashPointsByRegionId(townId):
  url = "http://www.banki.ru/api/"
  headers = {
    "Connection": "keep-alive",
    "Accept": "application/json, text/javascript, */*; q=0.01",
    "Origin": "http://www.banki.ru",
    "X-Requested-With": "XMLHttpRequest",
    "User-Agent": "Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.130 Safari/537.36",
    "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
    "Referer": "http://www.banki.ru/banks/map/Moskva/",
    "Accept-Encoding": "gzip, deflate",
    "Accept-Language": "ru-RU,ru;q=0.8,en-US;q=0.6,en;q=0.4"
  }
  data = createJsonDataPrefetch(townId)
  
  r = requests.post(url, headers = headers, data = json.dumps(data))
  responseJson = r.json()
  #print(json.dumps(responseJson))
  #print(responseJson['result']['total'])
  total = int(responseJson['result']['total']);
  if total == 0:
    print('Zero total count received in prefetch for townId:', townId)
    return

  print("townId: " + str(townId) + "; total:" + str(total))
  prepared_tuples = []

  reqCount = total // batchSize
  #reqCount = 0
  for i in range(0, reqCount + 1):
    data = createJsonData(i * 25, townId)
    r = requests.post(url, headers = headers, data = json.dumps(data))
    responseJson = r.json()
    cashPoints = responseJson['result']['data']
    
    idList = []
    for cp in cashPoints:
      idList.append(cp['id'])
      #prepared_tuples.append( (cp['id'], cp['type'], cp['bank_id'], cp['is_main'], cp['longitude'], cp['latitude'], cp['address']) )

    fullData = createJsonFullData(idList)
    r = requests.post(url, headers = headers, data = json.dumps(fullData))
    responseFullJson = r.json()
    cashPoints = responseFullJson['result']['data']
    
    for cp in cashPoints:
      prepared_tuples.append(
    

    print(len(cashPoints))
    #print(responseJson)
    
  return prepared_tuples
    

if __name__ == "__main__":
  if len(sys.argv) < 2:
    print("town id is not specified")
    sys.exit(1)

  if len(sys.argv) < 3:
    print("output db file is not specified")
    sys.exit(2)

  townId = int(sys.argv[1])
  outputDB = sys.argv[2]
  
  if os.path.isfile(outputDB):
    os.remove(outputDB)
    print('removed file:', outputDB)
    
  bd = sqlite3.connect(outputDB)
  curs = bd.cursor()
  curs.execute('CREATE TABLE cashpoints (id integer primary key, type text, bank_id integer, class integer, longitude real, latitude real, address text)')

  prepared_tuples = getCashPointsByRegionId(townId)
  curs.executemany('INSERT INTO cashpoints VALUES (?,?,?,?,?,?,?)', prepared_tuples)
  
  bd.commit()
  bd.close()

#  r = requests.post(url, headers=headers, data=json.dumps(data));
#  responseJson = r.json()
#  print(responseJson)

#  'longitude', 'latitude', 'icon_url', 'is_main', 'address', 'name', 'bank_id', 'type'
  