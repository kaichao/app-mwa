#!/bin/bash

# t1257010784/p00001
m=$1
if [ $LOCAL_FITS_ROOT ]; then
    DIR_FITS="/local${LOCAL_FITS_ROOT}/mwa/24ch"
else
    DIR_FITS=/data/mwa/24ch
fi

if [ $LOCAL_DEDISP_ROOT ]; then
    DIR_DEDISP="/local${LOCAL_DEDISP_ROOT}/mwa/dedisp"
else
    DIR_DEDISP=/data/mwa/dedisp
fi

echo "DIR_FITS:${DIR_FITS}"
echo "message: $m"

rm -r $DIR_FITS/$m
rm -f $DIR_DEDISP/$m/RFIfile*
exit $?
