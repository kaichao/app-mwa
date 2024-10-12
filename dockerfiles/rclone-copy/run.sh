#!/bin/bash
#输入文件去空
echo "111"$1
result=$1
echo $result
resultname=${TARGET_URL}
IFS='/' read -ra dirs <<< "$resultname"

path=""
remote="common:fast-obs"
for dir in "${dirs[@]}"; do
    path="$path/$dir"
    if ! rclone ls "$remote$path"  >/dev/null 2>&1; then
        echo "Directory does not exist: $remote$path"
        echo "Creating directory: $remote$path"
       
        if ! rclone mkdir "$remote$path"; then
            echo "Failed to create directory: $remote$path"
            exit 1
        fi
        
        echo "Directory created: $remote$path"
    else
        echo "Directory exists: $remote$path"
    fi
done



SOURCE_file="/local"${SOURCE_URL}/$result
source_file_size=$(stat -c%s $SOURCE_file)
echo "SOURCE_file:"$SOURCE_file
echo "TARGET_URL:"$TARGET_URL
echo $source_file_size
echo $ACTION
basename=$(basename $result)
TARGET_file=$resultname/$basename
echo $TARGET_file
if [ -n "$SOURCE_URL" ] && [ -n "$TARGET_URL" ]; then

    if [ "$ACTION" == "PUSH_RECHECK" ]; then
       echo "开始"
       rclone copy $SOURCE_file common:fast-obs/$resultname
       exit_code=$?
       out_file_size=$(rclone size common:fast-obs/$TARGET_file --json)
       echo $out_file_size
       out_file_bytes=$(echo $out_file_size | awk -F'[:,}]' '{for(i=1;i<=NF;i++){if($i~/bytes/){print $(i+1)}}}')
       echo $out_file_bytes

       if [ $source_file_size -eq $out_file_bytes ]; then
            echo "文件大小相等"
            if [ "$RM_FILE" = "yesd" ]; then
               echo "SOURCE_file1111:"$SOURCE_file
               rm -f $SOURCE_file
               exit_code=$?
               exit ${exit_code}
            fi
            exit 0
        else
            echo "文件大小不相等"
            exit 1
        fi

    else
        echo "REMOTE_ACTION 不等于 PUSH_RECHECK"
    fi

else
    echo "WARNING: SOURCE_URL||TARGET_URL is null" >&2
    exit 2
fi

