#!/bin/bash

for i in {1..4};do
echo $i
ssh node$i rm -rf /dev/shm/scalebox/ /tmp/scalebox/
done