#!/usr/bin/python3

import sys
import codecs
import collections
import fnmatch
from bs4 import BeautifulSoup

# site schema
CP_INDEX       = 0
CP_BANK        = 1
CP_CARDS       = 2
CP_TOWN        = 3
CP_ADDR_METRO  = 4
CP_DESCRIPTION = 5
CP_CURRENCY    = 6
CP_TIME        = 7

class CashPoint:
  def __init__(self):
    self.bank = ""         # bank name
    self.cards = []        # visa = 0, electron = 1, plus = 2, master = 3, maestro = 4, amex = 5
    self.town = ""         # town name
    self.region = ""       # region name
    self.address = ""      # free fromat addr string
    self.metro = []        # metro index as in metro.yandex.ru
    self.description = ""  # free format string
    self.currency = []     # RUR = 0, USD = 1, EUR = 2, unknown = -1
    self.time = ""         # time in format 13:30-22:00
    self.atc = False       # is around the clock cash point
    self.org_s = False     # organization schedule
  
  def __str__(self):
    return "\tbank\t\t= %s\n\tcards\t\t= %s\n\ttown\t\t= %s\n\tregion\t\t= %s\n\taddress\t\t= %s\n\tmetro\t\t= %s\n\tdescription\t= %s\n\tcurrency\t= %s\n\ttime\t\t= %s\n\tatc\t\t= %s\n\torg_s\t\t= %s\n" % (self.bank, str(self.cards), self.town, self.region, self.address, str(self.metro), str(self.description), str(self.currency), self.time, str(self.atc), str(self.org_s))
  
  def __repr__(self):
    return "\n{\n" + self.__str__() + "}"
    
def convertMoneyType(moneyStr):
  if moneyStr == "p":
    return 0
  if moneyStr == "$":
    return 1
  if moneyStr == "€":
    return 2
  return -1

def convertTime(timeStr):
  timeAtcOrg = collections.namedtuple('TimeAndAtc', ['time', 'atc', 'org_s'])
  
  if timeStr == "":
    return timeAtcOrg("", False, False)
  
  if timeStr == "24 часа":
    return timeAtcOrg("", True, False)
  
  if fnmatch.fnmatch(timeStr, "* работы *"):
    return timeAtcOrg("", False, True)
  
  return timeAtcOrg(timeStr, False, False)

def parseHtmlData(data):
  result = []
  data = str(data).replace("<br>",";")
  soup = BeautifulSoup(data)
#  table = soup.find('table', attrs={'bgcolor':"#333333"})
#  rows = table.find('tbody').find_all('tr')
  rows = soup.find_all('tr', attrs={'bgcolor':"#FFFFFF"})
  for row in rows:
    cols = row.find_all('td')
    
    #print("here0: " + str(cols[0].getText()))
    #print("here1: " + str(cols[1].getText()))
    #print("here2: " + str(cols[2].getText()))
    #print("here3: " + str(cols[3].getText()))
    
    if not all(char.isdigit() for char in cols[CP_INDEX].getText()):
      continue
    
#    print("there")
    
    index = int(cols[CP_INDEX].getText())
    bank  = cols[CP_BANK].getText().strip()
    cards = [card.strip() for card in cols[CP_CARDS].getText().split(',')]
    townAndRegion = cols[CP_TOWN].getText().strip().split(';')
    #print(cols[CP_TOWN].getText())
    #print(':'.join(hex(ord(x))[2:] for x in cols[CP_TOWN].getText()))
    town = townAndRegion[0]
    region = townAndRegion[1]
    addrAndMetro = cols[CP_ADDR_METRO].getText().replace('\n', ' ').replace('\t', '').strip().split(' м.')
    addr = addrAndMetro[0]
    
    metro = ""
    if len(addrAndMetro) > 1:
      metro = [m.strip() for m in addrAndMetro[1].split(',')]
      
    descr = cols[CP_DESCRIPTION].getText().strip()
    if len(descr) > 0 and descr[len(descr) - 1] == ';':
      descr = descr[:-1]
      descr = descr.strip()
      
    if len(addr) > 0 and addr[len(addr) - 1] == ';':
      addr = addr[:-1]
      addr = addr.strip()
    
    
    #print('index: ' + str(index))
    #print('cards: ' + str(cards))
    #print('addr:  ' + addr)
    #print('metro: ' + str(metro))
    #print('descr: ' + cols[3].getText().strip())
    #print('type:  ' + cols[4].getText().strip())
    #print('time:  ' + cols[5].getText())
    #print('============================================')
    
    cp = CashPoint()
    cp.bank = bank
    cp.cards = cards
    cp.town = town
    cp.region = region
    cp.address = addr
    cp.metro = metro
    cp.description = descr
    
    currencyStr = cols[CP_CURRENCY].getText().strip()
    if currencyStr == "н/д":
      cp.currency = [-1]
    else:
      cp.currency = [convertMoneyType(m) for m in cols[CP_CURRENCY].getText().strip().split(' ')]
    
    tao = convertTime(cols[CP_TIME].getText().strip())
    cp.time  = tao.time
#    cp.time  = cols[5].getText().strip()
    cp.atc   = tao.atc
    cp.org_s = tao.org_s
    
    result.append(cp)
  
  return result

def parseHtmlFile(filenameStr):
  f = codecs.open(filenameStr, encoding='cp1251')
  data = f.read()
  return parseHtmlData(data)

if __name__ == "__main__":
  if len(sys.argv) != 2:
    print("credcard_parser: wrong argv count!")
    sys.exit(1)
  
  print("argc: " + str(len(sys.argv)))
  print("credcard_parser: " + sys.argv[1])
  
  cashpoints = parseHtmlFile(sys.argv[1])
  
  i = 0
  for cp in cashpoints:
    print("#### " + str(i) + " ####")
    print(str(cp))
    i = i + 1