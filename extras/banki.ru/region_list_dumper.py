#!/usr/bin/python3

import os, sys, sqlite3
import json

def makeDBRow(town):
  l = town['region_code'].split('/')

  #regionCenter = False
  latinName = town['region_code']

  if len(l) > 1:
    #regionCenter = True
    latinName = l[1]
    latinName = latinName.replace('~', '').replace('_', ' ').replace('tss', 'z').replace('ssh', 'sh')

  #townName = town['region_name_full'];
  regionName = ""
  l = town['region_name_full'].split('(')
  if len(l) > 1:
    #townName = l[0].strip()
    regionName = l[1].replace(')', '').strip()

  return (town['region_id'], town['region_name'], latinName, regionName)


if __name__ == "__main__":
  if len(sys.argv) < 2:
    print("regions json file is not specified")
    sys.exit(1)

  if len(sys.argv) < 3:
    print("output db file is not specified")
    sys.exit(2)

  regionsJson = sys.argv[1]
  outputDB = sys.argv[2]

  if os.path.isfile(outputDB):
    os.remove(outputDB)
    print('removed file:', outputDB)

  index = 1
  tuple_index = 1
  count = 0

  prepared_tuples = []

  jsonData = ""
  with open(regionsJson, "r") as regionsJsonFile:
    jsonData = regionsJsonFile.read()

  jsonTree = json.loads(jsonData);
  data = jsonTree['data']

  for town in data:
    t = makeDBRow(town)
    prepared_tuples.append(t)

  regionList = { town[3] for town in prepared_tuples }
  #print(regionList)

  regionMap = {}
  regionId = 0
  for region in regionList:
    regionMap[region] = regionId
    regionId += 1

  #print(regionMap)

  bd = sqlite3.connect(outputDB)
  c = bd.cursor()
  c.execute('CREATE TABLE towns (id integer primary key, name text, name_tr text, region_id integer)')
  c.executemany('INSERT INTO towns VALUES (?,?,?,?)', [(town[0], town[1], town[2], regionMap[town[3]]) for town in prepared_tuples])
  c.execute('CREATE TABLE regions (id integer primary key, name text)')
  c.executemany('INSERT INTO regions VALUES (?, ?)', [(v, k) for k, v in regionMap.items()])
  bd.commit()
  bd.close()

  #print("Total count: ", count)
