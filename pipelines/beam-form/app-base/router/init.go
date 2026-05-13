package main

import (
	"os"
	"strconv"
)

var appID int

func init() {
	appID, _ = strconv.Atoi(os.Getenv("APP_ID"))
}
