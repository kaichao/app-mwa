#!/bin/bash

code_dir=`dirname $0`

echo "num-args:"$#
echo "args:"$*

# headers='{"from_ip":"10.11.16.76","from_job":"fits-merger","from_host":"n2.dcu"}'
headers=$2
echo headers:$headers

pattern='"from_job":"([^"]+)"'
if [[ $headers =~ $pattern ]]; then
    from_job="${BASH_REMATCH[1]}"
    echo "from_job: $from_job"
else
    # no from_job in json 
    from_job=""
fi

case $from_job in
    "beam-maker")
        ${code_dir}/from-beam-maker.sh "$1"
        ;;
    "fits-merger")
        ${code_dir}/from-fits-merger.sh "$1"
        ;;
    *)  # default
        ${code_dir}/default.sh "$1"
        ;;
esac

exit $?
