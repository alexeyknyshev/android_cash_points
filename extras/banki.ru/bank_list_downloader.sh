#!/bin/bash

echo "started downloading..."
for i in {1..16}
do
  url="http://www.banki.ru/banks/?order=fin_rating&PAGEN_1=$i"
  output="data/page_$i.html"
  wget "$url" --output-document="$output"
  if [ $? -eq 0 ]
  then
    echo "[OK] \"$url\" => \"$output\""
  else
    echo "[FAIL] \"$url\" => \"$output\""
  fi
done
echo "downloading finished!"

#for f in *.html; do iconv -f CP1251 -t UTF-8 "$f" > "$f.out"; done
