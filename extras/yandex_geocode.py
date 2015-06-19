#!/usr/bin/python3

import urllib.request
import urllib.parse
from bs4 import BeautifulSoup

def getGeocode(opener, town, address):
  urlMask = "http://geocode-maps.yandex.ru/1.x/?geocode=%s,%s"
  url = urlMask % (town, address)
  f = opener.open(url)
  data = str(f.read(), 'utf-8')
  soup = BeautifulSoup(data)
  points = soup.find_all('pos')
  for p in points:
    print(p.getText())

if __name__ == "__main__":
  opener = urllib.request.FancyURLopener({})
  getGeocode(opener, urllib.parse.quote("г.Братск"), urllib.parse.quote("Гагарина ул., 18, корп. 7"))

