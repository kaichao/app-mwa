# Dedisp-search module

## Run the test

### 1. Set test arguments

- LINEMODE: deal with 1 line if 1, else NCALLS each time.
- COMPRESSED_INPUT: set to yes if the fits needs decompress. in this case, the first
execution will uncompress the files, and other slots will wait.
- PLAN_FILE: file name of DDplan.

### 2. Create app

```sh
app_id=$( scalebox app create | cut -d':' -f2 | tr -d '}' )
```

### 3. Add messages

```sh
APP_ID=${app_id} scalebox task add --sink-job dedisp-search --task-file messages.txt
```
