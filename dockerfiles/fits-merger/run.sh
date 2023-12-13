#!/bin/bash

if [ $LOCAL_INPUT_ROOT ]; then
    DIR_1CH="/local${LOCAL_INPUT_ROOT}/mwa/1ch"
else
    DIR_1CH=/data/mwa/1ch
fi
if [ $LOCAL_OUTPUT_ROOT ]; then
    DIR_24CH="/local${LOCAL_OUTPUT_ROOT}/mwa/24ch"
else
    DIR_24CH=/data/mwa/24ch
fi
echo DIR_1CH:$DIR_1CH
echo DIR_24CH:$DIR_24CH


# 应该是 ${单通道目录根}/${观测号}/${起始时间戳}_${结尾时间戳}/指向号
# m="1257010784/1257010986_1257011185/00001"
cd ${DIR_1CH}/$1
input_files=$(ls *.fits)
echo input_files:${input_files}
splice_psrfits ${input_files} /work/all
code=$?
[[ $code -ne 0 ]] && echo exit after splice_psrfits, error_code:$code && exit $code

output_file=${DIR_24CH}/$1.fits
mkdir -p $(dirname ${output_file}) && mv /work/all*.fits ${output_file}
code=$?

if [ $code -eq 0 ]; then
    # for f in $input_files ; do
    #     echo ${DIR_1CH}/$1/$f >> /work/input-files.txt
    # done
    echo $output_file > /work/output-files.txt
fi

if [ "$KEEP_SOURCE_FILE" = "no" ]; then
    # PUSH
    full_path="${DIR_1CH}/$1"
    echo [DEBUG]full_path:$full_path
    rm -rf $full_path_file
fi


exit $code
