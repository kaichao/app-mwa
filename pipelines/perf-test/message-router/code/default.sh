#!/bin/bash

PB=1
PE=480

for ((p = PB; p <= PE; p += 1)); do
    sema=$(printf "fits-24ch-ready:1257010784/p%05d/t1257010786_1257010815" "$p")
    echo "$sema"
    scalebox semaphore create $sema 24
done

for ch in {109..132}; do
    for ((pb = PB; pb <= PE; pb += 24)); do
        pe=$((pb + 23))
        m=$(printf "1257010784/1257010786_1257010815/%03d/%05d_%05d\n" "$ch" "$pb" "$pe")
        echo "beam-maker,$m" >> ${WORK_DIR}/messages.txt
    done
done
