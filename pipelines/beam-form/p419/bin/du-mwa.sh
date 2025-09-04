#!/usr/bin/env bash

cd /public/home
for i in {32..35};do
  size1=$(du -ms cstu00${i} 2>/dev/null | cut -f1)
  size2=$(du -ms cstu00${i}/scalebox | cut -f1)
  echo "i=$i, home_size=${size1}MB, scalebox_size=${size2}MB"
done

for i in {37..80};do
  size1=$(du -ms cstu00${i} 2>/dev/null | cut -f1)
  size2=$(du -ms cstu00${i}/scalebox | cut -f1)
  echo "i=$i, home_size=${size1}MB, scalebox_size=${size2}MB"
done
