#!/bin/bash

# 1301240224/1301241455_1301241484_ch128.dat.tar.zst~b00

m=$1
# 从右到左删除第一次匹配的模式及其右边的所有内容
file="${m%%~*}"

/app/share/bin/run.sh "$file" "$2"
