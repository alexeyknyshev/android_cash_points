#!/usr/bin/python3

import urllib.request
from credcard_parser import parseHtmlData
import credcard_towns
import sqlite3

if __name__ == "__main__":
  opener = urllib.request.FancyURLopener({})
  urlMask = "http://www.credcard.ru/bankomat_pr.html?action=-1&rg=%s&town=-1&valut=0&sort=bank&vivod=50&search=&page=%i"

  #for region in range(0, 79):
  cp_count = 0
  for region in range(0, 80):
    #page = 43
    page = 1
    while True:
      url = urlMask % (region, page)
      print(url)
      f = opener.open(url)
      data = str(f.read(), 'cp1251')
      parsedData = parseHtmlData(data)
      cp_count += len(parsedData)
      if len(parsedData) == 0:
        break

#      for data in parsedData:



      print(parsedData)
      #break
      page += 1

  print("Total cp_count: " + str(cp_count))