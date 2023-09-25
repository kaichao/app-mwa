#!/bin/bash

# 运行中，出现了core文件，出错原因是什么？

# 应该是 ${单通道目录根}/${观测号}/${起始时间戳}_${结尾时间戳}/指向号
# m="1257010784/1257010986_1257011185/00001"
# 指向序号
m=$1
my_arr=($(echo $m | tr "_" "\n" | tr "/" "\n"))
n=${my_arr[3]}

PROJECT=G0057
OBSID=1257010784

LO=109
HI=132
PTLIST=${BASEDIR}/1257010784_grid_positions_f0.85_d0.3098_l102.txt
NSETS=0001

POINTING=$(awk "NR==${n}" ${PTLIST})
cd ${OUTDIR}/${POINTING}
rm -f splice_*
rm -f 1257010784_ch109-132_0001.fits
sname=splice_${NSETS}
for i in $(seq -f "%03g" $LO 1 $HI);do
  echo "${PROJECT}_${OBSID}_${POINTING}_ch${i}_${NSETS}.fits" >> ${sname}
done

splice_psrfits $(cat "$sname") "${OBSID}_tmp" 
code=$?
mv "${OBSID}_tmp_0001.fits" "${OBSID}_${POINTING}_ch${LO}-${HI}_${NSETS}.fits"
exit $code
