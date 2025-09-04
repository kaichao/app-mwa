#!/usr/bin/env bash

cd /public/home
for i in {32..35};do
  echo "i=$i"
  rm -rf cstu00${i}/scalebox/mydata/mwa
done

for i in {37..80};do
  echo "i=$i"
  rm -rf cstu00${i}/scalebox/mydata/mwa
done
