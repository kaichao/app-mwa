echo "get message: $1"

pattern='"source_url":"([^"]+)"'
if [[ $2 =~ $pattern ]]; then
    source_url="${BASH_REMATCH[1]}"
    echo "source_url: $source_url"
else
    # no from_job in json 
    source_url=""
fi
sleep 1
scalebox task add -h source_url=$source_url $1