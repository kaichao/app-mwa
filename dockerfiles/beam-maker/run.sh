#!/bin/bash

# OBSID_BEG_END_ch_PTHEAD_PTTAIL
m=$1
# m="1257010784/1257010986_1257011185/132/001_003"
my_arr=($(echo $m | tr "_" "\n" | tr "/" "\n"))
OBSID=${my_arr[0]}
BEG=${my_arr[1]}
END=${my_arr[2]}
ch=${my_arr[3]}
PTHEAD=${my_arr[4]}
PTTAIL=${my_arr[5]}
let ii=$((10#${ch}))-108
printf -v i "%02d" $ii

echo OBSID=$OBSID
echo BEG=$BEG
echo END=$END
echo ch=$ch
echo PTHEAD=$PTHEAD
echo PTTAIL=$PTTAIL
echo i=$i


# OBSID=1257010784
# BEG=1257010986
# END=1257011185
# ch=108
# PTHEAD=001
# PTTAIL=003


UTT=2019-11-05T17:43:25.00
PTLIST=${BASEDIR}/1257010784_grid_positions_f0.85_d0.3098_l102.txt
POINTS=$(awk "NR>=${PTHEAD} && NR<=${PTTAIL} {printf \"%s\", \$0; if (NR!=${PTTAIL}) printf \",\"}" ${PTLIST})
mkdir -p ${OUTDIR} && cd ${OUTDIR}

make_beam -o ${OBSID} -b ${BEG} -e ${END} \
        -P ${POINTS} \
        -z ${UTT} \
        -d ${DATDIR} -f ${ch} \
        -m ${DATDIR}/*metafits_ppds.fits \
        -F ${CALDIR}/flagged_tiles.txt \
        -J ${CALDIR}/DI_JonesMatrices_node0${i}.dat \
        -B ${CALDIR}/BandpassCalibration_node0${i}.dat \
        -t 6000 -W 10000 -s 

code=$?

echo $1 > /work/messages.txt

exit $code
