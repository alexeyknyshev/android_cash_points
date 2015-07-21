#!/usr/bin/python3

import os, sys, sqlite3
import json
import requests

def getBanksData():
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
  reqData = {
    "jsonrpc": "2.0",
    "method": "bankInfo/getBankList",
    "params": {},
    "id": "test"
  }
  r = requests.post(url, headers = headers, data = json.dumps(reqData))
  responseJson = r.json()
  bankList = responseJson['result']['data']

  prepared_tuples = []
  for b in bankList:
    prepared_tuples.append( (b['bank_id'], b['bank_name'], b['licence'], b['name_eng'], b['region']) )

  return prepared_tuples

if __name__ == "__main__":
  if len(sys.argv) < 2:
    print("output db file is not specified")
    sys.exit(1)

  outputDB = sys.argv[1]

  if os.path.isfile(outputDB):
    os.remove(outputDB)
    print('removed file:', outputDB)

  bd = sqlite3.connect(outputDB)
  curs = bd.cursor()
  curs.execute('CREATE TABLE banks (id integer primary key, name text, licence integer, name_tr text, region text)')

  prepared_tuples = getBanksData()
  curs.executemany('INSERT INTO banks VALUES (?,?,?,?,?)', prepared_tuples)

  bd.commit()
  bd.close()
