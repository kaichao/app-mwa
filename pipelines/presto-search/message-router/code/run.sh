#!/bin/bash

code_dir=`dirname $0`

echo "num-args:"$#
echo "args:"$*

# headers='{"from_ip":"10.11.16.80","from_job":"dir-list","from_host":"n1.dcu"}'
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

pattern='"from_ip":"([^"]+)"'
if [[ $headers =~ $pattern ]]; then
    from_ip="${BASH_REMATCH[1]}"
    echo "from_ip: $from_ip"
else
    # no from_job in json 
    from_ip=""
fi

case $from_job in
    # "dir-list-1")
    #     ${code_dir}/from-dir-list-1.sh "$1"
    #     ;;
    "dir-list")
        ${code_dir}/from-dir-list.sh "$1"
        ;;
    "local-copy-unpack")
        ${code_dir}/from-local-copy-unpack.sh "$1" "$from_ip"
        ;;
    "local-copy")
        ${code_dir}/from-local-copy.sh "$1"
        ;;
    "rfi-find")
        ${code_dir}/from-rfi-find.sh "$1" "$from_ip"
        ;;
    # "dedisp")
    #     ${code_dir}/from-dedisp.sh "$1" "$from_ip"
    #     ;;
    "dedisp-search")
        ${code_dir}/from-dedisp-search.sh "$1" "$from_ip"
        ;;
    "fold")
        # echo "dm $1 completed."
        ${code_dir}/from-fold.sh "$1" "$from_ip"
        exit 0
        ;;
    "result-push")
        exit 0
        ;;
    *)  # default
        ${code_dir}/default.sh "$1"
        ;;
esac

exit $?
